package plan

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Plan struct {
	ID          uuid.UUID
	Name        string
	Description string
	Slug        string

	StripePriceID string
	IsActive      bool

	BillingInterval BillingInterval
	Price           float64
	Currency        Currency

	QuotaID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

func NewPlan(
	name string,
	description string,
	slug string,
	stripePriceID string,
	isActive bool,
	billingInterval BillingInterval,
	price float64,
	currency Currency,
	quotaID uuid.UUID,
) *Plan {
	now := time.Now().UTC()
	p := &Plan{
		ID:              uuid.New(),
		Name:            name,
		Description:     description,
		Slug:            slug,
		StripePriceID:   stripePriceID,
		IsActive:        isActive,
		BillingInterval: billingInterval,
		Price:           price,
		Currency:        currency,
		QuotaID:         quotaID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	p.recordEvent(PlanCreated{
		ID:              uuid.New().String(),
		PlanID:          p.ID.String(),
		Name:            p.Name,
		Description:     p.Description,
		Slug:            p.Slug,
		StripePriceID:   p.StripePriceID,
		IsActive:        p.IsActive,
		BillingInterval: string(p.BillingInterval),
		Price:           p.Price,
		Currency:        string(p.Currency),
		QuotaID:         p.QuotaID.String(),
		Timestamp:       now,
	})
	return p
}

func (p *Plan) PullEvents() []event.DomainEvent {
	events := p.events
	p.events = nil
	return events
}

func (p *Plan) recordEvent(e event.DomainEvent) {
	p.events = append(p.events, e)
}

func (p *Plan) ApplyUpdate(
	name string,
	description string,
	slug string,
	stripePriceID string,
	isActive bool,
	billingInterval BillingInterval,
	price float64,
	currency Currency,
	quotaID uuid.UUID,
) {
	p.Name = name
	p.Description = description
	p.Slug = slug
	p.StripePriceID = stripePriceID
	p.IsActive = isActive
	p.BillingInterval = billingInterval
	p.Price = price
	p.Currency = currency
	p.QuotaID = quotaID
	p.UpdatedAt = time.Now().UTC()
	p.recordEvent(PlanUpdated{
		ID:              uuid.New().String(),
		PlanID:          p.ID.String(),
		Name:            p.Name,
		Description:     p.Description,
		Slug:            p.Slug,
		StripePriceID:   p.StripePriceID,
		IsActive:        p.IsActive,
		BillingInterval: string(p.BillingInterval),
		Price:           p.Price,
		Currency:        string(p.Currency),
		QuotaID:         p.QuotaID.String(),
		Timestamp:       p.UpdatedAt,
	})
}

func (p *Plan) MarkDeleted() {
	p.UpdatedAt = time.Now().UTC()
	p.recordEvent(PlanDeleted{
		ID:        uuid.New().String(),
		PlanID:    p.ID.String(),
		Timestamp: p.UpdatedAt,
	})
}
