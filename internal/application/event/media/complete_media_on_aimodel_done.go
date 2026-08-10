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

type CompleteMediaOnAIModelDoneHandler struct {
	completeMediaHandler *mediacmd.CompleteMediaHandler
}

func NewCompleteMediaOnAIModelDoneHandler(
	completeMediaHandler *mediacmd.CompleteMediaHandler,
) *CompleteMediaOnAIModelDoneHandler {
	return &CompleteMediaOnAIModelDoneHandler{
		completeMediaHandler: completeMediaHandler,
	}
}

func (h *CompleteMediaOnAIModelDoneHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaAnalyzeAIModelDone
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	mediaID, err := uuid.Parse(evt.MediaID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	err = h.completeMediaHandler.Handle(ctx, mediacmd.CompleteMediaCommand{MediaID: mediaID})
	if err != nil {
		if errors.Is(err, mediacmd.ErrCompleteMediaNotFound) {
			return messaging.NonRetryable(err)
		}
		log.Printf(
			"complete media failed eventId=%s mediaId=%s: %v",
			evt.ID,
			evt.MediaID,
			err,
		)
		return messaging.Retryable(err)
	}

	log.Printf(
		"event handled %s eventId=%s mediaId=%s action=complete_media",
		domainmedia.EventTypeMediaAnalyzeAIModelDone,
		evt.ID,
		evt.MediaID,
	)
	return nil
}
