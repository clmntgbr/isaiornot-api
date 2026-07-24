package stripe

import (
	"context"
	"fmt"
	"go-api/domain/entity"
	"go-api/infrastructure/config"
	"strconv"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

type CheckoutSessionGateway struct {
	redirectSuccessURL string
	redirectCancelURL  string
	secretKey          string
}

func NewCheckoutSessionGateway(cfg *config.Config) *CheckoutSessionGateway {
	return &CheckoutSessionGateway{
		redirectSuccessURL: cfg.RedirectSuccessURL,
		redirectCancelURL:  cfg.RedirectCancelURL,
		secretKey:          cfg.StripeSecretKey,
	}
}

func (g *CheckoutSessionGateway) Create(
	ctx context.Context,
	plan *entity.Plan,
	user *entity.User,
	email string,
) (string, error) {
	if g.secretKey == "" {
		return "", fmt.Errorf("stripe secret key is not configured")
	}
	if plan.StripePriceID == "" {
		return "", fmt.Errorf("plan has no stripe price id")
	}

	stripe.Key = g.secretKey

	params := &stripe.CheckoutSessionParams{
		ClientReferenceID: stripe.String(user.ID.String()),
		SuccessURL:        stripe.String(g.redirectSuccessURL),
		CancelURL:         stripe.String(g.redirectCancelURL),
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(plan.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"plan_id":       plan.ID.String(),
			"plan_name":     plan.Name,
			"plan_price":    strconv.FormatFloat(plan.Price, 'f', -1, 64),
			"currency":      string(plan.Currency),
			"plan_interval": string(plan.BillingInterval),
			"user_id":       user.ID.String(),
			"user_email":    email,
			"user_name":     user.FirstName + " " + user.LastName,
		},
	}
	params.Context = ctx

	if email != "" {
		params.CustomerEmail = stripe.String(email)
	}

	s, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}
	if s.URL == "" {
		return "", fmt.Errorf("checkout session URL is required")
	}

	return s.URL, nil
}
