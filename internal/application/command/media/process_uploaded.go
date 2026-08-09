package media

import (
	"context"
	"errors"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

var (
	ErrMediaNotFound = errors.New("media not found")
	ErrScanNotFound  = errors.New("scan not found")
	ErrForbidden     = errors.New("media does not belong to user")
)

type ProcessUploadedMediaCommand struct {
	UserID      uuid.UUID
	FileKey     string
	ContentType string
	Size        int64
}

type ProcessUploadedMediaHandler struct {
	scanRepo  domainscan.ScanWriteRepository
	mediaRepo domainmedia.MediaWriteRepository
	outbox    port.OutboxRepository
}

func NewProcessUploadedMediaHandler(
	scanRepo domainscan.ScanWriteRepository,
	mediaRepo domainmedia.MediaWriteRepository,
	outbox port.OutboxRepository,
) *ProcessUploadedMediaHandler {
	return &ProcessUploadedMediaHandler{
		scanRepo:  scanRepo,
		mediaRepo: mediaRepo,
		outbox:    outbox,
	}
}

func (h *ProcessUploadedMediaHandler) Handle(ctx context.Context, cmd ProcessUploadedMediaCommand) error {
	contentType := domainmedia.ContentTypeFromKey(cmd.FileKey, cmd.ContentType)

	mediaEntity, err := h.mediaRepo.GetByKey(ctx, cmd.FileKey)
	if err != nil {
		return err
	}
	if mediaEntity == nil {
		return ErrMediaNotFound
	}

	scanEntity, err := h.scanRepo.GetByID(ctx, mediaEntity.ScanID)
	if err != nil {
		return err
	}
	if scanEntity == nil {
		return ErrScanNotFound
	}
	if scanEntity.UserID != cmd.UserID {
		return ErrForbidden
	}

	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		mediaEntity.RecordUpload(contentType, cmd.Size)
		if err := h.mediaRepo.Update(txCtx, mediaEntity); err != nil {
			return err
		}

		scanEntity.MarkProcessing()
		if err := h.scanRepo.Update(txCtx, scanEntity); err != nil {
			return err
		}

		events := append(mediaEntity.PullEvents(), scanEntity.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
}
