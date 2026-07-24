package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"time"
)

type CreateFreeSubscriptionUseCase struct {
	planRepo         *repository.PlanRepository
	subscriptionRepo *repository.SubscriptionRepository
	userRepo         *repository.UserRepository
}

func NewCreateFreeSubscriptionUseCase(
	planRepo *repository.PlanRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	userRepo *repository.UserRepository,
) *CreateFreeSubscriptionUseCase {
	return &CreateFreeSubscriptionUseCase{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
	}
}

func (u *CreateFreeSubscriptionUseCase) Execute(ctx context.Context, user *entity.User) (*entity.Subscription, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}

	plan, err := (*u.planRepo).GetBySlug(ctx, entity.FreePlanSlug)
	if err != nil {
		return nil, errors.New("failed to get free plan")
	}
	if plan == nil {
		return nil, errors.New("free plan not found")
	}

	now := time.Now().UTC()
	subscription := entity.Subscription{
		PlanID:                plan.ID,
		SubscriptionStatus:    entity.SubscriptionStatusActive,
		SubscriptionStartDate: now,
		SubscriptionEndDate:   now.AddDate(100, 0, 0),
		QuotaPeriodStart:      now,
	}

	if err := (*u.subscriptionRepo).Create(ctx, &subscription); err != nil {
		return nil, errors.New("failed to create free subscription")
	}

	user.SubscriptionID = &subscription.ID
	user.Subscription = &subscription
	if err := (*u.userRepo).Update(ctx, user); err != nil {
		return nil, errors.New("failed to link subscription to user")
	}

	return &subscription, nil
}
