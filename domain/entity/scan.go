package entity

import (
	"go-api/domain/enum"
	"time"

	"github.com/google/uuid"
)

type Scan struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_scan_user_id" json:"userId"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`

	Medias []Media `gorm:"foreignKey:ScanID" json:"medias"`

	Status   enum.ScanStatus   `gorm:"type:varchar(20);not null;default:'uploaded';check:status IN ('uploaded','processing','completed','failed');index:idx_scan_status" json:"status"`
	Statuses []enum.ScanStatus `gorm:"serializer:json;type:jsonb;default:'[]'" json:"statuses"`
	Message  string            `gorm:"type:text;not null;default:''" json:"message"`

	FinalScore float64         `gorm:"default:-1" json:"finalScore"`
	Confidence ConfidenceLevel `gorm:"type:varchar(20);default:'unknown'" json:"confidence"`
	Verdict    string          `gorm:"type:varchar(20);default:''" json:"verdict"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Scan) TableName() string {
	return "scans"
}
