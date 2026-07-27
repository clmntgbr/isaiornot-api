package middleware

import (
	"encoding/json"
	"go-api/handler/http/dto"
	"net/http"

	"github.com/gofiber/fiber/v3"
	svix "github.com/svix/svix-webhooks/go"
)

type ClerkMiddleware struct {
	secret string
}

func NewClerkMiddleware(secret string) *ClerkMiddleware {
	return &ClerkMiddleware{
		secret: secret,
	}
}

func (m *ClerkMiddleware) Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		wh, err := svix.NewWebhook(m.secret)
		if err != nil {
			panic(err)
		}

		payload := c.Body()

		headers := http.Header{}
		headers.Set("svix-id", c.Get("svix-id"))
		headers.Set("svix-timestamp", c.Get("svix-timestamp"))
		headers.Set("svix-signature", c.Get("svix-signature"))

		if err := wh.Verify(payload, headers); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid signature",
			})
		}

		var clerkEvent dto.ClerkEvent
		if err := json.Unmarshal(payload, &clerkEvent); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid payload",
			})
		}

		c.Locals("payload", clerkEvent)

		return c.Next()
	}
}
