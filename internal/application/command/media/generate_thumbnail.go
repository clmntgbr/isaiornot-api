package media

import (
	"bytes"
	"context"
	"errors"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

var (
	ErrUnsupportedContentType = errors.New("unsupported content type for thumbnail")
	ErrThumbnailMediaNotFound = errors.New("media not found")
	ErrThumbnailScanNotFound  = errors.New("scan not found")
)

type GenerateThumbnailCommand struct {
	MediaID uuid.UUID
}

type GenerateThumbnailHandler struct {
	mediaRepo   domainmedia.MediaWriteRepository
	scanRepo    domainscan.ScanWriteRepository
	outbox      port.OutboxRepository
	storage     port.Storage
	thumbnailer port.ImageThumbnailer
}

func NewGenerateThumbnailHandler(
	mediaRepo domainmedia.MediaWriteRepository,
	scanRepo domainscan.ScanWriteRepository,
	outbox port.OutboxRepository,
	storage port.Storage,
	thumbnailer port.ImageThumbnailer,
) *GenerateThumbnailHandler {
	return &GenerateThumbnailHandler{
		mediaRepo:   mediaRepo,
		scanRepo:    scanRepo,
		outbox:      outbox,
		storage:     storage,
		thumbnailer: thumbnailer,
	}
}

func (h *GenerateThumbnailHandler) Handle(ctx context.Context, cmd GenerateThumbnailCommand) error {
	mediaEntity, err := h.mediaRepo.GetByID(ctx, cmd.MediaID)
	if err != nil {
		return err
	}
	if mediaEntity == nil {
		return ErrThumbnailMediaNotFound
	}
	if mediaEntity.Thumbnail != "" {
		return nil
	}

	contentType := domainmedia.ContentTypeFromKey(mediaEntity.Key, mediaEntity.ContentType)
	if !domainmedia.IsImageContentType(contentType) {
		return ErrUnsupportedContentType
	}

	scanEntity, err := h.scanRepo.GetByID(ctx, mediaEntity.ScanID)
	if err != nil {
		return err
	}
	if scanEntity == nil {
		return ErrThumbnailScanNotFound
	}

	original, err := h.storage.Get(ctx, domainmedia.NewObjectKey(scanEntity.UserID, mediaEntity.ScanID, mediaEntity.Key))
	if err != nil {
		return errors.New("failed to fetch original")
	}
	defer original.Close()

	thumbBytes, err := h.thumbnailer.GenerateJPEG(ctx, original, 400)
	if err != nil {
		return err
	}

	thumbKey := domainmedia.NewThumbnailFileKey(mediaEntity.ID)
	if err := h.storage.PutThumbnail(
		ctx,
		domainmedia.NewThumbnailObjectKey(scanEntity.UserID, mediaEntity.ScanID, mediaEntity.ID),
		bytes.NewReader(thumbBytes),
		int64(len(thumbBytes)),
		"image/jpeg",
	); err != nil {
		return errors.New("failed to store thumbnail")
	}

	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		fresh, err := h.mediaRepo.GetByID(txCtx, mediaEntity.ID)
		if err != nil {
			return err
		}
		if fresh == nil {
			return ErrThumbnailMediaNotFound
		}
		if fresh.Thumbnail != "" {
			return nil
		}

		fresh.SetThumbnail(thumbKey)
		if err := h.mediaRepo.Update(txCtx, fresh); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, fresh.PullEvents())
	})
}
