package http

import (
	"go-api/presenter"
	"go-api/usecase/plan"

	"github.com/gofiber/fiber/v3"
)

type PlanHandler struct {
	getPlansUseCase *plan.GetPlansUseCase
}

func NewPlanHandler(
	getPlansUseCase *plan.GetPlansUseCase,
) *PlanHandler {
	return &PlanHandler{
		getPlansUseCase: getPlansUseCase,
	}
}

func (h *PlanHandler) GetPlans(c fiber.Ctx) error {
	plans, err := h.getPlansUseCase.Execute(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.JSON(presenter.NewPlanResponses(plans))
}
