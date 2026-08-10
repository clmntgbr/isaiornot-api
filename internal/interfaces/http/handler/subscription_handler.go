package handler

import (
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
)

type SubscriptionHandler struct {
	getCurrentSubscriptionHandler *querysubscription.GetCurrentSubscriptionHandler
	getQuotaUsageHandler          *querysubscription.GetQuotaUsageHandler
}

func NewSubscriptionHandler(
	getCurrentSubscriptionHandler *querysubscription.GetCurrentSubscriptionHandler,
	getQuotaUsageHandler *querysubscription.GetQuotaUsageHandler,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		getCurrentSubscriptionHandler: getCurrentSubscriptionHandler,
		getQuotaUsageHandler:          getQuotaUsageHandler,
	}
}

func (h *SubscriptionHandler) GetSubscription(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	view, err := h.getCurrentSubscriptionHandler.Handle(c.Context(), querysubscription.GetCurrentSubscriptionQuery{
		UserID: user.ID,
	})

	if err != nil {
		if errors.Is(err, querysubscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Subscription not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get subscription",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewSubscriptionResponse(view))
}

func (h *SubscriptionHandler) GetQuota(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	usage, err := h.getQuotaUsageHandler.Handle(c.Context(), querysubscription.GetQuotaUsageQuery{
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, querysubscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Subscription not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get quota usage",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewQuotaUsageResponse(usage))
}
