package repository

import (
	"context"
	"time"

	"go-api/domain/entity"
	"go-api/domain/scan"

	"github.com/google/uuid"
)

type ScanRepository interface {
	Create(ctx context.Context, scan *entity.Scan) error
	Update(ctx context.Context, scan *entity.Scan) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Scan, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, query scan.ListQuery, since time.Time) ([]*entity.Scan, int64, error)
	GetStatisticsByUserID(ctx context.Context, userID uuid.UUID, since time.Time) (*entity.MediaStatistics, error)
}
