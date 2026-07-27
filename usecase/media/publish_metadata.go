package media

import (
	"context"
	"errors"
	"go-api/domain/repository"
	"go-api/infrastructure/centrifugo"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"log"

	"github.com/google/uuid"
)

type PublishMetadataUseCase struct {
	mediaRepo           *repository.MediaRepository
	publisher           rabbitmq.Publisher
	centrifugoPublisher *centrifugo.Publisher
	config              *config.Config
}

func NewPublishMetadataUseCase(
	mediaRepo *repository.MediaRepository,
	publisher rabbitmq.Publisher,
	centrifugoPublisher *centrifugo.Publisher,
	config *config.Config,
) *PublishMetadataUseCase {
	return &PublishMetadataUseCase{
		mediaRepo:           mediaRepo,
		publisher:           publisher,
		centrifugoPublisher: centrifugoPublisher,
		config:              config,
	}
}

func (u *PublishMetadataUseCase) Execute(ctx context.Context, mediaID uuid.UUID) error {
	media, err := (*u.mediaRepo).GetByID(ctx, mediaID)
	if err != nil {
		return errors.New("failed to get media")
	}

	event := rabbitmq.AnalyzeMessage{
		UserID:       media.UserID,
		MediaID:      mediaID,
		MediaKey:     media.Key,
		ThumbnailKey: media.Thumbnail,
	}

	err = u.publisher.Publish(ctx, u.config.AnalyzeRequestQueueName, event)
	if err != nil {
		return errors.New("failed to publish metadata event")
	}

	realtimeEvent, err := centrifugo.NewScanStartedEvent(media)
	if err != nil {
		return errors.New("failed to build scan started event")
	}

	if err := u.centrifugoPublisher.PublishToUser(ctx, media.UserID, realtimeEvent); err != nil {
		log.Printf("publish metadata: failed to publish scan_started to centrifugo for media %s: %v", mediaID, err)
	}

	return nil
}
