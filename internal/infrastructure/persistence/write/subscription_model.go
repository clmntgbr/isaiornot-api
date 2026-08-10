package write

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionModel struct {
	ID                   uuid.UUID `gorm:"column:id;primaryKey"`
	PlanID               uuid.UUID `gorm:"column:plan_id"`
	StripeCustomerID     string    `gorm:"column:stripe_customer_id"`
	StripeSubscriptionID string    `gorm:"column:stripe_subscription_id"`
	Status               string    `gorm:"column:status"`
	StartDate            time.Time `gorm:"column:start_date"`
	EndDate              time.Time `gorm:"column:end_date"`
	CancelAtPeriodEnd    bool      `gorm:"column:cancel_at_period_end"`
	QuotaPeriodStart     time.Time `gorm:"column:quota_period_start"`
	CreatedAt            time.Time `gorm:"column:created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

func (SubscriptionModel) TableName() string {
	return "subscriptions"
}
