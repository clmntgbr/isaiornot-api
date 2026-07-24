package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"log"
)

// InvoicePaymentFailedUseCase handles invoice.payment_failed:
// a charge failed (expired card, insufficient funds...). The subscription is
// moved to past_due; Stripe automatically retries according to its policy.
type InvoicePaymentFailedUseCase struct {
	subscriptionRepo *repository.SubscriptionRepository
	notifier         *Notifier
}

func NewInvoicePaymentFailedUseCase(
	subscriptionRepo *repository.SubscriptionRepository,
	notifier *Notifier,
) *InvoicePaymentFailedUseCase {
	return &InvoicePaymentFailedUseCase{
		subscriptionRepo: subscriptionRepo,
		notifier:         notifier,
	}
}

func (u *InvoicePaymentFailedUseCase) Execute(ctx context.Context, stripeSubscriptionID string) error {
	if stripeSubscriptionID == "" {
		return nil
	}

	subscription, err := (*u.subscriptionRepo).GetByStripeSubscriptionID(ctx, stripeSubscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}
	if subscription == nil {
		log.Printf("invoice payment failed: no local subscription for stripe sub %s, skipping", stripeSubscriptionID)
		return nil
	}

	subscription.SubscriptionStatus = entity.SubscriptionStatusPastDue
	if err := (*u.subscriptionRepo).Update(ctx, subscription); err != nil {
		return errors.New("failed to update subscription")
	}

	u.notifier.Notify(ctx, subscription.ID)
	u.notifier.NotifyPaymentFailedAfter(subscription.ID)

	// TODO: notify the user by email that their payment failed.
	log.Printf("invoice payment failed: stripe sub %s marked past_due", stripeSubscriptionID)
	return nil
}
