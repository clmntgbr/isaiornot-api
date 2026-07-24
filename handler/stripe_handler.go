package handler

import (
	"context"
	"encoding/json"
	infraStripe "go-api/infrastructure/stripe"
	"go-api/usecase/subscription"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
)

type StripeHandler struct {
	checkoutCompletedUseCase       *subscription.CheckoutCompletedUseCase
	subscriptionUpdatedUseCase     *subscription.SubscriptionUpdatedUseCase
	subscriptionDeletedUseCase     *subscription.SubscriptionDeletedUseCase
	invoicePaymentSucceededUseCase *subscription.InvoicePaymentSucceededUseCase
	invoicePaymentFailedUseCase    *subscription.InvoicePaymentFailedUseCase
}

func NewStripeHandler(
	checkoutCompletedUseCase *subscription.CheckoutCompletedUseCase,
	subscriptionUpdatedUseCase *subscription.SubscriptionUpdatedUseCase,
	subscriptionDeletedUseCase *subscription.SubscriptionDeletedUseCase,
	invoicePaymentSucceededUseCase *subscription.InvoicePaymentSucceededUseCase,
	invoicePaymentFailedUseCase *subscription.InvoicePaymentFailedUseCase,
) *StripeHandler {
	return &StripeHandler{
		checkoutCompletedUseCase:       checkoutCompletedUseCase,
		subscriptionUpdatedUseCase:     subscriptionUpdatedUseCase,
		subscriptionDeletedUseCase:     subscriptionDeletedUseCase,
		invoicePaymentSucceededUseCase: invoicePaymentSucceededUseCase,
		invoicePaymentFailedUseCase:    invoicePaymentFailedUseCase,
	}
}

func (h *StripeHandler) Execute(c fiber.Ctx) error {
	event := c.Locals("payload").(stripe.Event)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := h.dispatch(ctx, event); err != nil {
			log.Printf("failed to handle stripe event %s: %v", event.Type, err)
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

func (h *StripeHandler) dispatch(ctx context.Context, event stripe.Event) error {
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
		return nil
	}
}

func (h *StripeHandler) handleCheckoutCompleted(ctx context.Context, event stripe.Event) error {
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

func (h *StripeHandler) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
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

func (h *StripeHandler) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	return h.subscriptionDeletedUseCase.Execute(ctx, sub.ID)
}

func (h *StripeHandler) handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return err
	}

	return h.invoicePaymentSucceededUseCase.Execute(ctx, subscription.InvoicePaymentSucceededInput{
		StripeSubscriptionID: subscriptionIDFromInvoice(&invoice),
		StripeCustomerID:     customerIDFromInvoice(&invoice),
		BillingReason:        string(invoice.BillingReason),
		PeriodStart:          unixToTime(invoice.PeriodStart),
		PeriodEnd:            unixToTime(invoice.PeriodEnd),
	})
}

func (h *StripeHandler) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return err
	}

	return h.invoicePaymentFailedUseCase.Execute(ctx, subscription.InvoicePaymentFailedInput{
		StripeSubscriptionID: subscriptionIDFromInvoice(&invoice),
		StripeCustomerID:     customerIDFromInvoice(&invoice),
	})
}

func subscriptionIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice.Parent != nil &&
		invoice.Parent.SubscriptionDetails != nil &&
		invoice.Parent.SubscriptionDetails.Subscription != nil {
		return invoice.Parent.SubscriptionDetails.Subscription.ID
	}
	return ""
}

func customerIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice.Customer != nil {
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
