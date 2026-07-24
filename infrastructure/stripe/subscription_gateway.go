package stripe

import (
	"context"
	"fmt"
	"go-api/infrastructure/config"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/subscription"
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

type SubscriptionGateway struct {
	secretKey string
}

func NewSubscriptionGateway(cfg *config.Config) *SubscriptionGateway {
	return &SubscriptionGateway{
		secretKey: cfg.StripeSecretKey,
	}
}

func (g *SubscriptionGateway) Retrieve(ctx context.Context, subscriptionID string) (*SubscriptionData, error) {
	if g.secretKey == "" {
		return nil, fmt.Errorf("stripe secret key is not configured")
	}

	stripe.Key = g.secretKey

	params := &stripe.SubscriptionParams{}
	params.Context = ctx

	sub, err := subscription.Get(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve subscription: %w", err)
	}

	return ExtractSubscriptionData(sub), nil
}

func ExtractSubscriptionData(sub *stripe.Subscription) *SubscriptionData {
	data := &SubscriptionData{
		ID:                sub.ID,
		Status:            string(sub.Status),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
	}

	if sub.Customer != nil {
		data.CustomerID = sub.Customer.ID
	}

	if sub.Items != nil && len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		if item.Price != nil {
			data.PriceID = item.Price.ID
		}
		if item.CurrentPeriodStart > 0 {
			data.CurrentPeriodStart = time.Unix(item.CurrentPeriodStart, 0).UTC()
		}
		if item.CurrentPeriodEnd > 0 {
			data.CurrentPeriodEnd = time.Unix(item.CurrentPeriodEnd, 0).UTC()
		}
	}

	return data
}
