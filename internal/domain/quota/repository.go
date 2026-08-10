package quota

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type QuotaWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, quota *Quota) error
	Update(ctx context.Context, quota *Quota) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Quota, error)
}

type QuotaReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*QuotaView, error)
	FindAll(ctx context.Context) ([]*QuotaView, error)
}

type QuotaView struct {
	ID                uuid.UUID
	MaxImagesPerMonth int
	MaxVideosPerMonth int
	MaxFileSizeImage  int64
	MaxFileSizeVideo  int64
	FullPipeline      bool
	HistoryRetention  time.Duration
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
