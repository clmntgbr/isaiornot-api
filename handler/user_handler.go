package handler

import (
	"errors"
	"go-api/handler/context"
	"go-api/presenter"
	"go-api/usecase/subscription"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	getUserSubscriptionUseCase *subscription.GetUserSubscriptionUseCase
}

func NewUserHandler(
	getUserSubscriptionUseCase *subscription.GetUserSubscriptionUseCase,
) *UserHandler {
	return &UserHandler{
		getUserSubscriptionUseCase: getUserSubscriptionUseCase,
	}
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

func (h *UserHandler) GetSubscription(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	result, err := h.getUserSubscriptionUseCase.Execute(c.Context(), user)
	if err != nil {
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Subscription not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewSubscriptionResponse(result.Subscription, result.EffectivePlan))
}
