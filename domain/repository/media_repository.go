package repository

import (
	"context"
	"go-api/domain/entity"
	"time"

	"github.com/google/uuid"
)

type MediaUsageCounts struct {
	Images int64
	Videos int64
}

type MediaRepository interface {
	Create(ctx context.Context, media *entity.Media) error
	Update(ctx context.Context, media *entity.Media) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Media, error)
	GetByKey(ctx context.Context, key string) (*entity.Media, error)
	CountUsageInPeriod(ctx context.Context, userID uuid.UUID, from, to time.Time) (*MediaUsageCounts, error)
}
