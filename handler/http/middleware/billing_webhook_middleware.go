package middleware

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type BillingWebhookMiddleware struct {
	secret string
}

func NewBillingWebhookMiddleware(secret string) *BillingWebhookMiddleware {
	return &BillingWebhookMiddleware{
		secret: secret,
	}
}

func (m *BillingWebhookMiddleware) Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		if m.secret == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "stripe webhook secret is not configured",
			})
		}

		payload := c.Body()
		signature := c.Get("Stripe-Signature")
		if signature == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing stripe signature",
			})
		}

		if err := webhook.ValidatePayload(payload, signature, m.secret); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid signature",
			})
		}

		var event stripe.Event
		if err := json.Unmarshal(payload, &event); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid payload",
			})
		}

		c.Locals("payload", event)

		return c.Next()
	}
}
