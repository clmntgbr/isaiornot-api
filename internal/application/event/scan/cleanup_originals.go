package scan

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"go-api/internal/application/messaging"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

type CleanupOriginalsHandler struct {
	mediaRepo domainmedia.MediaWriteRepository
	storage   port.Storage
}

func NewCleanupOriginalsHandler(
	mediaRepo domainmedia.MediaWriteRepository,
	storage port.Storage,
) *CleanupOriginalsHandler {
	return &CleanupOriginalsHandler{
		mediaRepo: mediaRepo,
		storage:   storage,
	}
}

func (h *CleanupOriginalsHandler) OnCompleted(ctx context.Context, payload []byte) error {
	var evt domainscan.ScanCompleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.cleanup(ctx, evt.ScanID, evt.UserID, evt.ID)
}

func (h *CleanupOriginalsHandler) OnFailed(ctx context.Context, payload []byte) error {
	var evt domainscan.ScanFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.cleanup(ctx, evt.ScanID, evt.UserID, evt.ID)
}

func (h *CleanupOriginalsHandler) cleanup(ctx context.Context, scanIDRaw, userIDRaw, eventID string) error {
	scanID, err := uuid.Parse(scanIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	medias, err := h.mediaRepo.GetByScanID(ctx, scanID)
	if err != nil {
		return messaging.Retryable(err)
	}

	for _, media := range medias {
		if media == nil || media.Key == "" {
			continue
		}
		if isThumbnailKey(media.Key) {
			continue
		}

		objectKey := domainmedia.NewObjectKey(userID, scanID, media.Key)
		if err := h.storage.Delete(ctx, objectKey); err != nil {
			log.Printf(
				"failed to delete original object eventId=%s scanId=%s key=%q: %v",
				eventID,
				scanIDRaw,
				objectKey,
				err,
			)
			return messaging.Retryable(err)
		}
	}

	log.Printf(
		"event handled cleanup_originals eventId=%s scanId=%s medias=%d",
		eventID,
		scanIDRaw,
		len(medias),
	)
	return nil
}

func isThumbnailKey(key string) bool {
	return strings.Contains(key, "/thumbnails/") || strings.HasPrefix(key, "thumbnails/")
}
