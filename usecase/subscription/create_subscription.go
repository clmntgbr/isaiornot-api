package subscription

import (
	"context"
	"errors"
	"log"

	"go-api/domain/entity"
	"go-api/domain/port"
	"go-api/domain/repository"
	"go-api/usecase/identity"

	"github.com/google/uuid"
)

var (
	ErrPlanNotFound       = errors.New("plan not found")
	ErrPlanInactive       = errors.New("plan is inactive")
	ErrFreePlanCheckout   = errors.New("free plan does not require checkout")
	ErrMissingStripePrice = errors.New("plan has no stripe price")
	ErrAlreadyOnPlan      = errors.New("user is already on this plan")
	ErrMissingStripeSub   = errors.New("subscription has no stripe subscription id")
)

type CreateSubscriptionUseCase struct {
	planRepo               repository.PlanRepository
	subscriptionRepo       repository.SubscriptionRepository
	fetchUserUseCase       *identity.FetchUserUseCase
	checkoutSessionGateway port.CheckoutSessionGateway
	subscriptionGateway    port.SubscriptionGateway
	notifier               *Notifier
}

func NewCreateSubscriptionUseCase(
	planRepo repository.PlanRepository,
	subscriptionRepo repository.SubscriptionRepository,
	fetchUserUseCase *identity.FetchUserUseCase,
	checkoutSessionGateway port.CheckoutSessionGateway,
	subscriptionGateway port.SubscriptionGateway,
	notifier *Notifier,
) *CreateSubscriptionUseCase {
	return &CreateSubscriptionUseCase{
		planRepo:               planRepo,
		subscriptionRepo:       subscriptionRepo,
		fetchUserUseCase:       fetchUserUseCase,
		checkoutSessionGateway: checkoutSessionGateway,
		subscriptionGateway:    subscriptionGateway,
		notifier:               notifier,
	}
}

type ChangeSubscriptionResult struct {
	URL     string
	Updated bool
}

type ChangeSubscriptionInput struct {
	PlanID        uuid.UUID
	ProrationDate *int64
}

func (u *CreateSubscriptionUseCase) Execute(
	ctx context.Context,
	user *entity.User,
	input ChangeSubscriptionInput,
) (*ChangeSubscriptionResult, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}

	plan, err := u.planRepo.GetByID(ctx, input.PlanID)
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

	if current != nil && current.PlanID == plan.ID && current.StripeSubscriptionID != "" {
		return nil, ErrAlreadyOnPlan
	}

	if canUpdateStripeSubscription(current) {
		if err := u.updateExistingSubscription(ctx, current, plan, input.ProrationDate); err != nil {
			return nil, err
		}
		return &ChangeSubscriptionResult{Updated: true}, nil
	}

	clerkUser, err := u.fetchUserUseCase.Execute(ctx, user.ClerkID)
	if err != nil {
		return nil, errors.New("failed to get user email")
	}

	stripeCustomerID := ""
	if current != nil {
		stripeCustomerID = current.StripeCustomerID
	}

	url, err := u.checkoutSessionGateway.Create(ctx, plan, user, clerkUser.Email, stripeCustomerID)
	if err != nil {
		return nil, err
	}

	return &ChangeSubscriptionResult{URL: url}, nil
}

func canUpdateStripeSubscription(subscription *entity.Subscription) bool {
	if subscription == nil || subscription.StripeSubscriptionID == "" {
		return false
	}

	switch subscription.SubscriptionStatus {
	case entity.SubscriptionStatusActive, entity.SubscriptionStatusPastDue, entity.SubscriptionStatusPending:
		return true
	default:
		return false
	}
}

func (u *CreateSubscriptionUseCase) updateExistingSubscription(
	ctx context.Context,
	subscription *entity.Subscription,
	plan *entity.Plan,
	prorationDate *int64,
) error {
	stripeSub, err := u.subscriptionGateway.Retrieve(ctx, subscription.StripeSubscriptionID)
	if err != nil {
		return err
	}
	if stripeSub.ItemID == "" {
		return ErrMissingStripeSub
	}
	if stripeSub.PriceID == plan.StripePriceID {
		return ErrAlreadyOnPlan
	}

	updated, err := u.subscriptionGateway.UpdatePrice(
		ctx,
		subscription.StripeSubscriptionID,
		stripeSub.ItemID,
		plan.StripePriceID,
		prorationDate,
	)
	if err != nil {
		return err
	}

	subscription.PlanID = plan.ID
	subscription.Plan = *plan
	subscription.SubscriptionStatus = MapBillingStatus(updated.Status)
	subscription.CancelAtPeriodEnd = updated.CancelAtPeriodEnd
	if updated.CustomerID != "" {
		subscription.StripeCustomerID = updated.CustomerID
	}
	if !updated.CurrentPeriodStart.IsZero() {
		subscription.SubscriptionStartDate = updated.CurrentPeriodStart
	}
	if !updated.CurrentPeriodEnd.IsZero() {
		subscription.SubscriptionEndDate = updated.CurrentPeriodEnd
	}

	if err := u.subscriptionRepo.Update(ctx, subscription); err != nil {
		return errors.New("failed to update subscription")
	}

	log.Printf(
		"subscription plan changed locally id=%s plan=%s stripeSub=%s",
		subscription.ID,
		plan.Slug,
		subscription.StripeSubscriptionID,
	)

	u.notifier.Notify(ctx, subscription.ID)
	return nil
}
