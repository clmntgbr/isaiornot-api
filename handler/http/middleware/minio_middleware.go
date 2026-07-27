package middleware

import (
	"github.com/gofiber/fiber/v3"
)

type MinIOMiddleware struct {
	secret string
}

func NewMinIOMiddleware(secret string) *MinIOMiddleware {
	return &MinIOMiddleware{
		secret: secret,
	}
}

func (m *MinIOMiddleware) Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		if m.secret == "" {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		expected := "Bearer " + m.secret

		if authHeader != expected && authHeader != m.secret {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization",
			})
		}

		return c.Next()
	}
}
