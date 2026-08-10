package plan

import (
	"context"
	"errors"

	domainplan "go-api/internal/domain/plan"
)

type ListPlansHandler struct {
	planRepo domainplan.PlanReadRepository
}

func NewListPlansHandler(planRepo domainplan.PlanReadRepository) *ListPlansHandler {
	return &ListPlansHandler{planRepo: planRepo}
}

func (h *ListPlansHandler) Handle(ctx context.Context) ([]*domainplan.PlanView, error) {
	plans, err := h.planRepo.FindActive(ctx)
	if err != nil {
		return nil, errors.New("failed to list plans")
	}
	return plans, nil
}
