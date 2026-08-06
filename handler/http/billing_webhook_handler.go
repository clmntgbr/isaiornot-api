package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	infraStripe "go-api/infrastructure/stripe"
	"go-api/usecase/subscription"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
)

type BillingWebhookHandler struct {
	mu                             sync.Mutex
	checkoutCompletedUseCase       *subscription.CheckoutCompletedUseCase
	subscriptionUpdatedUseCase     *subscription.SubscriptionUpdatedUseCase
	subscriptionDeletedUseCase     *subscription.SubscriptionDeletedUseCase
	invoicePaymentSucceededUseCase *subscription.InvoicePaymentSucceededUseCase
	invoicePaymentFailedUseCase    *subscription.InvoicePaymentFailedUseCase
	upsertInvoiceUseCase           *subscription.UpsertInvoiceUseCase
}

func NewBillingWebhookHandler(
	checkoutCompletedUseCase *subscription.CheckoutCompletedUseCase,
	subscriptionUpdatedUseCase *subscription.SubscriptionUpdatedUseCase,
	subscriptionDeletedUseCase *subscription.SubscriptionDeletedUseCase,
	invoicePaymentSucceededUseCase *subscription.InvoicePaymentSucceededUseCase,
	invoicePaymentFailedUseCase *subscription.InvoicePaymentFailedUseCase,
	upsertInvoiceUseCase *subscription.UpsertInvoiceUseCase,
) *BillingWebhookHandler {
	return &BillingWebhookHandler{
		checkoutCompletedUseCase:       checkoutCompletedUseCase,
		subscriptionUpdatedUseCase:     subscriptionUpdatedUseCase,
		subscriptionDeletedUseCase:     subscriptionDeletedUseCase,
		invoicePaymentSucceededUseCase: invoicePaymentSucceededUseCase,
		invoicePaymentFailedUseCase:    invoicePaymentFailedUseCase,
		upsertInvoiceUseCase:           upsertInvoiceUseCase,
	}
}

func (h *BillingWebhookHandler) Execute(c fiber.Ctx) error {
	event := c.Locals("payload").(stripe.Event)
	log.Printf("stripe webhook: received event id=%s type=%s", event.ID, event.Type)

	h.mu.Lock()
	defer h.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("stripe webhook: processing event id=%s type=%s", event.ID, event.Type)
	if err := h.dispatch(ctx, event); err != nil {
		log.Printf("stripe webhook: failed event id=%s type=%s: %v", event.ID, event.Type, err)
		if errors.Is(err, subscription.ErrStripeSubscriptionNotLinked) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "subscription not linked yet",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process event",
		})
	}

	log.Printf("stripe webhook: processed event id=%s type=%s", event.ID, event.Type)
	return c.SendStatus(fiber.StatusOK)
}

func (h *BillingWebhookHandler) dispatch(ctx context.Context, event stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		return h.handleCheckoutCompleted(ctx, event)

	case "customer.subscription.updated":
		return h.handleSubscriptionUpdated(ctx, event)

	case "customer.subscription.deleted":
		return h.handleSubscriptionDeleted(ctx, event)

	case "invoice.payment_succeeded":
		return h.handleInvoicePaymentSucceeded(ctx, event)

	case "invoice.payment_failed":
		return h.handleInvoicePaymentFailed(ctx, event)

	default:
		log.Printf("stripe webhook: ignoring unhandled event id=%s type=%s", event.ID, event.Type)
		return nil
	}
}

func (h *BillingWebhookHandler) handleCheckoutCompleted(ctx context.Context, event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return err
	}

	userID, err := uuid.Parse(session.ClientReferenceID)
	if err != nil {
		return err
	}

	input := subscription.CheckoutCompletedInput{
		UserID: userID,
	}
	if session.Customer != nil {
		input.StripeCustomerID = session.Customer.ID
	}
	if session.Subscription != nil {
		input.StripeSubscriptionID = session.Subscription.ID
	}

	return h.checkoutCompletedUseCase.Execute(ctx, input)
}

func (h *BillingWebhookHandler) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	data := infraStripe.ExtractSubscriptionData(&sub)

	return h.subscriptionUpdatedUseCase.Execute(ctx, subscription.SubscriptionUpdatedInput{
		StripeSubscriptionID: data.ID,
		StripeCustomerID:     data.CustomerID,
		StripePriceID:        data.PriceID,
		Status:               data.Status,
		CancelAtPeriodEnd:    data.CancelAtPeriodEnd,
		CurrentPeriodStart:   data.CurrentPeriodStart,
		CurrentPeriodEnd:     data.CurrentPeriodEnd,
	})
}

