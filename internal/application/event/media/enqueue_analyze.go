package media

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/application/messaging"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type EnqueueAnalyzeHandler struct {
	publisher port.EventPublisher
}

func NewEnqueueAnalyzeHandler(publisher port.EventPublisher) *EnqueueAnalyzeHandler {
	return &EnqueueAnalyzeHandler{publisher: publisher}
}

func (h *EnqueueAnalyzeHandler) OnMediaUploaded(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaUploaded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	if !domainmedia.IsImageContentType(domainmedia.ContentTypeFromKey(evt.Key, evt.ContentType)) {
		return nil
	}
	return h.enqueue(ctx, domainmedia.EventTypeMediaAnalyzeMetadata, domainmedia.StageMetadata, evt.MediaID, evt.ScanID)
}

func (h *EnqueueAnalyzeHandler) OnMetadataDone(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaAnalyzeMetadataDone
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.enqueue(ctx, domainmedia.EventTypeMediaAnalyzeHeuristics, domainmedia.StageHeuristics, evt.MediaID, evt.ScanID)
}

func (h *EnqueueAnalyzeHandler) OnHeuristicsDone(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaAnalyzeHeuristicsDone
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.enqueue(ctx, domainmedia.EventTypeMediaAnalyzeAIModel, domainmedia.StageAIModel, evt.MediaID, evt.ScanID)
}

func (h *EnqueueAnalyzeHandler) enqueue(
	ctx context.Context,
	eventType string,
	stage string,
	mediaID string,
	scanID string,
) error {
	if _, err := uuid.Parse(mediaID); err != nil {
		return messaging.NonRetryable(err)
	}

	now := time.Now().UTC()
	eventID := uuid.New().String()
	body, err := json.Marshal(domainmedia.MediaAnalyzeRequested{
		ID:        eventID,
		MediaID:   mediaID,
		ScanID:    scanID,
		Stage:     stage,
		Timestamp: now,
	})
	if err != nil {
		return messaging.Retryable(err)
	}

	if err := h.publisher.Publish(ctx, port.EventEnvelope{
		EventID:     eventID,
		Type:        eventType,
		AggregateID: mediaID,
		OccurredAt:  now.Format(time.RFC3339Nano),
		Payload:     body,
	}); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}
