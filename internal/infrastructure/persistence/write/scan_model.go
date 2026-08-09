package write

import (
	"time"

	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

// ScanModel is the persistence mapping for table scans.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type ScanModel struct {
	ID         uuid.UUID    `gorm:"column:id;primaryKey"`
	UserID     uuid.UUID    `gorm:"column:user_id"`
	Status     string       `gorm:"column:status"`
	Statuses   dbtype.JSONB `gorm:"column:statuses"`
	Message    string       `gorm:"column:message"`
	FinalScore float64      `gorm:"column:final_score"`
	Confidence string       `gorm:"column:confidence"`
	Verdict    string       `gorm:"column:verdict"`
	Duration   int          `gorm:"column:duration"`
	RetryCount int          `gorm:"column:retry_count"`
	CreatedAt  time.Time    `gorm:"column:created_at"`
	UpdatedAt  time.Time    `gorm:"column:updated_at"`
}

func (ScanModel) TableName() string {
	return "scans"
}
