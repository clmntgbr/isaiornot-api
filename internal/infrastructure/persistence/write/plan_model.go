package write

import (
	"time"

	"github.com/google/uuid"
)

type PlanModel struct {
	ID              uuid.UUID `gorm:"column:id;primaryKey"`
	Name            string    `gorm:"column:name"`
	Description     string    `gorm:"column:description"`
	Slug            string    `gorm:"column:slug"`
	StripePriceID   string    `gorm:"column:stripe_price_id"`
	IsActive        bool      `gorm:"column:is_active"`
	BillingInterval string    `gorm:"column:billing_interval"`
	Price           float64   `gorm:"column:price"`
	Currency        string    `gorm:"column:currency"`
	QuotaID         uuid.UUID `gorm:"column:quota_id"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (PlanModel) TableName() string {
	return "plans"
}
