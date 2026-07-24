package repository

import (
	"context"
	"go-api/domain/entity"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *entity.Subscription) error
	Update(ctx context.Context, subscription *entity.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)
	GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*entity.Subscription, error)
	GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*entity.Subscription, error)
}
