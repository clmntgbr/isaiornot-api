package realtime

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const EventQuotaUpdated = "quota_updated"

type QuotaEvent struct {
	Type      string    `json:"type"`
	UserID    string    `json:"userId"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewQuotaUpdatedEvent(userID uuid.UUID) QuotaEvent {
	return QuotaEvent{
		Type:      EventQuotaUpdated,
		UserID:    userID.String(),
		UpdatedAt: time.Now().UTC(),
	}
}

func (e QuotaEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
