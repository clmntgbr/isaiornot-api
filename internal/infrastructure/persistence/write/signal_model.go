package write

import (
	"time"

	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

// SignalModel is the persistence mapping for table signals.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type SignalModel struct {
	ID         uuid.UUID    `gorm:"column:id;primaryKey"`
	MediaID    uuid.UUID    `gorm:"column:media_id"`
	Name       string       `gorm:"column:name"`
	Score      int          `gorm:"column:score"`
	Confidence string       `gorm:"column:confidence"`
	Details    dbtype.JSONB `gorm:"column:details"`
	CreatedAt  time.Time    `gorm:"column:created_at"`
	UpdatedAt  time.Time    `gorm:"column:updated_at"`
}

func (SignalModel) TableName() string {
	return "signals"
}
