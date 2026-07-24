package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"go-api/infrastructure/stripe"
	"log"

	"github.com/google/uuid"
)

// CheckoutCompletedUseCase handles the checkout.session.completed event:
// first successful payment (free -> paid). It links the Stripe customer and
// subscription to the user and activates the chosen plan.
type CheckoutCompletedUseCase struct {
	userRepo            *repository.UserRepository
	planRepo            *repository.PlanRepository
	subscriptionRepo    *repository.SubscriptionRepository
	subscriptionGateway *stripe.SubscriptionGateway
	notifier            *Notifier
}

func NewCheckoutCompletedUseCase(
	userRepo *repository.UserRepository,
	planRepo *repository.PlanRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	subscriptionGateway *stripe.SubscriptionGateway,
	notifier *Notifier,
) *CheckoutCompletedUseCase {
	return &CheckoutCompletedUseCase{
		userRepo:            userRepo,
		planRepo:            planRepo,
		subscriptionRepo:    subscriptionRepo,
		subscriptionGateway: subscriptionGateway,
		notifier:            notifier,
	}
}

type CheckoutCompletedInput struct {
	UserID               uuid.UUID
	StripeCustomerID     string
	StripeSubscriptionID string
}

func (u *CheckoutCompletedUseCase) Execute(ctx context.Context, input CheckoutCompletedInput) error {
	if input.StripeSubscriptionID == "" {
		return errors.New("stripe subscription id is required")
	}

	user, err := (*u.userRepo).GetByID(ctx, input.UserID)
	if err != nil {
		return errors.New("failed to get user")
	}
	if user == nil {
		return errors.New("user not found")
	}

	subData, err := u.subscriptionGateway.Retrieve(ctx, input.StripeSubscriptionID)
	if err != nil {
		return err
	}

	plan, err := (*u.planRepo).GetByStripePriceID(ctx, subData.PriceID)
	if err != nil {
		return errors.New("failed to get plan")
	}
	if plan == nil {
		return errors.New("plan not found for stripe price id")
	}

	customerID := input.StripeCustomerID
	if customerID == "" {
		customerID = subData.CustomerID
	}

	status := MapStripeStatus(subData.Status)

	var subscription *entity.Subscription
	if user.SubscriptionID != nil {
		subscription, err = (*u.subscriptionRepo).GetByID(ctx, *user.SubscriptionID)
		if err != nil {
			return errors.New("failed to get subscription")
		}
	}

	if subscription == nil {
		subscription = &entity.Subscription{}
	}

	subscription.PlanID = plan.ID
	subscription.StripeCustomerID = customerID
	subscription.StripeSubscriptionID = subData.ID
	subscription.SubscriptionStatus = status
	if !subData.CurrentPeriodStart.IsZero() {
		subscription.SubscriptionStartDate = subData.CurrentPeriodStart
	}
	if !subData.CurrentPeriodEnd.IsZero() {
		subscription.SubscriptionEndDate = subData.CurrentPeriodEnd
	}

	if subscription.ID == uuid.Nil {
		if err := (*u.subscriptionRepo).Create(ctx, subscription); err != nil {
			return errors.New("failed to create subscription")
		}
	} else {
		if err := (*u.subscriptionRepo).Update(ctx, subscription); err != nil {
			return errors.New("failed to update subscription")
		}
	}

	if user.SubscriptionID == nil || *user.SubscriptionID != subscription.ID {
		user.SubscriptionID = &subscription.ID
		if err := (*u.userRepo).Update(ctx, user); err != nil {
			return errors.New("failed to link subscription to user")
		}
	}

	// payment_succeeded is notified from invoice.payment_succeeded only (fires on
	// both the first payment and renewals), so we avoid a duplicate here.
	u.notifier.Notify(ctx, subscription.ID)

	log.Printf("checkout completed: user %s subscribed to plan %s (stripe sub %s)", user.ID, plan.Slug, subData.ID)
	return nil
}
