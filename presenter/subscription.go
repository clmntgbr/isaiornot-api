package presenter

import (
	"go-api/domain/entity"
	"time"
)

type CheckoutSessionResponse struct {
	URL string `json:"url"`
}

func NewCheckoutSessionResponse(url string) CheckoutSessionResponse {
	return CheckoutSessionResponse{URL: url}
}

type SubscriptionResponse struct {
	ID                   string        `json:"id"`
	Status               string        `json:"status"`
	StripeCustomerID     string        `json:"stripeCustomerId"`
	StripeSubscriptionID string        `json:"stripeSubscriptionId"`
	StartDate            time.Time     `json:"startDate"`
	EndDate              time.Time     `json:"endDate"`
	Plan                 *PlanResponse `json:"plan"`
	CreatedAt            time.Time     `json:"createdAt"`
	UpdatedAt            time.Time     `json:"updatedAt"`
}

func NewSubscriptionResponse(subscription *entity.Subscription) *SubscriptionResponse {
	return &SubscriptionResponse{
		ID:                   subscription.ID.String(),
		Status:               string(subscription.SubscriptionStatus),
		StripeCustomerID:     subscription.StripeCustomerID,
		StripeSubscriptionID: subscription.StripeSubscriptionID,
		StartDate:            subscription.SubscriptionStartDate,
		EndDate:              subscription.SubscriptionEndDate,
		Plan:                 NewPlanResponse(&subscription.Plan),
		CreatedAt:            subscription.CreatedAt,
		UpdatedAt:            subscription.UpdatedAt,
	}
}
