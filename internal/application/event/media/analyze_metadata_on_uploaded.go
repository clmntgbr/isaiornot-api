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

type AnalyzeMetadataOnUploadedHandler struct {
	analyzeMetadataHandler *mediacmd.AnalyzeMetadataHandler
}

func NewAnalyzeMetadataOnUploadedHandler(
	analyzeMetadataHandler *mediacmd.AnalyzeMetadataHandler,
) *AnalyzeMetadataOnUploadedHandler {
	return &AnalyzeMetadataOnUploadedHandler{
		analyzeMetadataHandler: analyzeMetadataHandler,
	}
}

func (h *AnalyzeMetadataOnUploadedHandler) Handle(ctx context.Context, payload []byte) error {
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

	err = h.analyzeMetadataHandler.Handle(ctx, mediacmd.AnalyzeMetadataCommand{MediaID: mediaID})
	if err != nil {
		if errors.Is(err, mediacmd.ErrAnalyzeMediaNotFound) || errors.Is(err, mediacmd.ErrAnalyzeScanNotFound) {
			return messaging.NonRetryable(err)
		}
		log.Printf(
			"metadata analysis failed eventId=%s mediaId=%s: %v",
			evt.ID,
			evt.MediaID,
			err,
		)
		return messaging.Retryable(err)
	}

	log.Printf(
		"event handled %s eventId=%s mediaId=%s action=analyze_metadata",
		domainmedia.EventTypeMediaUploaded,
		evt.ID,
		evt.MediaID,
	)
	return nil
}
