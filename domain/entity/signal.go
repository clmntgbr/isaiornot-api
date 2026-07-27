package entity

import (
	"time"

	"github.com/google/uuid"
)

type ConfidenceLevel string

const (
	ConfidenceHigh    ConfidenceLevel = "high"
	ConfidenceMedium  ConfidenceLevel = "medium"
	ConfidenceLow     ConfidenceLevel = "low"
	ConfidenceUnknown ConfidenceLevel = "unknown"
)

type Signal struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MediaID    uuid.UUID       `gorm:"type:uuid;not null" json:"mediaId"`
	Media      Media           `gorm:"foreignKey:MediaID" json:"media"`
	Name       string          `json:"name"`
	Score      int             `json:"score"`
	Confidence ConfidenceLevel `json:"confidence"`
	Details    []string        `json:"details" gorm:"serializer:json;type:jsonb;default:'[]'"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Signal) TableName() string {
	return "signals"
}
