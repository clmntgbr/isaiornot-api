package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type GetUserSubscriptionUseCase struct {
	subscriptionRepo *repository.SubscriptionRepository
}

func NewGetUserSubscriptionUseCase(
	subscriptionRepo *repository.SubscriptionRepository,
) *GetUserSubscriptionUseCase {
	return &GetUserSubscriptionUseCase{
		subscriptionRepo: subscriptionRepo,
	}
}

func (u *GetUserSubscriptionUseCase) Execute(ctx context.Context, user *entity.User) (*entity.Subscription, error) {
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

	return subscription, nil
}
