package entity

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusInactive  SubscriptionStatus = "inactive"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusPending   SubscriptionStatus = "pending"
	SubscriptionStatusPastDue   SubscriptionStatus = "past_due"
)

type Subscription struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	PlanID uuid.UUID `gorm:"type:uuid;not null" json:"plan_id"`
	Plan   Plan      `gorm:"foreignKey:PlanID" json:"plan"`

	StripeCustomerID     string `gorm:"type:varchar(255);null" json:"stripe_customer_id"`
	StripeSubscriptionID string `gorm:"type:varchar(255);null" json:"stripe_subscription_id"`

	SubscriptionStatus    SubscriptionStatus `json:"subscription_status" gorm:"type:varchar(255);not null"`
	SubscriptionStartDate time.Time          `json:"subscription_start_date" gorm:"type:timestamp;not null"`
	SubscriptionEndDate   time.Time          `json:"subscription_end_date" gorm:"type:timestamp;not null"`

	// QuotaPeriodStart is the anniversary anchor used to compute the current
	// monthly quota window. It only changes on free ↔ paid transitions — not on
	// paid → paid plan changes or billing renewals.
	QuotaPeriodStart time.Time `json:"quota_period_start" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
