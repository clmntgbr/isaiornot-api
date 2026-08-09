package media

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MediaWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, media *Media) error
	Update(ctx context.Context, media *Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
	GetByKey(ctx context.Context, key string) (*Media, error)
}

type MediaReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*MediaView, error)
	FindByScanID(ctx context.Context, scanID uuid.UUID) ([]*MediaView, error)
}

type MediaView struct {
	ID          uuid.UUID
	ScanID      uuid.UUID
	Key         string
	Filename    string
	Thumbnail   string
	ContentType string
	Size        int64
	Status      Status
	Statuses    []Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
