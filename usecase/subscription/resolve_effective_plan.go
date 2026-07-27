package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
)

type ResolveEffectivePlanUseCase struct {
	planRepo repository.PlanRepository
}

func NewResolveEffectivePlanUseCase(
	planRepo repository.PlanRepository,
) *ResolveEffectivePlanUseCase {
	return &ResolveEffectivePlanUseCase{
		planRepo: planRepo,
	}
}

func (u *ResolveEffectivePlanUseCase) Execute(ctx context.Context, subscription *entity.Subscription) (*entity.Plan, error) {
	if subscription != nil && subscription.SubscriptionStatus == entity.SubscriptionStatusActive {
		return &subscription.Plan, nil
	}

	if subscription != nil && subscription.Plan.Slug == entity.FreePlanSlug {
		return &subscription.Plan, nil
	}

	freePlan, err := u.planRepo.GetBySlug(ctx, entity.FreePlanSlug)
	if err != nil {
		return nil, errors.New("failed to get free plan")
	}
	if freePlan == nil {
		return nil, errors.New("free plan not found")
	}

	return freePlan, nil
}
