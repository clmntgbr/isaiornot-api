package repository

import (
	"context"
	"go-api/domain/entity"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByClerkID(ctx context.Context, clerkID string) (*entity.User, error)
	GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByClerkID(ctx context.Context, clerkID string) error
}
