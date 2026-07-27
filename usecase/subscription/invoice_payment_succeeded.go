package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"time"
)

type InvoicePaymentSucceededUseCase struct {
	subscriptionRepo repository.SubscriptionRepository
	notifier         *Notifier
}

func NewInvoicePaymentSucceededUseCase(
	subscriptionRepo repository.SubscriptionRepository,
	notifier *Notifier,
) *InvoicePaymentSucceededUseCase {
	return &InvoicePaymentSucceededUseCase{
		subscriptionRepo: subscriptionRepo,
		notifier:         notifier,
	}
}

type InvoicePaymentSucceededInput struct {
	StripeSubscriptionID string
	StripeCustomerID     string
	BillingReason        string
	PeriodStart          time.Time
	PeriodEnd            time.Time
}

func (u *InvoicePaymentSucceededUseCase) Execute(ctx context.Context, input InvoicePaymentSucceededInput) error {
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
	subscription.SubscriptionStatus = entity.SubscriptionStatusActive
	subscription.CancelAtPeriodEnd = false
	if !input.PeriodStart.IsZero() {
		subscription.SubscriptionStartDate = input.PeriodStart
	}
	if !input.PeriodEnd.IsZero() {
		subscription.SubscriptionEndDate = input.PeriodEnd
	}

	if err := u.subscriptionRepo.Update(ctx, subscription); err != nil {
		return errors.New("failed to update subscription")
	}

	u.notifier.Notify(ctx, subscription.ID)
	if input.BillingReason != "subscription_create" {
		u.notifier.NotifyPaymentSucceededAfter(subscription.ID)
	}

	return nil
}
