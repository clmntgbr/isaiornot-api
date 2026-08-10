package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

var (
	ErrCompleteMediaNotFound = errors.New("media not found")
)

type CompleteMediaCommand struct {
	MediaID uuid.UUID
}

type CompleteMediaHandler struct {
	mediaRepo domainmedia.MediaWriteRepository
	outbox    port.OutboxRepository
	publisher port.EventPublisher
}

func NewCompleteMediaHandler(
	mediaRepo domainmedia.MediaWriteRepository,
	outbox port.OutboxRepository,
	publisher port.EventPublisher,
) *CompleteMediaHandler {
	return &CompleteMediaHandler{
		mediaRepo: mediaRepo,
		outbox:    outbox,
		publisher: publisher,
	}
}

func (h *CompleteMediaHandler) Handle(ctx context.Context, cmd CompleteMediaCommand) error {
	mediaEntity, err := h.mediaRepo.GetByID(ctx, cmd.MediaID)
	if err != nil {
		return err
	}
	if mediaEntity == nil {
		return ErrCompleteMediaNotFound
	}

	if mediaEntity.Status != domainmedia.StatusCompleted {
		err = h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
			mediaEntity.MarkCompleted()
			if err := h.mediaRepo.Update(txCtx, mediaEntity); err != nil {
				return err
			}
			return h.outbox.StoreEvents(txCtx, mediaEntity.PullEvents())
		})
		if err != nil {
			return fmt.Errorf("failed to complete media: %w", err)
		}
	}

	medias, err := h.mediaRepo.GetByScanID(ctx, mediaEntity.ScanID)
	if err != nil {
		return err
	}
	for _, media := range medias {
		if media.Status != domainmedia.StatusCompleted {
			return nil
		}
	}

	return h.publishScanFinalize(ctx, mediaEntity.ScanID)
}

func (h *CompleteMediaHandler) publishScanFinalize(ctx context.Context, scanID uuid.UUID) error {
	if h.publisher == nil {
		return nil
	}

	now := time.Now().UTC()
	eventID := uuid.New().String()
	payload, err := json.Marshal(domainscan.ScanFinalizeRequested{
		ID:        eventID,
		ScanID:    scanID.String(),
		Timestamp: now,
	})
	if err != nil {
		return err
	}

	return h.publisher.Publish(ctx, port.EventEnvelope{
		EventID:     eventID,
		Type:        domainscan.EventTypeScanFinalize,
		AggregateID: scanID.String(),
		OccurredAt:  now.Format(time.RFC3339Nano),
		Payload:     payload,
	})
}
