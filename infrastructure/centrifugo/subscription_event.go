package centrifugo

import (
	"encoding/json"
	"time"

	"go-api/domain/entity"

	"github.com/google/uuid"
)

const (
	EventSubscriptionUpdated = "subscription_updated"
	EventPaymentSucceeded    = "payment_succeeded"
	EventPaymentFailed       = "payment_failed"
)

type SubscriptionEvent struct {
	Type              string    `json:"type"`
	UserID            string    `json:"userId"`
	SubscriptionID    string    `json:"subscriptionId"`
	Status            string    `json:"status"`
	CancelAtPeriodEnd bool      `json:"cancelAtPeriodEnd"`
	PlanID            string    `json:"planId"`
	PlanSlug          string    `json:"planSlug"`
	PlanName          string    `json:"planName"`
	StartDate         time.Time `json:"startDate"`
	EndDate           time.Time `json:"endDate"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func NewSubscriptionEvent(subscription *entity.Subscription, userID uuid.UUID) SubscriptionEvent {
	return NewSubscriptionEventWithType(subscription, userID, EventSubscriptionUpdated)
}

func NewSubscriptionEventWithType(subscription *entity.Subscription, userID uuid.UUID, eventType string) SubscriptionEvent {
	return SubscriptionEvent{
		Type:              eventType,
		UserID:            userID.String(),
		SubscriptionID:    subscription.ID.String(),
		Status:            string(subscription.SubscriptionStatus),
		CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
		PlanID:            subscription.PlanID.String(),
		PlanSlug:          subscription.Plan.Slug,
		PlanName:          subscription.Plan.Name,
		StartDate:         subscription.SubscriptionStartDate,
		EndDate:           subscription.SubscriptionEndDate,
		UpdatedAt:         subscription.UpdatedAt,
	}
}

func (e SubscriptionEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
