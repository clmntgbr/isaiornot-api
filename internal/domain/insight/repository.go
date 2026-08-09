package insight

import (
	"context"

	"github.com/google/uuid"
)

type InsightWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, insight *Insight) error
	Update(ctx context.Context, insight *Insight) error
	GetByID(ctx context.Context, id uuid.UUID) (*Insight, error)
}
