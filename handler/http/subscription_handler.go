package http

import (
	"errors"

	"go-api/handler/http/context"
	"go-api/handler/http/dto"
	"go-api/presenter"
	"go-api/usecase/subscription"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	createSubscriptionUseCase  *subscription.CreateSubscriptionUseCase
	previewPlanChangeUseCase   *subscription.PreviewPlanChangeUseCase
	createBillingPortalUseCase *subscription.CreateBillingPortalUseCase
	getUserSubscriptionUseCase *subscription.GetUserSubscriptionUseCase
	getUserQuotaUsageUseCase   *subscription.GetUserQuotaUsageUseCase
}

func NewSubscriptionHandler(
	createSubscriptionUseCase *subscription.CreateSubscriptionUseCase,
	previewPlanChangeUseCase *subscription.PreviewPlanChangeUseCase,
	createBillingPortalUseCase *subscription.CreateBillingPortalUseCase,
	getUserSubscriptionUseCase *subscription.GetUserSubscriptionUseCase,
	getUserQuotaUsageUseCase *subscription.GetUserQuotaUsageUseCase,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		createSubscriptionUseCase:  createSubscriptionUseCase,
		previewPlanChangeUseCase:   previewPlanChangeUseCase,
		createBillingPortalUseCase: createBillingPortalUseCase,
		getUserSubscriptionUseCase: getUserSubscriptionUseCase,
		getUserQuotaUsageUseCase:   getUserQuotaUsageUseCase,
	}
}

func (h *SubscriptionHandler) GetSubscription(c fiber.Ctx) error {
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

	return c.Status(fiber.StatusOK).JSON(presenter.NewSubscriptionResponse(
		result.Subscription,
		result.EffectivePlan,
	))
}

func (h *SubscriptionHandler) GetQuota(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	usage, err := h.getUserQuotaUsageUseCase.Execute(c.Context(), user)
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

	return c.Status(fiber.StatusOK).JSON(presenter.NewQuotaUsageResponse(usage))
}

func (h *SubscriptionHandler) PreviewSubscription(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var request dto.PreviewSubscriptionRequest
	if err := c.Bind().JSON(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"errors":  err.Error(),
		})
	}

	planID, err := uuid.Parse(request.PlanID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid planId",
		})
	}

	preview, err := h.previewPlanChangeUseCase.Execute(c.Context(), user, planID)
	if err != nil {
		switch {
		case errors.Is(err, subscription.ErrPlanNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Plan not found",
			})
		case errors.Is(err, subscription.ErrPlanInactive),
			errors.Is(err, subscription.ErrFreePlanCheckout),
			errors.Is(err, subscription.ErrMissingStripePrice),
			errors.Is(err, subscription.ErrAlreadyOnPlan):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
				"errors":  err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewPlanChangePreviewResponse(preview))
}

func (h *SubscriptionHandler) CreateSubscription(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var request dto.CreateSubscriptionRequest
	if err := c.Bind().JSON(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"errors":  err.Error(),
		})
	}

	planID, err := uuid.Parse(request.PlanID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid planId",
		})
	}

	result, err := h.createSubscriptionUseCase.Execute(c.Context(), user, subscription.ChangeSubscriptionInput{
		PlanID:        planID,
		ProrationDate: request.ProrationDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, subscription.ErrPlanNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Plan not found",
			})
		case errors.Is(err, subscription.ErrPlanInactive),
			errors.Is(err, subscription.ErrFreePlanCheckout),
			errors.Is(err, subscription.ErrMissingStripePrice),
			errors.Is(err, subscription.ErrAlreadyOnPlan):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
				"errors":  err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewChangeSubscriptionResponse(result))
}

func (h *SubscriptionHandler) CreateBillingPortal(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	url, err := h.createBillingPortalUseCase.Execute(c.Context(), user)
	if err != nil {
		switch {
		case errors.Is(err, subscription.ErrSubscriptionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Subscription not found",
			})
		case errors.Is(err, subscription.ErrMissingStripeCustomer):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "User has no Stripe customer",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
				"errors":  err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewCheckoutSessionResponse(url))
}
