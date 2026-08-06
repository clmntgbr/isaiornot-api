package port

import (
	"context"
	"time"

	"go-api/domain/entity"
)

type SubscriptionData struct {
	ID                 string
	ItemID             string
	CustomerID         string
	PriceID            string
	Status             string
	CancelAtPeriodEnd  bool
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
}

type ProrationPreviewLine struct {
	Description string
	Amount      int64
	Proration   bool
}

type ProrationPreview struct {
	Currency      string
	AmountDue     int64
	Subtotal      int64
	Total         int64
	ProrationDate int64
	PeriodStart   time.Time
	PeriodEnd     time.Time
	Lines         []ProrationPreviewLine
}

type CheckoutSessionGateway interface {
	Create(ctx context.Context, plan *entity.Plan, user *entity.User, email string, stripeCustomerID string) (string, error)
}

type SubscriptionGateway interface {
	Retrieve(ctx context.Context, subscriptionID string) (*SubscriptionData, error)
	UpdatePrice(ctx context.Context, subscriptionID string, itemID string, priceID string, prorationDate *int64) (*SubscriptionData, error)
	PreviewPriceChange(ctx context.Context, subscriptionID string, itemID string, priceID string) (*ProrationPreview, error)
}

type BillingPortalGateway interface {
	Create(ctx context.Context, stripeCustomerID string) (string, error)
}
