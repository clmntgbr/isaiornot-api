package entity

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	UserID         uuid.UUID  `gorm:"type:uuid;not null;index:idx_invoices_user_id" json:"userId"`
	SubscriptionID *uuid.UUID `gorm:"type:uuid;index:idx_invoices_subscription_id" json:"subscriptionId"`

	StripeInvoiceID      string `gorm:"type:varchar(255);not null;uniqueIndex" json:"stripeInvoiceId"`
	StripeCustomerID     string `gorm:"type:varchar(255);null" json:"stripeCustomerId"`
	StripeSubscriptionID string `gorm:"type:varchar(255);null" json:"stripeSubscriptionId"`

	Number   string `gorm:"type:varchar(255);null" json:"number"`
	Status   string `gorm:"type:varchar(64);not null" json:"status"`
	Currency string `gorm:"type:varchar(16);not null" json:"currency"`

	AmountDue  int64 `gorm:"not null;default:0" json:"amountDue"`
	AmountPaid int64 `gorm:"not null;default:0" json:"amountPaid"`
	Total      int64 `gorm:"not null;default:0" json:"total"`

	HostedInvoiceURL string `gorm:"type:text;null" json:"hostedInvoiceUrl"`
	InvoicePDF       string `gorm:"type:text;null" json:"invoicePdf"`

	BillingReason string `gorm:"type:varchar(255);null" json:"billingReason"`
	Description   string `gorm:"type:text;null" json:"description"`
	AttemptCount  int64  `gorm:"not null;default:0" json:"attemptCount"`

	PeriodStart     time.Time  `gorm:"type:timestamp;null" json:"periodStart"`
	PeriodEnd       time.Time  `gorm:"type:timestamp;null" json:"periodEnd"`
	PaidAt          *time.Time `gorm:"type:timestamp;null" json:"paidAt"`
	StripeCreatedAt time.Time  `gorm:"type:timestamp;not null" json:"stripeCreatedAt"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Invoice) TableName() string {
	return "invoices"
}
