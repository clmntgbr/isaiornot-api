package messaging

import "github.com/google/uuid"

type AnalyzeMessage struct {
	UserID       uuid.UUID `json:"user_id"`
	MediaID      uuid.UUID `json:"media_id"`
	MediaKey     string    `json:"media_key"`
	ThumbnailKey string    `json:"thumbnail_key"`
}

type StageDoneMessage struct {
	MediaID uuid.UUID `json:"media_id"`
	Stage   string    `json:"stage"`
}

type FailedMessage struct {
	MediaID uuid.UUID `json:"media_id"`
	Stage   string    `json:"stage"`
	Error   string    `json:"error"`
}
