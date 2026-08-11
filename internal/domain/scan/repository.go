package scan

import (
	"context"
	"time"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type ScanWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, scan *Scan) error
	Update(ctx context.Context, scan *Scan) error
	GetByID(ctx context.Context, id uuid.UUID) (*Scan, error)
	ListInProgressCreatedBefore(ctx context.Context, before time.Time) ([]*Scan, error)
}

type ScanReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*ScanView, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery, since time.Time) ([]*ScanView, int64, error)
	GetStatisticsByUserID(ctx context.Context, userID uuid.UUID, since time.Time) (*StatisticsView, error)
	CountUsageInPeriod(ctx context.Context, userID uuid.UUID, from, to time.Time) (*UsageCounts, error)
}

type UsageCounts struct {
	Images int64
	Videos int64
}

type ScanView struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Type       ScanType
	Status     Status
	Statuses   []Status
	Message    string
	FinalScore float64
	Confidence ConfidenceLevel
	Verdict    string
	Duration   int
	RetryCount int
	Medias     []*domainmedia.MediaView
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type StatisticsView struct {
	ScansCount     int64
	RealImageCount int64
	AIImageCount   int64
	AverageScore   float64
}
