package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	cmdsubscription "go-api/internal/application/command/subscription"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/domain/event"
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
	scanRepo            domainscan.ScanWriteRepository
	mediaRepo           domainmedia.MediaWriteRepository
	outbox              port.OutboxRepository
	storage             port.Storage
	frameExtractor      port.FrameExtractor
	thumbnailer         port.ImageThumbnailer
	generateThumbnail   *GenerateThumbnailHandler
	assertUploadAllowed *cmdsubscription.AssertUploadAllowedHandler
}

func NewProcessUploadedMediaHandler(
	scanRepo domainscan.ScanWriteRepository,
	mediaRepo domainmedia.MediaWriteRepository,
	outbox port.OutboxRepository,
	storage port.Storage,
	frameExtractor port.FrameExtractor,
	thumbnailer port.ImageThumbnailer,
	generateThumbnail *GenerateThumbnailHandler,
	assertUploadAllowed *cmdsubscription.AssertUploadAllowedHandler,
) *ProcessUploadedMediaHandler {
	return &ProcessUploadedMediaHandler{
		scanRepo:            scanRepo,
		mediaRepo:           mediaRepo,
		outbox:              outbox,
		storage:             storage,
		frameExtractor:      frameExtractor,
		thumbnailer:         thumbnailer,
		generateThumbnail:   generateThumbnail,
		assertUploadAllowed: assertUploadAllowed,
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

	if domainmedia.IsVideoContentType(contentType) {
		return h.processVideo(ctx, scanEntity, mediaEntity, contentType, cmd.Size)
	}

	return h.processImage(ctx, scanEntity, mediaEntity, contentType, cmd.Size)
}

func (h *ProcessUploadedMediaHandler) assertBeforePipeline(
	ctx context.Context,
	userID uuid.UUID,
	contentType string,
	size int64,
) (bool, string, error) {
	err := h.assertUploadAllowed.Handle(ctx, cmdsubscription.AssertUploadAllowedCommand{
		UserID:              userID,
		ContentType:         contentType,
		Size:                size,
		MediaAlreadyCounted: true,
	})
	if err == nil {
		return true, "", nil
	}

	message := err.Error()
	if errors.Is(err, querysubscription.ErrSubscriptionNotFound) {
		message = "no active subscription found"
	}
	return false, message, nil
}

func (h *ProcessUploadedMediaHandler) failUpload(
	ctx context.Context,
	scanEntity *domainscan.Scan,
	mediaEntity *domainmedia.Media,
	message string,
) error {
	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		freshMedia, err := h.mediaRepo.GetByID(txCtx, mediaEntity.ID)
		if err != nil {
			return err
		}
		if freshMedia == nil {
			return ErrMediaNotFound
		}
		freshMedia.MarkFailed()
		if err := h.mediaRepo.Update(txCtx, freshMedia); err != nil {
			return err
		}

		freshScan, err := h.scanRepo.GetByID(txCtx, scanEntity.ID)
		if err != nil {
			return err
		}
		if freshScan == nil {
			return ErrScanNotFound
		}
		freshScan.MarkFailed(message)
		if err := h.scanRepo.Update(txCtx, freshScan); err != nil {
			return err
		}

		events := append(freshMedia.PullEvents(), freshScan.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
}

func (h *ProcessUploadedMediaHandler) processImage(
	ctx context.Context,
	scanEntity *domainscan.Scan,
	mediaEntity *domainmedia.Media,
	contentType string,
	size int64,
) error {
	if err := h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		mediaEntity.ApplyContent(contentType, size)
		if err := h.mediaRepo.Update(txCtx, mediaEntity); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, mediaEntity.PullEvents())
	}); err != nil {
		return err
	}

	if err := h.generateThumbnail.Handle(ctx, GenerateThumbnailCommand{MediaID: mediaEntity.ID}); err != nil {
		if !errors.Is(err, ErrUnsupportedContentType) {
			return fmt.Errorf("failed to generate thumbnail: %w", err)
		}
	}

	allowed, message, err := h.assertBeforePipeline(ctx, scanEntity.UserID, contentType, size)
	if err != nil {
		return err
	}
	if !allowed {
		return h.failUpload(ctx, scanEntity, mediaEntity, message)
	}

	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		freshMedia, err := h.mediaRepo.GetByID(txCtx, mediaEntity.ID)
		if err != nil {
			return err
		}
		if freshMedia == nil {
			return ErrMediaNotFound
		}
		freshMedia.RecordUpload(contentType, size)
		if err := h.mediaRepo.Update(txCtx, freshMedia); err != nil {
			return err
		}

		freshScan, err := h.scanRepo.GetByID(txCtx, scanEntity.ID)
		if err != nil {
			return err
		}
		if freshScan == nil {
			return ErrScanNotFound
		}
		freshScan.MarkProcessing()
		if err := h.scanRepo.Update(txCtx, freshScan); err != nil {
			return err
		}

		events := append(freshMedia.PullEvents(), freshScan.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
}

