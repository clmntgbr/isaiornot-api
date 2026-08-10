package scan

import (
	"context"
	"errors"

	"go-api/internal/domain/event"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

type FailScanCommand struct {
	ScanID  uuid.UUID
	Message string
}

type FailScanHandler struct {
	scanRepo  domainscan.ScanWriteRepository
	mediaRepo domainmedia.MediaWriteRepository
	outbox    port.OutboxRepository
}

func NewFailScanHandler(
	scanRepo domainscan.ScanWriteRepository,
	mediaRepo domainmedia.MediaWriteRepository,
	outbox port.OutboxRepository,
) *FailScanHandler {
	return &FailScanHandler{
		scanRepo:  scanRepo,
		mediaRepo: mediaRepo,
		outbox:    outbox,
	}
}

func (h *FailScanHandler) Handle(ctx context.Context, cmd FailScanCommand) error {
	return h.scanRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		scanEntity, err := h.scanRepo.GetByID(txCtx, cmd.ScanID)
		if err != nil {
			return errors.New("failed to get scan")
		}
		if scanEntity == nil {
			return errors.New("scan not found")
		}

		medias, err := h.mediaRepo.GetByScanID(txCtx, scanEntity.ID)
		if err != nil {
			return errors.New("failed to get medias")
		}

		events := make([]event.DomainEvent, 0)
		for _, mediaEntity := range medias {
			if mediaEntity.Status == domainmedia.StatusFailed {
				continue
			}
			mediaEntity.MarkFailed()
			if err := h.mediaRepo.Update(txCtx, mediaEntity); err != nil {
				return errors.New("failed to update media status")
			}
			events = append(events, mediaEntity.PullEvents()...)
		}

		if scanEntity.Status != domainscan.StatusFailed {
			scanEntity.MarkFailed(cmd.Message)
			if err := h.scanRepo.Update(txCtx, scanEntity); err != nil {
				return errors.New("failed to update scan")
			}
			events = append(events, scanEntity.PullEvents()...)
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}
