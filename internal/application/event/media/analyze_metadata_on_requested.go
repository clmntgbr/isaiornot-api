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

type AnalyzeMetadataOnRequestedHandler struct {
	analyzeMetadataHandler *mediacmd.AnalyzeMetadataHandler
}

func NewAnalyzeMetadataOnRequestedHandler(
	analyzeMetadataHandler *mediacmd.AnalyzeMetadataHandler,
) *AnalyzeMetadataOnRequestedHandler {
	return &AnalyzeMetadataOnRequestedHandler{
		analyzeMetadataHandler: analyzeMetadataHandler,
	}
}

func (h *AnalyzeMetadataOnRequestedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaAnalyzeRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
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
		domainmedia.EventTypeMediaAnalyzeMetadata,
		evt.ID,
		evt.MediaID,
	)
	return nil
}
