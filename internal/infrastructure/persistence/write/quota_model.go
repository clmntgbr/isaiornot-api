package write

import (
	"time"

	"github.com/google/uuid"
)

type QuotaModel struct {
	ID                uuid.UUID `gorm:"column:id;primaryKey"`
	MaxImagesPerMonth int       `gorm:"column:max_images_per_month"`
	MaxVideosPerMonth int       `gorm:"column:max_videos_per_month"`
	MaxFileSizeImage  int64     `gorm:"column:max_file_size_image"`
	MaxFileSizeVideo  int64     `gorm:"column:max_file_size_video"`
	FullPipeline      bool      `gorm:"column:full_pipeline"`
	HistoryRetention  int64     `gorm:"column:history_retention"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (QuotaModel) TableName() string {
	return "quotas"
}
