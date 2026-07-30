package repository

import (
	"context"
	"go-api/domain/entity"

	"github.com/google/uuid"
)

type SignalRepository interface {
	Create(ctx context.Context, signal *entity.Signal) error
	Update(ctx context.Context, signal *entity.Signal) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByMediaID(ctx context.Context, mediaID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Signal, error)
	GetByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*entity.Signal, error)
}