func (h *BillingWebhookHandler) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	return h.subscriptionDeletedUseCase.Execute(ctx, sub.ID)
}

func (h *BillingWebhookHandler) handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("stripe webhook: failed to unmarshal invoice.payment_succeeded: %v", err)
		return err
	}

	upsertInput := upsertInvoiceInputFromStripe(&invoice)
	log.Printf(
		"stripe webhook: invoice.payment_succeeded mapped invoiceID=%s subscriptionID=%s customerID=%s parent=%v",
		upsertInput.StripeInvoiceID,
		upsertInput.StripeSubscriptionID,
		upsertInput.StripeCustomerID,
		invoice.Parent != nil,
	)

	if err := h.upsertInvoiceUseCase.Execute(ctx, upsertInput); err != nil {
		return err
	}

	return h.invoicePaymentSucceededUseCase.Execute(ctx, subscription.InvoicePaymentSucceededInput{
		StripeSubscriptionID: upsertInput.StripeSubscriptionID,
		StripeCustomerID:     upsertInput.StripeCustomerID,
		BillingReason:        string(invoice.BillingReason),
		PeriodStart:          unixToTime(invoice.PeriodStart),
		PeriodEnd:            unixToTime(invoice.PeriodEnd),
	})
}

func (h *BillingWebhookHandler) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return err
	}

	upsertInput := upsertInvoiceInputFromStripe(&invoice)
	if err := h.upsertInvoiceUseCase.Execute(ctx, upsertInput); err != nil {
		return err
	}

	return h.invoicePaymentFailedUseCase.Execute(ctx, subscription.InvoicePaymentFailedInput{
		StripeSubscriptionID: upsertInput.StripeSubscriptionID,
		StripeCustomerID:     upsertInput.StripeCustomerID,
	})
}

func upsertInvoiceInputFromStripe(invoice *stripe.Invoice) subscription.UpsertInvoiceInput {
	input := subscription.UpsertInvoiceInput{
		StripeInvoiceID:      invoice.ID,
		StripeCustomerID:     customerIDFromInvoice(invoice),
		StripeSubscriptionID: subscriptionIDFromInvoice(invoice),
		Number:               invoice.Number,
		Status:               string(invoice.Status),
		Currency:             string(invoice.Currency),
		AmountDue:            invoice.AmountDue,
		AmountPaid:           invoice.AmountPaid,
		Total:                invoice.Total,
		HostedInvoiceURL:     invoice.HostedInvoiceURL,
		InvoicePDF:           invoice.InvoicePDF,
		BillingReason:        string(invoice.BillingReason),
		Description:          invoiceDescription(invoice),
		AttemptCount:         invoice.AttemptCount,
		PeriodStart:          unixToTime(invoice.PeriodStart),
		PeriodEnd:            unixToTime(invoice.PeriodEnd),
		StripeCreatedAt:      unixToTime(invoice.Created),
	}

	if invoice.StatusTransitions != nil {
		if paidAt := unixToTime(invoice.StatusTransitions.PaidAt); !paidAt.IsZero() {
			input.PaidAt = &paidAt
		}
	}

	return input
}

func invoiceDescription(invoice *stripe.Invoice) string {
	if invoice.Lines == nil || len(invoice.Lines.Data) == 0 {
		return ""
	}
	return invoice.Lines.Data[0].Description
}

func subscriptionIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice.Parent != nil &&
		invoice.Parent.SubscriptionDetails != nil &&
		invoice.Parent.SubscriptionDetails.Subscription != nil &&
		invoice.Parent.SubscriptionDetails.Subscription.ID != "" {
		return invoice.Parent.SubscriptionDetails.Subscription.ID
	}

	if invoice.Lines != nil {
		for _, line := range invoice.Lines.Data {
			if line == nil || line.Parent == nil || line.Parent.SubscriptionItemDetails == nil {
				continue
			}
			if line.Parent.SubscriptionItemDetails.Subscription != "" {
				return line.Parent.SubscriptionItemDetails.Subscription
			}
		}
	}

	return ""
}

func customerIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice.Customer != nil && invoice.Customer.ID != "" {
		return invoice.Customer.ID
	}
	return ""
}

func unixToTime(ts int64) time.Time {
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0).UTC()
}
