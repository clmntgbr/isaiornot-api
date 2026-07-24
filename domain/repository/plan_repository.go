package repository

import (
	"context"
	"go-api/domain/entity"

	"github.com/google/uuid"
)

type PlanRepository interface {
	Create(ctx context.Context, plan *entity.Plan) error
	Update(ctx context.Context, plan *entity.Plan) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*entity.Plan, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Plan, error)
	GetByStripePriceID(ctx context.Context, stripePriceID string) (*entity.Plan, error)
}
