package scan

import (
	"context"
	"time"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type ScanWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, scan *Scan) error
	Update(ctx context.Context, scan *Scan) error
	GetByID(ctx context.Context, id uuid.UUID) (*Scan, error)
}

type ScanReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*ScanView, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery) ([]*ScanView, int64, error)
}

type ScanView struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Status     Status
	Statuses   []Status
	Message    string
	FinalScore float64
	Confidence ConfidenceLevel
	Verdict    string
	Duration   int
	RetryCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
