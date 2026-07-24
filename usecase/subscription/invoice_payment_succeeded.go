package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"log"
	"time"
)

type InvoicePaymentSucceededUseCase struct {
	subscriptionRepo *repository.SubscriptionRepository
	notifier         *Notifier
}

func NewInvoicePaymentSucceededUseCase(
	subscriptionRepo *repository.SubscriptionRepository,
	notifier *Notifier,
) *InvoicePaymentSucceededUseCase {
	return &InvoicePaymentSucceededUseCase{
		subscriptionRepo: subscriptionRepo,
		notifier:         notifier,
	}
}

type InvoicePaymentSucceededInput struct {
	StripeSubscriptionID string
	BillingReason        string
	PeriodStart          time.Time
	PeriodEnd            time.Time
}

func (u *InvoicePaymentSucceededUseCase) Execute(ctx context.Context, input InvoicePaymentSucceededInput) error {
	if input.StripeSubscriptionID == "" {
		return nil
	}

	subscription, err := (*u.subscriptionRepo).GetByStripeSubscriptionID(ctx, input.StripeSubscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}
	if subscription == nil {
		log.Printf("invoice payment succeeded: no local subscription for stripe sub %s, skipping", input.StripeSubscriptionID)
		return nil
	}

	subscription.SubscriptionStatus = entity.SubscriptionStatusActive
	if !input.PeriodStart.IsZero() {
		subscription.SubscriptionStartDate = input.PeriodStart
	}
	if !input.PeriodEnd.IsZero() {
		subscription.SubscriptionEndDate = input.PeriodEnd
	}

	if err := (*u.subscriptionRepo).Update(ctx, subscription); err != nil {
		return errors.New("failed to update subscription")
	}

	if input.BillingReason == "subscription_cycle" {
		log.Printf("invoice payment succeeded: new billing period for stripe sub %s", input.StripeSubscriptionID)
	}

	u.notifier.Notify(ctx, subscription.ID)
	u.notifier.NotifyPaymentSucceededAfter(subscription.ID)

	log.Printf("invoice payment succeeded: stripe sub %s confirmed active", input.StripeSubscriptionID)
	return nil
}
