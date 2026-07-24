package subscription

import (
	"context"
	"errors"
	"go-api/domain/repository"
	"log"
	"time"
)

// SubscriptionUpdatedUseCase handles customer.subscription.updated:
// plan change (upgrade/downgrade), renewal, past_due, scheduled cancellation.
type SubscriptionUpdatedUseCase struct {
	planRepo         *repository.PlanRepository
	subscriptionRepo *repository.SubscriptionRepository
	notifier         *Notifier
}

func NewSubscriptionUpdatedUseCase(
	planRepo *repository.PlanRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	notifier *Notifier,
) *SubscriptionUpdatedUseCase {
	return &SubscriptionUpdatedUseCase{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		notifier:         notifier,
	}
}

type SubscriptionUpdatedInput struct {
	StripeSubscriptionID string
	StripePriceID        string
	Status               string
	CancelAtPeriodEnd    bool
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time
}

func (u *SubscriptionUpdatedUseCase) Execute(ctx context.Context, input SubscriptionUpdatedInput) error {
	if input.StripeSubscriptionID == "" {
		return errors.New("stripe subscription id is required")
	}

	subscription, err := (*u.subscriptionRepo).GetByStripeSubscriptionID(ctx, input.StripeSubscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}
	if subscription == nil {
		log.Printf("subscription updated: no local subscription for stripe sub %s, skipping", input.StripeSubscriptionID)
		return nil
	}

	if input.StripePriceID != "" {
		plan, err := (*u.planRepo).GetByStripePriceID(ctx, input.StripePriceID)
		if err != nil {
			return errors.New("failed to get plan")
		}
		if plan != nil {
			subscription.PlanID = plan.ID
		}
	}

	subscription.SubscriptionStatus = MapStripeStatus(input.Status)
	if !input.CurrentPeriodStart.IsZero() {
		subscription.SubscriptionStartDate = input.CurrentPeriodStart
	}
	if !input.CurrentPeriodEnd.IsZero() {
		subscription.SubscriptionEndDate = input.CurrentPeriodEnd
	}

	if err := (*u.subscriptionRepo).Update(ctx, subscription); err != nil {
		return errors.New("failed to update subscription")
	}

	u.notifier.Notify(ctx, subscription.ID)

	log.Printf("subscription updated: stripe sub %s -> status %s (cancel_at_period_end=%t)", input.StripeSubscriptionID, subscription.SubscriptionStatus, input.CancelAtPeriodEnd)
	return nil
}
