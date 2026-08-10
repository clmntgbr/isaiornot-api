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

type AnalyzeAIModelOnRequestedHandler struct {
	analyzeAIModelHandler *mediacmd.AnalyzeAIModelHandler
}

func NewAnalyzeAIModelOnRequestedHandler(
	analyzeAIModelHandler *mediacmd.AnalyzeAIModelHandler,
) *AnalyzeAIModelOnRequestedHandler {
	return &AnalyzeAIModelOnRequestedHandler{
		analyzeAIModelHandler: analyzeAIModelHandler,
	}
}

func (h *AnalyzeAIModelOnRequestedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaAnalyzeRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	mediaID, err := uuid.Parse(evt.MediaID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	err = h.analyzeAIModelHandler.Handle(ctx, mediacmd.AnalyzeAIModelCommand{MediaID: mediaID})
	if err != nil {
		if errors.Is(err, mediacmd.ErrAIModelMediaNotFound) || errors.Is(err, mediacmd.ErrAIModelScanNotFound) {
			return messaging.NonRetryable(err)
		}
		log.Printf(
			"ai_model analysis failed eventId=%s mediaId=%s: %v",
			evt.ID,
			evt.MediaID,
			err,
		)
		return messaging.Retryable(err)
	}

	log.Printf(
		"event handled %s eventId=%s mediaId=%s action=analyze_ai_model",
		domainmedia.EventTypeMediaAnalyzeAIModel,
		evt.ID,
		evt.MediaID,
	)
	return nil
}
