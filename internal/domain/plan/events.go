package plan

import "time"

const (
	EventTypePlanCreated = "plan.created.v1"
	EventTypePlanUpdated = "plan.updated.v1"
	EventTypePlanDeleted = "plan.deleted.v1"
)

type PlanCreated struct {
	ID              string    `json:"eventId"`
	PlanID          string    `json:"planId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Slug            string    `json:"slug"`
	StripePriceID   string    `json:"stripePriceId"`
	IsActive        bool      `json:"isActive"`
	BillingInterval string    `json:"billingInterval"`
	Price           float64   `json:"price"`
	Currency        string    `json:"currency"`
	QuotaID         string    `json:"quotaId"`
	Timestamp       time.Time `json:"timestamp"`
}

func (e PlanCreated) EventID() string       { return e.ID }
func (e PlanCreated) EventType() string     { return EventTypePlanCreated }
func (e PlanCreated) AggregateID() string   { return e.PlanID }
func (e PlanCreated) OccurredAt() time.Time { return e.Timestamp }

type PlanUpdated struct {
	ID              string    `json:"eventId"`
	PlanID          string    `json:"planId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Slug            string    `json:"slug"`
	StripePriceID   string    `json:"stripePriceId"`
	IsActive        bool      `json:"isActive"`
	BillingInterval string    `json:"billingInterval"`
	Price           float64   `json:"price"`
	Currency        string    `json:"currency"`
	QuotaID         string    `json:"quotaId"`
	Timestamp       time.Time `json:"timestamp"`
}

func (e PlanUpdated) EventID() string       { return e.ID }
func (e PlanUpdated) EventType() string     { return EventTypePlanUpdated }
func (e PlanUpdated) AggregateID() string   { return e.PlanID }
func (e PlanUpdated) OccurredAt() time.Time { return e.Timestamp }

type PlanDeleted struct {
	ID        string    `json:"eventId"`
	PlanID    string    `json:"planId"`
	Timestamp time.Time `json:"timestamp"`
}

func (e PlanDeleted) EventID() string       { return e.ID }
func (e PlanDeleted) EventType() string     { return EventTypePlanDeleted }
func (e PlanDeleted) AggregateID() string   { return e.PlanID }
func (e PlanDeleted) OccurredAt() time.Time { return e.Timestamp }
