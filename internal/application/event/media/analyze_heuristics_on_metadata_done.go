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

type AnalyzeHeuristicsOnMetadataDoneHandler struct {
	analyzeHeuristicsHandler *mediacmd.AnalyzeHeuristicsHandler
}

func NewAnalyzeHeuristicsOnMetadataDoneHandler(
	analyzeHeuristicsHandler *mediacmd.AnalyzeHeuristicsHandler,
) *AnalyzeHeuristicsOnMetadataDoneHandler {
	return &AnalyzeHeuristicsOnMetadataDoneHandler{
		analyzeHeuristicsHandler: analyzeHeuristicsHandler,
	}
}

func (h *AnalyzeHeuristicsOnMetadataDoneHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaAnalyzeMetadataDone
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	mediaID, err := uuid.Parse(evt.MediaID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	err = h.analyzeHeuristicsHandler.Handle(ctx, mediacmd.AnalyzeHeuristicsCommand{MediaID: mediaID})
	if err != nil {
		if errors.Is(err, mediacmd.ErrHeuristicsMediaNotFound) || errors.Is(err, mediacmd.ErrHeuristicsScanNotFound) {
			return messaging.NonRetryable(err)
		}
		log.Printf(
			"heuristics analysis failed eventId=%s mediaId=%s: %v",
			evt.ID,
			evt.MediaID,
			err,
		)
		return messaging.Retryable(err)
	}

	log.Printf(
		"event handled %s eventId=%s mediaId=%s action=analyze_heuristics",
		domainmedia.EventTypeMediaAnalyzeMetadataDone,
		evt.ID,
		evt.MediaID,
	)
	return nil
}
