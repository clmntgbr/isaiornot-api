package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"log"
	"time"
)

// SubscriptionDeletedUseCase handles customer.subscription.deleted:
// the subscription actually ended. The user is reverted to the free plan.
type SubscriptionDeletedUseCase struct {
	planRepo         *repository.PlanRepository
	subscriptionRepo *repository.SubscriptionRepository
	notifier         *Notifier
}

func NewSubscriptionDeletedUseCase(
	planRepo *repository.PlanRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	notifier *Notifier,
) *SubscriptionDeletedUseCase {
	return &SubscriptionDeletedUseCase{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		notifier:         notifier,
	}
}

func (u *SubscriptionDeletedUseCase) Execute(ctx context.Context, stripeSubscriptionID string) error {
	if stripeSubscriptionID == "" {
		return errors.New("stripe subscription id is required")
	}

	subscription, err := (*u.subscriptionRepo).GetByStripeSubscriptionID(ctx, stripeSubscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}
	if subscription == nil {
		log.Printf("subscription deleted: no local subscription for stripe sub %s, skipping", stripeSubscriptionID)
		return nil
	}

	freePlan, err := (*u.planRepo).GetBySlug(ctx, entity.FreePlanSlug)
	if err != nil {
		return errors.New("failed to get free plan")
	}
	if freePlan == nil {
		return errors.New("free plan not found")
	}

	now := time.Now().UTC()
	subscription.PlanID = freePlan.ID
	subscription.StripeSubscriptionID = ""
	subscription.SubscriptionStatus = entity.SubscriptionStatusActive
	subscription.SubscriptionStartDate = now
	subscription.SubscriptionEndDate = now.AddDate(100, 0, 0)
	subscription.QuotaPeriodStart = now

	if err := (*u.subscriptionRepo).Update(ctx, subscription); err != nil {
		return errors.New("failed to update subscription")
	}

	u.notifier.Notify(ctx, subscription.ID)

	log.Printf("subscription deleted: stripe sub %s reverted to free plan", stripeSubscriptionID)
	return nil
}