func (h *ProcessUploadedMediaHandler) processVideo(
	ctx context.Context,
	scanEntity *domainscan.Scan,
	sourceMedia *domainmedia.Media,
	contentType string,
	size int64,
) error {
	objectKey := domainmedia.NewObjectKey(scanEntity.UserID, sourceMedia.ScanID, sourceMedia.Key)
	reader, err := h.storage.Get(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}
	defer reader.Close()

	videoData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read video: %w", err)
	}

	frames, err := h.frameExtractor.ExtractFrames(videoData, domainmedia.MaxVideoFrames())
	if err != nil {
		return fmt.Errorf("failed to extract frames: %w", err)
	}
	if len(frames) == 0 {
		return errors.New("no frames extracted from video")
	}

	baseFilename := sourceMedia.Filename
	scanID := sourceMedia.ScanID
	frameMedias := make([]*domainmedia.Media, 0, len(frames))

	for i, frame := range frames {
		frameKey := domainmedia.NewFrameFileKey()
		frameObjectKey := domainmedia.NewObjectKey(scanEntity.UserID, scanID, frameKey)
		filename := fmt.Sprintf("%s#frame-%02d", baseFilename, i+1)

		if err := h.storage.Put(
			ctx,
			frameObjectKey,
			bytes.NewReader(frame),
			int64(len(frame)),
			"image/jpeg",
		); err != nil {
			return fmt.Errorf("failed to store frame %d: %w", i, err)
		}

		var frameMedia *domainmedia.Media
		if i == 0 {
			if err := h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
				sourceMedia.ReplaceWithFrame(frameKey, filename, int64(len(frame)))
				if err := h.mediaRepo.Update(txCtx, sourceMedia); err != nil {
					return err
				}
				return h.outbox.StoreEvents(txCtx, sourceMedia.PullEvents())
			}); err != nil {
				return fmt.Errorf("failed to update source media as first frame: %w", err)
			}
			frameMedia = sourceMedia
		} else {
			created := domainmedia.NewMedia(scanID, frameKey, filename, "image/jpeg", int64(len(frame)))
			if err := h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
				if err := h.mediaRepo.Save(txCtx, created); err != nil {
					return err
				}
				return h.outbox.StoreEvents(txCtx, created.PullEvents())
			}); err != nil {
				return fmt.Errorf("failed to create frame media %d: %w", i, err)
			}
			frameMedia = created
		}

		if err := h.storeFrameThumbnail(ctx, scanEntity.UserID, scanID, frameMedia, frame); err != nil {
			return fmt.Errorf("failed to store thumbnail for frame %d: %w", i, err)
		}

		frameMedias = append(frameMedias, frameMedia)
	}

	allowed, message, err := h.assertBeforePipeline(ctx, scanEntity.UserID, contentType, size)
	if err != nil {
		return err
	}
	if !allowed {
		return h.failUpload(ctx, scanEntity, sourceMedia, message)
	}

	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		allEvents := make([]event.DomainEvent, 0)
		for i, frameMedia := range frameMedias {
			fresh, err := h.mediaRepo.GetByID(txCtx, frameMedia.ID)
			if err != nil {
				return err
			}
			if fresh == nil {
				return fmt.Errorf("frame media %d not found", i)
			}
			fresh.MarkUploaded()
			if err := h.mediaRepo.Update(txCtx, fresh); err != nil {
				return fmt.Errorf("failed to update frame %d status: %w", i, err)
			}
			allEvents = append(allEvents, fresh.PullEvents()...)
		}

		freshScan, err := h.scanRepo.GetByID(txCtx, scanEntity.ID)
		if err != nil {
			return err
		}
		if freshScan == nil {
			return ErrScanNotFound
		}
		freshScan.MarkProcessing()
		if err := h.scanRepo.Update(txCtx, freshScan); err != nil {
			return err
		}
		allEvents = append(allEvents, freshScan.PullEvents()...)

		return h.outbox.StoreEvents(txCtx, allEvents)
	})
}

func (h *ProcessUploadedMediaHandler) storeFrameThumbnail(
	ctx context.Context,
	userID uuid.UUID,
	scanID uuid.UUID,
	mediaEntity *domainmedia.Media,
	frame []byte,
) error {
	thumbBytes, err := h.thumbnailer.GenerateJPEG(ctx, bytes.NewReader(frame), 400)
	if err != nil {
		return err
	}

	thumbKey := domainmedia.NewThumbnailFileKey(mediaEntity.ID)
	if err := h.storage.PutThumbnail(
		ctx,
		domainmedia.NewThumbnailObjectKey(userID, scanID, mediaEntity.ID),
		bytes.NewReader(thumbBytes),
		int64(len(thumbBytes)),
		"image/jpeg",
	); err != nil {
		return err
	}

	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		fresh, err := h.mediaRepo.GetByID(txCtx, mediaEntity.ID)
		if err != nil {
			return err
		}
		if fresh == nil {
			return ErrMediaNotFound
		}
		fresh.SetThumbnail(thumbKey)
		if err := h.mediaRepo.Update(txCtx, fresh); err != nil {
			return err
		}
		mediaEntity.Thumbnail = thumbKey
		return h.outbox.StoreEvents(txCtx, fresh.PullEvents())
	})
}
