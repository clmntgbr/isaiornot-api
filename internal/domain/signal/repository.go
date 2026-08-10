package signal

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SignalWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, signal *Signal) error
	Update(ctx context.Context, signal *Signal) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByMediaID(ctx context.Context, mediaID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Signal, error)
	GetByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*Signal, error)
	GetByMediaIDAndName(ctx context.Context, mediaID uuid.UUID, name string) (*Signal, error)
}

type SignalReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*SignalView, error)
	FindByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*SignalView, error)
	FindByMediaIDs(ctx context.Context, mediaIDs []uuid.UUID) (map[uuid.UUID][]*SignalView, error)
}

type SignalView struct {
	ID         uuid.UUID
	MediaID    uuid.UUID
	Name       string
	Score      int
	Confidence ConfidenceLevel
	Details    []string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
