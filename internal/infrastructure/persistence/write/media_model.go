package write

import (
	"time"

	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

// MediaModel is the persistence mapping for table medias.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type MediaModel struct {
	ID          uuid.UUID    `gorm:"column:id;primaryKey"`
	ScanID      uuid.UUID    `gorm:"column:scan_id"`
	Key         string       `gorm:"column:key"`
	Filename    string       `gorm:"column:filename"`
	Thumbnail   string       `gorm:"column:thumbnail"`
	ContentType string       `gorm:"column:content_type"`
	Size        int64        `gorm:"column:size"`
	InsightID   *uuid.UUID   `gorm:"column:insight_id"`
	Status      string       `gorm:"column:status"`
	Statuses    dbtype.JSONB `gorm:"column:statuses"`
	CreatedAt   time.Time    `gorm:"column:created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at"`
}

func (MediaModel) TableName() string {
	return "medias"
}
