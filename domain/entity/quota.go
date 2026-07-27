package entity

import (
	"time"

	"github.com/google/uuid"
)

type Quota struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	MaxImagesPerMonth int           `json:"maxImagesPerMonth"`
	MaxVideosPerMonth int           `json:"maxVideosPerMonth"`
	MaxFileSizeImage  int64         `json:"maxFileSizeImage"`
	MaxFileSizeVideo  int64         `json:"maxFileSizeVideo"`
	FullPipeline      bool          `json:"fullPipeline"`
	HistoryRetention  time.Duration `json:"historyRetention"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Quota) TableName() string {
	return "quotas"
}
