package subscription

import (
	"context"
	"errors"
	"log"
	"time"

	"go-api/domain/entity"
	"go-api/domain/repository"
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
	log.Printf(
		"invoice payment succeeded: start stripeSubscriptionID=%s stripeCustomerID=%s billingReason=%s",
		input.StripeSubscriptionID,
		input.StripeCustomerID,
		input.BillingReason,
	)

	if input.StripeSubscriptionID == "" {
		log.Printf("invoice payment succeeded: skip, missing stripeSubscriptionID")
		return nil
	}

	subscription, err := u.subscriptionRepo.GetByStripeSubscriptionID(ctx, input.StripeSubscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}

	if subscription == nil && input.StripeCustomerID != "" {
		log.Printf("invoice payment succeeded: fallback lookup by stripeCustomerID=%s", input.StripeCustomerID)
		subscription, err = u.subscriptionRepo.GetByStripeCustomerID(ctx, input.StripeCustomerID)
		if err != nil {
			return errors.New("failed to get subscription by customer")
		}
	}

	if subscription == nil {
		log.Printf(
			"invoice payment succeeded: skip, no local subscription for stripeSubscriptionID=%s stripeCustomerID=%s",
			input.StripeSubscriptionID,
			input.StripeCustomerID,
		)
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
		log.Printf("invoice payment succeeded: update failed subscriptionID=%s: %v", subscription.ID, err)
		return errors.New("failed to update subscription")
	}

	log.Printf("invoice payment succeeded: updated subscriptionID=%s status=active", subscription.ID)

	u.notifier.Notify(ctx, subscription.ID)
	if input.BillingReason != "subscription_create" {
		u.notifier.NotifyPaymentSucceededAfter(subscription.ID)
	}

	return nil
}
