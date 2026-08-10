package media

import (
	"context"
	"time"

	domainsignal "go-api/internal/domain/signal"

	"github.com/google/uuid"
)

type MediaWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, media *Media) error
	Update(ctx context.Context, media *Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
	GetByKey(ctx context.Context, key string) (*Media, error)
	GetByScanID(ctx context.Context, scanID uuid.UUID) ([]*Media, error)
}

type MediaReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*MediaView, error)
	FindByScanID(ctx context.Context, scanID uuid.UUID) ([]*MediaView, error)
	FindByScanIDs(ctx context.Context, scanIDs []uuid.UUID) (map[uuid.UUID][]*MediaView, error)
	CountUsageInPeriod(ctx context.Context, userID uuid.UUID, from, to time.Time) (*UsageCounts, error)
}

type UsageCounts struct {
	Images int64
	Videos int64
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
	Signals     []*domainsignal.SignalView
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
