package subscription

import (
	"context"
	"errors"
	"log"
	"time"

	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type UpsertInvoiceUseCase struct {
	invoiceRepo      repository.InvoiceRepository
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
}

func NewUpsertInvoiceUseCase(
	invoiceRepo repository.InvoiceRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
) *UpsertInvoiceUseCase {
	return &UpsertInvoiceUseCase{
		invoiceRepo:      invoiceRepo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
	}
}

type UpsertInvoiceInput struct {
	StripeInvoiceID      string
	StripeCustomerID     string
	StripeSubscriptionID string
	Number               string
	Status               string
	Currency             string
	AmountDue            int64
	AmountPaid           int64
	Total                int64
	HostedInvoiceURL     string
	InvoicePDF           string
	BillingReason        string
	Description          string
	AttemptCount         int64
	PeriodStart          time.Time
	PeriodEnd            time.Time
	PaidAt               *time.Time
	StripeCreatedAt      time.Time
}

func (u *UpsertInvoiceUseCase) Execute(ctx context.Context, input UpsertInvoiceInput) error {
	log.Printf(
		"upsert invoice: start stripeInvoiceID=%s stripeSubscriptionID=%s stripeCustomerID=%s status=%s total=%d",
		input.StripeInvoiceID,
		input.StripeSubscriptionID,
		input.StripeCustomerID,
		input.Status,
		input.Total,
	)

	if input.StripeInvoiceID == "" {
		log.Printf("upsert invoice: skip, missing stripeInvoiceID")
		return nil
	}

	userID, subscriptionID, err := u.resolveOwner(ctx, input.StripeSubscriptionID, input.StripeCustomerID)
	if err != nil {
		log.Printf("upsert invoice: resolve owner failed: %v", err)
		return err
	}
	if userID == uuid.Nil {
		log.Printf(
			"upsert invoice: skip, no local user/subscription for stripeSubscriptionID=%s stripeCustomerID=%s",
			input.StripeSubscriptionID,
			input.StripeCustomerID,
		)
		return nil
	}

	subscriptionIDValue := ""
	if subscriptionID != nil {
		subscriptionIDValue = subscriptionID.String()
	}
	log.Printf(
		"upsert invoice: resolved userID=%s subscriptionID=%s",
		userID,
		subscriptionIDValue,
	)

	invoice := &entity.Invoice{
		UserID:               userID,
		SubscriptionID:       subscriptionID,
		StripeInvoiceID:      input.StripeInvoiceID,
		StripeCustomerID:     input.StripeCustomerID,
		StripeSubscriptionID: input.StripeSubscriptionID,
		Number:               input.Number,
		Status:               input.Status,
		Currency:             input.Currency,
		AmountDue:            input.AmountDue,
		AmountPaid:           input.AmountPaid,
		Total:                input.Total,
		HostedInvoiceURL:     input.HostedInvoiceURL,
		InvoicePDF:           input.InvoicePDF,
		BillingReason:        input.BillingReason,
		Description:          input.Description,
		AttemptCount:         input.AttemptCount,
		PeriodStart:          input.PeriodStart,
		PeriodEnd:            input.PeriodEnd,
		PaidAt:               input.PaidAt,
		StripeCreatedAt:      input.StripeCreatedAt,
	}

	if err := u.invoiceRepo.UpsertByStripeInvoiceID(ctx, invoice); err != nil {
		log.Printf("upsert invoice: db upsert failed stripeInvoiceID=%s: %v", input.StripeInvoiceID, err)
		return errors.New("failed to upsert invoice")
	}

	log.Printf("upsert invoice: ok id=%s stripeInvoiceID=%s", invoice.ID, invoice.StripeInvoiceID)
	return nil
}

func (u *UpsertInvoiceUseCase) resolveOwner(
	ctx context.Context,
	stripeSubscriptionID string,
	stripeCustomerID string,
) (uuid.UUID, *uuid.UUID, error) {
	var (
		subscription *entity.Subscription
		err          error
	)

	if stripeSubscriptionID != "" {
		subscription, err = u.subscriptionRepo.GetByStripeSubscriptionID(ctx, stripeSubscriptionID)
		if err != nil {
			return uuid.Nil, nil, errors.New("failed to get subscription")
		}
		if subscription != nil {
			log.Printf("upsert invoice: found subscription by stripeSubscriptionID=%s id=%s", stripeSubscriptionID, subscription.ID)
		} else {
			log.Printf("upsert invoice: no subscription for stripeSubscriptionID=%s", stripeSubscriptionID)
		}
	}

	if subscription == nil && stripeCustomerID != "" {
		subscription, err = u.subscriptionRepo.GetByStripeCustomerID(ctx, stripeCustomerID)
		if err != nil {
			return uuid.Nil, nil, errors.New("failed to get subscription by customer")
		}
		if subscription != nil {
			log.Printf("upsert invoice: found subscription by stripeCustomerID=%s id=%s", stripeCustomerID, subscription.ID)
		} else {
			log.Printf("upsert invoice: no subscription for stripeCustomerID=%s", stripeCustomerID)
		}
	}

	if subscription == nil {
		return uuid.Nil, nil, nil
	}

	user, err := u.userRepo.GetBySubscriptionID(ctx, subscription.ID)
	if err != nil {
		return uuid.Nil, nil, errors.New("failed to get user by subscription")
	}
	if user == nil {
		log.Printf("upsert invoice: no user linked to subscriptionID=%s", subscription.ID)
		return uuid.Nil, nil, nil
	}

	subscriptionID := subscription.ID
	return user.ID, &subscriptionID, nil
}
