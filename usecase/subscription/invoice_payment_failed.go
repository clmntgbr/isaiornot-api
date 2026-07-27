package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
)

type InvoicePaymentFailedUseCase struct {
	subscriptionRepo repository.SubscriptionRepository
	notifier         *Notifier
}

func NewInvoicePaymentFailedUseCase(
	subscriptionRepo repository.SubscriptionRepository,
	notifier *Notifier,
) *InvoicePaymentFailedUseCase {
	return &InvoicePaymentFailedUseCase{
		subscriptionRepo: subscriptionRepo,
		notifier:         notifier,
	}
}

type InvoicePaymentFailedInput struct {
	StripeSubscriptionID string
	StripeCustomerID     string
}

func (u *InvoicePaymentFailedUseCase) Execute(ctx context.Context, input InvoicePaymentFailedInput) error {
	if input.StripeSubscriptionID == "" {
		return nil
	}

	subscription, err := u.subscriptionRepo.GetByStripeSubscriptionID(ctx, input.StripeSubscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}

	if subscription == nil && input.StripeCustomerID != "" {
		subscription, err = u.subscriptionRepo.GetByStripeCustomerID(ctx, input.StripeCustomerID)
		if err != nil {
			return errors.New("failed to get subscription by customer")
		}
	}

	if subscription == nil {
		return nil
	}

	subscription.StripeSubscriptionID = input.StripeSubscriptionID
	if input.StripeCustomerID != "" {
		subscription.StripeCustomerID = input.StripeCustomerID
	}
	subscription.SubscriptionStatus = entity.SubscriptionStatusPastDue
	if err := u.subscriptionRepo.Update(ctx, subscription); err != nil {
		return errors.New("failed to update subscription")
	}

	u.notifier.Notify(ctx, subscription.ID)
	u.notifier.NotifyPaymentFailedAfter(subscription.ID)

	return nil
}
