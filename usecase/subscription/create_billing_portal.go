package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"go-api/infrastructure/stripe"
)

var ErrMissingStripeCustomer = errors.New("user has no stripe customer id")

type CreateBillingPortalUseCase struct {
	subscriptionRepo     *repository.SubscriptionRepository
	billingPortalGateway *stripe.BillingPortalGateway
}

func NewCreateBillingPortalUseCase(
	subscriptionRepo *repository.SubscriptionRepository,
	billingPortalGateway *stripe.BillingPortalGateway,
) *CreateBillingPortalUseCase {
	return &CreateBillingPortalUseCase{
		subscriptionRepo:     subscriptionRepo,
		billingPortalGateway: billingPortalGateway,
	}
}

func (u *CreateBillingPortalUseCase) Execute(ctx context.Context, user *entity.User) (string, error) {
	if user == nil {
		return "", errors.New("user is required")
	}
	if user.SubscriptionID == nil {
		return "", ErrSubscriptionNotFound
	}

	subscription, err := (*u.subscriptionRepo).GetByID(ctx, *user.SubscriptionID)
	if err != nil {
		return "", errors.New("failed to get subscription")
	}
	if subscription == nil {
		return "", ErrSubscriptionNotFound
	}
	if subscription.StripeCustomerID == "" {
		return "", ErrMissingStripeCustomer
	}

	url, err := u.billingPortalGateway.Create(ctx, subscription.StripeCustomerID)
	if err != nil {
		return "", err
	}

	return url, nil
}
