package write

import (
	"time"

	"github.com/google/uuid"
)

type InsightModel struct {
	ID          uuid.UUID `gorm:"column:id;primaryKey"`
	Noise       float64   `gorm:"column:noise"`
	Compression float64   `gorm:"column:compression"`
	Frequency   float64   `gorm:"column:frequency"`
	Histogram   float64   `gorm:"column:histogram"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (InsightModel) TableName() string {
	return "insights"
}
