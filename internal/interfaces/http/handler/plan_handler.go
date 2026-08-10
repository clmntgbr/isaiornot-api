package handler

import (
	queryplan "go-api/internal/application/query/plan"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
)

type PlanHandler struct {
	listPlansHandler *queryplan.ListPlansHandler
}

func NewPlanHandler(listPlansHandler *queryplan.ListPlansHandler) *PlanHandler {
	return &PlanHandler{listPlansHandler: listPlansHandler}
}

func (h *PlanHandler) GetPlans(c fiber.Ctx) error {
	plans, err := h.listPlansHandler.Handle(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to list plans",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewPlanResponses(plans))
}
