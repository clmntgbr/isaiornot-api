package subscription

import (
	"context"
	"errors"
	"go-api/domain/repository"
	"time"
)

type SubscriptionUpdatedUseCase struct {
	planRepo         repository.PlanRepository
	subscriptionRepo repository.SubscriptionRepository
	notifier         *Notifier
}

func NewSubscriptionUpdatedUseCase(
	planRepo repository.PlanRepository,
	subscriptionRepo repository.SubscriptionRepository,
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
	StripeCustomerID     string
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

	if input.StripePriceID != "" {
		plan, err := u.planRepo.GetByStripePriceID(ctx, input.StripePriceID)
		if err != nil {
			return errors.New("failed to get plan")
		}
		if plan != nil {
			subscription.PlanID = plan.ID
		}
	}

	subscription.SubscriptionStatus = MapBillingStatus(input.Status)
	subscription.CancelAtPeriodEnd = input.CancelAtPeriodEnd
	if !input.CurrentPeriodStart.IsZero() {
		subscription.SubscriptionStartDate = input.CurrentPeriodStart
	}
	if !input.CurrentPeriodEnd.IsZero() {
		subscription.SubscriptionEndDate = input.CurrentPeriodEnd
	}

	if err := u.subscriptionRepo.Update(ctx, subscription); err != nil {
		return errors.New("failed to update subscription")
	}

	u.notifier.Notify(ctx, subscription.ID)

	return nil
}
