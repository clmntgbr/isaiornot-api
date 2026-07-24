package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"go-api/infrastructure/stripe"
	"go-api/usecase/clerk"

	"github.com/google/uuid"
)

var (
	ErrPlanNotFound       = errors.New("plan not found")
	ErrPlanInactive       = errors.New("plan is inactive")
	ErrFreePlanCheckout   = errors.New("free plan does not require checkout")
	ErrMissingStripePrice = errors.New("plan has no stripe price")
)

type CreateSubscriptionUseCase struct {
	planRepo               *repository.PlanRepository
	fetchUserUseCase       *clerk.FetchUserUseCase
	checkoutSessionGateway *stripe.CheckoutSessionGateway
}

func NewCreateSubscriptionUseCase(
	planRepo *repository.PlanRepository,
	fetchUserUseCase *clerk.FetchUserUseCase,
	checkoutSessionGateway *stripe.CheckoutSessionGateway,
) *CreateSubscriptionUseCase {
	return &CreateSubscriptionUseCase{
		planRepo:               planRepo,
		fetchUserUseCase:       fetchUserUseCase,
		checkoutSessionGateway: checkoutSessionGateway,
	}
}

func (u *CreateSubscriptionUseCase) Execute(
	ctx context.Context,
	user *entity.User,
	planID uuid.UUID,
) (string, error) {
	if user == nil {
		return "", errors.New("user is required")
	}

	plan, err := (*u.planRepo).GetByID(ctx, planID)
	if err != nil {
		return "", errors.New("failed to get plan")
	}
	if plan == nil {
		return "", ErrPlanNotFound
	}
	if !plan.IsActive {
		return "", ErrPlanInactive
	}
	if plan.Slug == entity.FreePlanSlug {
		return "", ErrFreePlanCheckout
	}

	clerkUser, err := u.fetchUserUseCase.Execute(ctx, user.ClerkID)
	if err != nil {
		return "", errors.New("failed to get user email")
	}

	url, err := u.checkoutSessionGateway.Create(ctx, plan, user, clerkUser.Email)
	if err != nil {
		return "", err
	}

	return url, nil
}
