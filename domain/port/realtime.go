package port

import (
	"context"

	"go-api/domain/realtime"

	"github.com/google/uuid"
)

type RealtimePublisher interface {
	PublishToUser(ctx context.Context, userID uuid.UUID, event realtime.MediaEvent) error
	PublishSubscriptionToUser(ctx context.Context, userID uuid.UUID, event realtime.SubscriptionEvent) error
	PublishUserToUser(ctx context.Context, userID uuid.UUID, event realtime.UserEvent) error
	PublishQuotaToUser(ctx context.Context, userID uuid.UUID, event realtime.QuotaEvent) error
}
