package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/stripe/stripe-go/v82"
)

type StripeHandler struct {
}

func NewStripeHandler() *StripeHandler {
	return &StripeHandler{}
}

func (h *StripeHandler) Execute(c fiber.Ctx) error {
	stripeEvent := c.Locals("payload").(stripe.Event)

	fmt.Printf("%+v", stripeEvent)

	return nil
}
