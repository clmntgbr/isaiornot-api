package http

import (
	"go-api/handler/http/context"
	"go-api/presenter"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewUserDetailResponse(*user))
}
