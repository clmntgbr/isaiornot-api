package subscription

import (
	"context"
	"errors"
	"time"

	"go-api/domain/entity"
	"go-api/domain/port"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type PreviewPlanChangeUseCase struct {
	planRepo            repository.PlanRepository
	subscriptionRepo    repository.SubscriptionRepository
	subscriptionGateway port.SubscriptionGateway
}

func NewPreviewPlanChangeUseCase(
	planRepo repository.PlanRepository,
	subscriptionRepo repository.SubscriptionRepository,
	subscriptionGateway port.SubscriptionGateway,
) *PreviewPlanChangeUseCase {
	return &PreviewPlanChangeUseCase{
		planRepo:            planRepo,
		subscriptionRepo:    subscriptionRepo,
		subscriptionGateway: subscriptionGateway,
	}
}

type PlanChangePreview struct {
	RequiresCheckout bool
	Currency         string
	AmountDue        int64
	Subtotal         int64
	Total            int64
	ProrationDate    int64
	PeriodStart      time.Time
	PeriodEnd        time.Time
	Lines            []port.ProrationPreviewLine
	CurrentPlanID    string
	CurrentPlanSlug  string
	TargetPlanID     string
	TargetPlanSlug   string
	TargetPlanName   string
	TargetPlanPrice  float64
}

func (u *PreviewPlanChangeUseCase) Execute(
	ctx context.Context,
	user *entity.User,
	planID uuid.UUID,
) (*PlanChangePreview, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}

	plan, err := u.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, errors.New("failed to get plan")
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}
	if !plan.IsActive {
		return nil, ErrPlanInactive
	}
	if plan.Slug == entity.FreePlanSlug {
		return nil, ErrFreePlanCheckout
	}
	if plan.StripePriceID == "" {
		return nil, ErrMissingStripePrice
	}

	var current *entity.Subscription
	if user.SubscriptionID != nil {
		current, err = u.subscriptionRepo.GetByID(ctx, *user.SubscriptionID)
		if err != nil {
			return nil, errors.New("failed to get subscription")
		}
	}

	preview := &PlanChangePreview{
		TargetPlanID:    plan.ID.String(),
		TargetPlanSlug:  plan.Slug,
		TargetPlanName:  plan.Name,
		TargetPlanPrice: plan.Price,
		Currency:        string(plan.Currency),
	}

	if current != nil {
		preview.CurrentPlanID = current.PlanID.String()
		preview.CurrentPlanSlug = current.Plan.Slug
	}

	if current != nil && current.PlanID == plan.ID && current.StripeSubscriptionID != "" {
		return nil, ErrAlreadyOnPlan
	}

	if !canUpdateStripeSubscription(current) {
		preview.RequiresCheckout = true
		return preview, nil
	}

	stripeSub, err := u.subscriptionGateway.Retrieve(ctx, current.StripeSubscriptionID)
	if err != nil {
		return nil, err
	}
	if stripeSub.ItemID == "" {
		return nil, ErrMissingStripeSub
	}
	if stripeSub.PriceID == plan.StripePriceID {
		return nil, ErrAlreadyOnPlan
	}

	stripePreview, err := u.subscriptionGateway.PreviewPriceChange(
		ctx,
		current.StripeSubscriptionID,
		stripeSub.ItemID,
		plan.StripePriceID,
	)
	if err != nil {
		return nil, err
	}

	preview.RequiresCheckout = false
	preview.Currency = stripePreview.Currency
	preview.AmountDue = stripePreview.AmountDue
	preview.Subtotal = stripePreview.Subtotal
	preview.Total = stripePreview.Total
	preview.ProrationDate = stripePreview.ProrationDate
	preview.PeriodStart = stripePreview.PeriodStart
	preview.PeriodEnd = stripePreview.PeriodEnd
	preview.Lines = stripePreview.Lines

	return preview, nil
}
