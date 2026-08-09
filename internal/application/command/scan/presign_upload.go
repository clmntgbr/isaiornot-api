package scan

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

var ErrUnsupportedMediaType = errors.New("unsupported media type")

type PresignUploadCommand struct {
	UserID      uuid.UUID
	Filename    string
	ContentType string
}

type PresignUploadResult struct {
	URL    string
	ScanID uuid.UUID
}

type PresignUploadHandler struct {
	scanRepo  domainscan.ScanWriteRepository
	mediaRepo domainmedia.MediaWriteRepository
	outbox    port.OutboxRepository
	storage   port.Storage
}

func NewPresignUploadHandler(
	scanRepo domainscan.ScanWriteRepository,
	mediaRepo domainmedia.MediaWriteRepository,
	outbox port.OutboxRepository,
	storage port.Storage,
) *PresignUploadHandler {
	return &PresignUploadHandler{
		scanRepo:  scanRepo,
		mediaRepo: mediaRepo,
		outbox:    outbox,
		storage:   storage,
	}
}

func (h *PresignUploadHandler) Handle(ctx context.Context, cmd PresignUploadCommand) (*PresignUploadResult, error) {
	input := domainmedia.PresignUploadInput{
		Filename:    cmd.Filename,
		ContentType: cmd.ContentType,
	}
	if err := domainmedia.ValidatePresignUploadInput(input); err != nil {
		return nil, errors.Join(ErrUnsupportedMediaType, err)
	}

	contentType := input.ContentType
	if contentType == "" {
		contentType = domainmedia.ContentTypeFromKey(input.Filename, "")
	}

	filename := filepath.Base(input.Filename)
	fileKey := domainmedia.NewFileKey(input.Filename)

	scanEntity := domainscan.NewScan(cmd.UserID)
	mediaEntity := domainmedia.NewMedia(scanEntity.ID, fileKey, filename, contentType, 0)
	objectKey := domainmedia.NewObjectKey(cmd.UserID, scanEntity.ID, fileKey)

	err := h.scanRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.scanRepo.Save(txCtx, scanEntity); err != nil {
			return err
		}
		if err := h.mediaRepo.Save(txCtx, mediaEntity); err != nil {
			return err
		}

		events := append(scanEntity.PullEvents(), mediaEntity.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		return nil, errors.New("failed to create scan upload")
	}

	url, err := h.storage.PresignedPutURL(ctx, objectKey, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	return &PresignUploadResult{
		URL:    url,
		ScanID: scanEntity.ID,
	}, nil
}
