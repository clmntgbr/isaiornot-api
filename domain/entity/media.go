package entity

import (
	"go-api/domain/enum"
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	ScanID uuid.UUID `gorm:"type:uuid;index:idx_media_scan_id" json:"scanId"`
	Scan   Scan      `gorm:"foreignKey:ScanID" json:"-"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`

	InsightID *uuid.UUID `gorm:"type:uuid;default:null;index:idx_media_insight_id" json:"insightId"`
	Insight   *Insight   `gorm:"foreignKey:InsightID" json:"insight"`

	Key         string `gorm:"uniqueIndex;not null" json:"key"`
	Filename    string `gorm:"not null;default:''" json:"filename"`
	Thumbnail   string `gorm:"not null" json:"thumbnail"`
	ContentType string `gorm:"not null" json:"contentType"`
	Size        int64  `gorm:"not null" json:"size"`

	Signals []Signal `gorm:"foreignKey:MediaID" json:"signals"`

	Status   enum.MediaStatus   `gorm:"type:varchar(20);not null;check:status IN ('processing','uploaded','analyzed','failed');index:idx_media_status" json:"status"`
	Statuses []enum.MediaStatus `gorm:"serializer:json;type:jsonb;default:'[]'" json:"statuses"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Media) TableName() string {
	return "medias"
}
