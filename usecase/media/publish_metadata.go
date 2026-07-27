package media

import (
	"context"
	"errors"
	"log"

	"go-api/domain/messaging"
	"go-api/domain/port"
	"go-api/domain/realtime"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type PublishMetadataUseCase struct {
	mediaRepo           repository.MediaRepository
	publisher           port.MessagePublisher
	centrifugoPublisher port.RealtimePublisher
	queues              port.AnalyzeQueues
}

func NewPublishMetadataUseCase(
	mediaRepo repository.MediaRepository,
	publisher port.MessagePublisher,
	centrifugoPublisher port.RealtimePublisher,
	queues port.AnalyzeQueues,
) *PublishMetadataUseCase {
	return &PublishMetadataUseCase{
		mediaRepo:           mediaRepo,
		publisher:           publisher,
		centrifugoPublisher: centrifugoPublisher,
		queues:              queues,
	}
}

func (u *PublishMetadataUseCase) Execute(ctx context.Context, mediaID uuid.UUID) error {
	media, err := u.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return errors.New("failed to get media")
	}

	event := messaging.AnalyzeMessage{
		UserID:       media.UserID,
		MediaID:      mediaID,
		MediaKey:     media.Key,
		ThumbnailKey: media.Thumbnail,
	}

	err = u.publisher.Publish(ctx, u.queues.Request, event)
	if err != nil {
		return errors.New("failed to publish metadata event")
	}

	realtimeEvent, err := realtime.NewScanStartedEvent(media)
	if err != nil {
		return errors.New("failed to build scan started event")
	}

	if err := u.centrifugoPublisher.PublishToUser(ctx, media.UserID, realtimeEvent); err != nil {
		log.Printf("publish metadata: failed to publish scan_started to centrifugo for media %s: %v", mediaID, err)
	}

	return nil
}
