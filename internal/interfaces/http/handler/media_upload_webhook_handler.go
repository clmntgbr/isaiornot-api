package handler

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	mediacmd "go-api/internal/application/command/media"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
)

type MediaUploadWebhookHandler struct {
	mediaBucket                 string
	processUploadedMediaHandler *mediacmd.ProcessUploadedMediaHandler
}

func NewMediaUploadWebhookHandler(
	mediaBucket string,
	processUploadedMediaHandler *mediacmd.ProcessUploadedMediaHandler,
) *MediaUploadWebhookHandler {
	return &MediaUploadWebhookHandler{
		mediaBucket:                 mediaBucket,
		processUploadedMediaHandler: processUploadedMediaHandler,
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
	if err := validation.Struct(event); err != nil {
		log.Printf("MinIO webhook: validation failed: %v", err)
		return validation.ValidationFailed(c, err)
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

		decodedKey, err := domainmedia.DecodeObjectKey(record.S3.Object.Key)
		if err != nil {
			log.Printf("MinIO webhook: invalid object key %q: %v", record.S3.Object.Key, err)
			continue
		}

		if strings.Contains(decodedKey, "/thumbnails/") {
			continue
		}

		if domainmedia.IsFrameObjectKey(decodedKey) {
			continue
		}

		userID, err := domainmedia.UserIDFromKey(record.S3.Object.Key)
		if err != nil {
			log.Printf("MinIO webhook: invalid object key %q: %v", record.S3.Object.Key, err)
			continue
		}

		fileKey, err := domainmedia.FileKeyFromObjectKey(record.S3.Object.Key)
		if err != nil {
			log.Printf("MinIO webhook: invalid object key %q: %v", record.S3.Object.Key, err)
			continue
		}

		err = h.processUploadedMediaHandler.Handle(ctx, mediacmd.ProcessUploadedMediaCommand{
			UserID:      userID,
			FileKey:     fileKey,
			ContentType: record.S3.Object.ContentType,
			Size:        record.S3.Object.Size,
		})
		if err != nil {
			log.Printf("MinIO webhook: failed to process upload for key %q: %v", fileKey, err)
			continue
		}
	}
}
