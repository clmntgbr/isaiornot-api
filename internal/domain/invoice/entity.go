package invoice

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID uuid.UUID

	UserID         uuid.UUID
	SubscriptionID *uuid.UUID

	StripeInvoiceID      string
	StripeCustomerID     string
	StripeSubscriptionID string

	Number   string
	Status   string
	Currency string

	AmountDue  int64
	AmountPaid int64
	Total      int64

	HostedInvoiceURL string
	InvoicePDF       string

	BillingReason string
	Description   string
	AttemptCount  int64

	PeriodStart     time.Time
	PeriodEnd       time.Time
	PaidAt          *time.Time
	StripeCreatedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewInvoice(userID uuid.UUID, subscriptionID *uuid.UUID, stripeInvoiceID string) *Invoice {
	now := time.Now().UTC()
	return &Invoice{
		ID:              uuid.New(),
		UserID:          userID,
		SubscriptionID:  subscriptionID,
		StripeInvoiceID: stripeInvoiceID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (i *Invoice) ApplyStripeSnapshot(
	stripeCustomerID string,
	stripeSubscriptionID string,
	number string,
	status string,
	currency string,
	amountDue int64,
	amountPaid int64,
	total int64,
	hostedInvoiceURL string,
	invoicePDF string,
	billingReason string,
	description string,
	attemptCount int64,
	periodStart time.Time,
	periodEnd time.Time,
	paidAt *time.Time,
	stripeCreatedAt time.Time,
) {
	i.StripeCustomerID = stripeCustomerID
	i.StripeSubscriptionID = stripeSubscriptionID
	i.Number = number
	i.Status = status
	i.Currency = currency
	i.AmountDue = amountDue
	i.AmountPaid = amountPaid
	i.Total = total
	i.HostedInvoiceURL = hostedInvoiceURL
	i.InvoicePDF = invoicePDF
	i.BillingReason = billingReason
	i.Description = description
	i.AttemptCount = attemptCount
	i.PeriodStart = periodStart
	i.PeriodEnd = periodEnd
	i.PaidAt = paidAt
	i.StripeCreatedAt = stripeCreatedAt
	i.UpdatedAt = time.Now().UTC()
}
