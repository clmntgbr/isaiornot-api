package port

import (
	"context"
	"time"

	"go-api/domain/entity"
)

type SubscriptionData struct {
	ID                 string
	CustomerID         string
	PriceID            string
	Status             string
	CancelAtPeriodEnd  bool
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
}

type CheckoutSessionGateway interface {
	Create(ctx context.Context, plan *entity.Plan, user *entity.User, email string) (string, error)
}

type SubscriptionGateway interface {
	Retrieve(ctx context.Context, subscriptionID string) (*SubscriptionData, error)
}

type BillingPortalGateway interface {
	Create(ctx context.Context, stripeCustomerID string) (string, error)
}
