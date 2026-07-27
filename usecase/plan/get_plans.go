package plan

import (
	"context"
	"go-api/domain/entity"
	"go-api/domain/repository"
)

type GetPlansUseCase struct {
	planRepo repository.PlanRepository
}

func NewGetPlansUseCase(planRepo repository.PlanRepository) *GetPlansUseCase {
	return &GetPlansUseCase{planRepo: planRepo}
}

func (u *GetPlansUseCase) Execute(ctx context.Context) ([]*entity.Plan, error) {
	plans, err := u.planRepo.GetAll(ctx)
	if err != nil {
		return []*entity.Plan{}, err
	}

	return plans, nil
}
