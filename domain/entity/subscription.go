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

	PlanID uuid.UUID `gorm:"type:uuid;not null" json:"planId"`
	Plan   Plan      `gorm:"foreignKey:PlanID" json:"plan"`

	StripeCustomerID     string `gorm:"type:varchar(255);null" json:"stripeCustomerId"`
	StripeSubscriptionID string `gorm:"type:varchar(255);null" json:"stripeSubscriptionId"`

	SubscriptionStatus    SubscriptionStatus `json:"subscriptionStatus" gorm:"type:varchar(255);not null"`
	SubscriptionStartDate time.Time          `json:"subscriptionStartDate" gorm:"type:timestamp;not null"`
	SubscriptionEndDate   time.Time          `json:"subscriptionEndDate" gorm:"type:timestamp;not null"`

	CancelAtPeriodEnd bool `json:"cancelAtPeriodEnd" gorm:"type:boolean;not null;default:false"`

	QuotaPeriodStart time.Time `json:"quotaPeriodStart" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
