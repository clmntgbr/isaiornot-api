package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

type StripeHandler struct {
}

func NewStripeHandler() *StripeHandler {
	return &StripeHandler{}
}

func (h *StripeHandler) Execute(c fiber.Ctx) error {
	fmt.Println("Stripe handler executed")
	return nil
}
