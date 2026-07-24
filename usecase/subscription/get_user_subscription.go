package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type GetUserSubscriptionUseCase struct {
	subscriptionRepo            *repository.SubscriptionRepository
	resolveEffectivePlanUseCase *ResolveEffectivePlanUseCase
	getQuotaUsageUseCase        *GetQuotaUsageUseCase
}

func NewGetUserSubscriptionUseCase(
	subscriptionRepo *repository.SubscriptionRepository,
	resolveEffectivePlanUseCase *ResolveEffectivePlanUseCase,
	getQuotaUsageUseCase *GetQuotaUsageUseCase,
) *GetUserSubscriptionUseCase {
	return &GetUserSubscriptionUseCase{
		subscriptionRepo:            subscriptionRepo,
		resolveEffectivePlanUseCase: resolveEffectivePlanUseCase,
		getQuotaUsageUseCase:        getQuotaUsageUseCase,
	}
}

type UserSubscription struct {
	Subscription  *entity.Subscription
	EffectivePlan *entity.Plan
	QuotaUsage    *QuotaUsage
}

func (u *GetUserSubscriptionUseCase) Execute(ctx context.Context, user *entity.User) (*UserSubscription, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}
	if user.SubscriptionID == nil {
		return nil, ErrSubscriptionNotFound
	}

	subscription, err := (*u.subscriptionRepo).GetByID(ctx, *user.SubscriptionID)
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

	quotaUsage, err := u.getQuotaUsageUseCase.Execute(ctx, user, subscription, effectivePlan)
	if err != nil {
		return nil, err
	}

	return &UserSubscription{
		Subscription:  subscription,
		EffectivePlan: effectivePlan,
		QuotaUsage:    quotaUsage,
	}, nil
}
