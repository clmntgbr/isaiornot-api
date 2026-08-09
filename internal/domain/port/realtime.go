package port

import (
	"context"

	"github.com/google/uuid"
)

// RealtimePublisher publishes client-facing events to a user channel.
type RealtimePublisher interface {
	PublishToUser(ctx context.Context, userID uuid.UUID, eventType string, payload any) error
}
