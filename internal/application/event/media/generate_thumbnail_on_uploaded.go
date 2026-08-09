package media

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	mediacmd "go-api/internal/application/command/media"
	"go-api/internal/application/messaging"
	domainmedia "go-api/internal/domain/media"

	"github.com/google/uuid"
)

type GenerateThumbnailOnUploadedHandler struct {
	generateThumbnailHandler *mediacmd.GenerateThumbnailHandler
}

func NewGenerateThumbnailOnUploadedHandler(
	generateThumbnailHandler *mediacmd.GenerateThumbnailHandler,
) *GenerateThumbnailOnUploadedHandler {
	return &GenerateThumbnailOnUploadedHandler{
		generateThumbnailHandler: generateThumbnailHandler,
	}
}

func (h *GenerateThumbnailOnUploadedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaUploaded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	if !domainmedia.IsImageContentType(domainmedia.ContentTypeFromKey(evt.Key, evt.ContentType)) {
		return nil
	}

	mediaID, err := uuid.Parse(evt.MediaID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	err = h.generateThumbnailHandler.Handle(ctx, mediacmd.GenerateThumbnailCommand{MediaID: mediaID})
	if err != nil {
		if errors.Is(err, mediacmd.ErrUnsupportedContentType) {
			return nil
		}
		if errors.Is(err, mediacmd.ErrThumbnailMediaNotFound) || errors.Is(err, mediacmd.ErrThumbnailScanNotFound) {
			return messaging.NonRetryable(err)
		}
		log.Printf(
			"thumbnail generation failed eventId=%s mediaId=%s: %v",
			evt.ID,
			evt.MediaID,
			err,
		)
		return messaging.Retryable(err)
	}

	log.Printf(
		"event handled %s eventId=%s mediaId=%s action=generate_thumbnail",
		domainmedia.EventTypeMediaUploaded,
		evt.ID,
		evt.MediaID,
	)
	return nil
}
