package media

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime port.RealtimePublisher
	scanRepo domainscan.ScanWriteRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	scanRepo domainscan.ScanWriteRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		realtime: realtimePublisher,
		scanRepo: scanRepo,
	}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForScan(ctx, evt.ScanID, realtime.ActionCreated, evt)
}

func (h *PublishRealtimeHandler) OnUpdated(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForScan(ctx, evt.ScanID, realtime.ActionUpdated, evt)
}

func (h *PublishRealtimeHandler) OnUploaded(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaUploaded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForScan(ctx, evt.ScanID, realtime.ActionUpdated, evt)
}

func (h *PublishRealtimeHandler) OnCompleted(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaCompleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForScan(ctx, evt.ScanID, realtime.ActionUpdated, evt)
}

func (h *PublishRealtimeHandler) OnFailed(ctx context.Context, payload []byte) error {
	var evt domainmedia.MediaFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForScan(ctx, evt.ScanID, realtime.ActionUpdated, evt)
}

func (h *PublishRealtimeHandler) publishForScan(ctx context.Context, scanIDRaw, action string, payload any) error {
	scanID, err := uuid.Parse(scanIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	scanEntity, err := h.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if scanEntity == nil {
		return messaging.NonRetryable(errScanNotFound)
	}

	eventType := realtime.EventType(realtime.EntityMedia, action)
	if err := h.realtime.PublishToUser(ctx, scanEntity.UserID, eventType, payload); err != nil {
		log.Printf("centrifugo publish failed type=%s scanId=%s: %v", eventType, scanIDRaw, err)
		return messaging.Retryable(err)
	}
	log.Printf("centrifugo published type=%s scanId=%s userId=%s", eventType, scanIDRaw, scanEntity.UserID)
	return nil
}

type scanNotFoundError struct{}

func (scanNotFoundError) Error() string { return "scan not found for media realtime publish" }

var errScanNotFound error = scanNotFoundError{}
