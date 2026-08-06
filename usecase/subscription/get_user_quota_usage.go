package subscription

import (
	"context"
	"errors"

	"go-api/domain/entity"
	"go-api/domain/repository"
)

type GetUserQuotaUsageUseCase struct {
	subscriptionRepo            repository.SubscriptionRepository
	resolveEffectivePlanUseCase *ResolveEffectivePlanUseCase
	getQuotaUsageUseCase        *GetQuotaUsageUseCase
}

func NewGetUserQuotaUsageUseCase(
	subscriptionRepo repository.SubscriptionRepository,
	resolveEffectivePlanUseCase *ResolveEffectivePlanUseCase,
	getQuotaUsageUseCase *GetQuotaUsageUseCase,
) *GetUserQuotaUsageUseCase {
	return &GetUserQuotaUsageUseCase{
		subscriptionRepo:            subscriptionRepo,
		resolveEffectivePlanUseCase: resolveEffectivePlanUseCase,
		getQuotaUsageUseCase:        getQuotaUsageUseCase,
	}
}

func (u *GetUserQuotaUsageUseCase) Execute(ctx context.Context, user *entity.User) (*QuotaUsage, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}
	if user.SubscriptionID == nil {
		return nil, ErrSubscriptionNotFound
	}

	subscription, err := u.subscriptionRepo.GetByID(ctx, *user.SubscriptionID)
	if err != nil {
		return nil, errors.New("failed to get subscription")
	}
	if subscription == nil {
		return nil, ErrSubscriptionNotFound
	}

	effectivePlan, err := u.resolveEffectivePlanUseCase.Execute(ctx, subscription)
	if err != nil {
		return nil, err
	}

	return u.getQuotaUsageUseCase.Execute(ctx, user, subscription, effectivePlan)
}
