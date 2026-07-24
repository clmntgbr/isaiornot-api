package entity

import (
	"go-api/domain/enum"
	"time"

	"github.com/google/uuid"
)

type Analysis struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_analysis_user_id" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`

	Medias []Media `gorm:"foreignKey:AnalysisID" json:"medias"`

	Status   enum.AnalysisStatus   `gorm:"type:varchar(20);not null;default:'processing';check:status IN ('processing','uploaded','analyzed');index:idx_analysis_status" json:"status"`
	Statuses []enum.AnalysisStatus `gorm:"serializer:json;type:jsonb;default:'[]'" json:"statuses"`

	FinalScore         float64         `gorm:"default:-1" json:"final_score"`
	AnalysisConfidence ConfidenceLevel `gorm:"type:varchar(20);default:'unknown'" json:"confidence"`
	Verdict            string          `gorm:"type:varchar(20);default:''" json:"verdict"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Analysis) TableName() string {
	return "analyses"
}
