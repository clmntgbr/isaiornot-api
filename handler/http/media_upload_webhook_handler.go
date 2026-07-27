package http

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	mediadto "go-api/domain/media"
	"go-api/handler/http/dto"
	"go-api/usecase/media"

	"github.com/gofiber/fiber/v3"
)

type MediaUploadWebhookHandler struct {
	mediaBucket                 string
	processUploadedMediaUseCase *media.ProcessUploadedMediaUseCase
}

func NewMediaUploadWebhookHandler(
	mediaBucket string,
	processUploadedMediaUseCase *media.ProcessUploadedMediaUseCase,
) *MediaUploadWebhookHandler {
	return &MediaUploadWebhookHandler{
		mediaBucket:                 mediaBucket,
		processUploadedMediaUseCase: processUploadedMediaUseCase,
	}
}

func (h *MediaUploadWebhookHandler) ObjectCreated(c fiber.Ctx) error {
	payload := c.Body()

	var event dto.ObjectCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Printf("MinIO webhook: invalid payload: %v body=%s", err, string(payload))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		h.process(ctx, event)
	}()

	return c.SendStatus(fiber.StatusOK)
}

func (h *MediaUploadWebhookHandler) process(ctx context.Context, event dto.ObjectCreatedEvent) {
	for _, record := range event.Records {
		if record.S3.Bucket.Name != h.mediaBucket {
			continue
		}

		decodedKey, err := mediadto.DecodeObjectKey(record.S3.Object.Key)
		if err != nil {
			log.Printf("MinIO webhook: invalid object key %q: %v", record.S3.Object.Key, err)
			continue
		}

		if strings.Contains(decodedKey, "/thumbnails/") {
			continue
		}

		if mediadto.IsFrameObjectKey(decodedKey) {
			continue
		}

		userID, err := mediadto.UserIDFromKey(record.S3.Object.Key)
		if err != nil {
			log.Printf("MinIO webhook: invalid object key %q: %v", record.S3.Object.Key, err)
			continue
		}

		fileKey, err := mediadto.FileKeyFromObjectKey(record.S3.Object.Key)
		if err != nil {
			log.Printf("MinIO webhook: invalid object key %q: %v", record.S3.Object.Key, err)
			continue
		}

		err = h.processUploadedMediaUseCase.Execute(
			ctx,
			userID,
			fileKey,
			record.S3.Object.ContentType,
			record.S3.Object.Size,
		)
		if err != nil {
			log.Printf("MinIO webhook: failed to process upload for key %q: %v", fileKey, err)
			continue
		}
	}
}
