package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/repository"
	mediadto "go-api/infrastructure/media"
	"go-api/infrastructure/storage"
	"go-api/infrastructure/video"
	"go-api/usecase/scan"
	"go-api/usecase/subscription"
	"go-api/usecase/thumbnail"

	"github.com/google/uuid"
)

type ProcessUploadedMediaUseCase struct {
	storage                    *storage.MinIOStorage
	mediaRepo                  *repository.MediaRepository
	createMediaUseCase         *CreateMediaUseCase
	generateThumbnailUseCase   *GenerateThumbnailUseCase
	updateMediaStatusUseCase   *UpdateMediaStatusUseCase
	publishMetadataUseCase     *PublishMetadataUseCase
	assertUploadAllowedUseCase *subscription.AssertUploadAllowedUseCase
	failScanUseCase            *scan.FailScanUseCase
	frameExtractor             *video.FrameExtractor
	imageThumbnailUseCase      *thumbnail.GenerateImageThumbnailUseCase
}

func NewProcessUploadedMediaUseCase(
	storage *storage.MinIOStorage,
	mediaRepo *repository.MediaRepository,
	createMediaUseCase *CreateMediaUseCase,
	generateThumbnailUseCase *GenerateThumbnailUseCase,
	updateMediaStatusUseCase *UpdateMediaStatusUseCase,
	publishMetadataUseCase *PublishMetadataUseCase,
	assertUploadAllowedUseCase *subscription.AssertUploadAllowedUseCase,
	failScanUseCase *scan.FailScanUseCase,
	frameExtractor *video.FrameExtractor,
	imageThumbnailUseCase *thumbnail.GenerateImageThumbnailUseCase,
) *ProcessUploadedMediaUseCase {
	return &ProcessUploadedMediaUseCase{
		storage:                    storage,
		mediaRepo:                  mediaRepo,
		createMediaUseCase:         createMediaUseCase,
		generateThumbnailUseCase:   generateThumbnailUseCase,
		updateMediaStatusUseCase:   updateMediaStatusUseCase,
		publishMetadataUseCase:     publishMetadataUseCase,
		assertUploadAllowedUseCase: assertUploadAllowedUseCase,
		failScanUseCase:            failScanUseCase,
		frameExtractor:             frameExtractor,
		imageThumbnailUseCase:      imageThumbnailUseCase,
	}
}

func (u *ProcessUploadedMediaUseCase) Execute(ctx context.Context, userID uuid.UUID, fileKey string, contentType string, size int64) error {
	contentType = mediadto.ContentTypeFromKey(fileKey, contentType)

	sourceMedia, err := u.createMediaUseCase.Execute(ctx, userID, fileKey, contentType, size)
	if err != nil {
		return err
	}

	if mediadto.IsVideoContentType(contentType) {
		return u.processVideo(ctx, userID, sourceMedia, contentType, size)
	}

	return u.processImage(ctx, userID, sourceMedia, contentType, size)
}

func (u *ProcessUploadedMediaUseCase) processImage(
	ctx context.Context,
	userID uuid.UUID,
	media *entity.Media,
	contentType string,
	size int64,
) error {
	if err := u.generateThumbnailUseCase.Execute(ctx, userID, media.ID); err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	if err := u.assertBeforePipeline(ctx, userID, media.ScanID, contentType, size); err != nil {
		return err
	}

	if err := u.updateMediaStatusUseCase.Execute(ctx, media.ID, enum.MediaStatusUploaded); err != nil {
		return fmt.Errorf("failed to update media status: %w", err)
	}

	if err := u.publishMetadataUseCase.Execute(ctx, media.ID); err != nil {
		return fmt.Errorf("failed to publish metadata: %w", err)
	}

	return nil
}

func (u *ProcessUploadedMediaUseCase) processVideo(
	ctx context.Context,
	userID uuid.UUID,
	sourceMedia *entity.Media,
	contentType string,
	size int64,
) error {
	objectKey := mediadto.NewObjectKey(userID, sourceMedia.Key)
	reader, err := u.storage.Get(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}
	defer reader.Close()

	videoData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read video: %w", err)
	}

	frames, err := u.frameExtractor.ExtractFrames(videoData, mediadto.MaxVideoFrames())
	if err != nil {
		return fmt.Errorf("failed to extract frames: %w", err)
	}
	if len(frames) == 0 {
		return errors.New("no frames extracted from video")
	}

	baseFilename := sourceMedia.Filename
	scanID := sourceMedia.ScanID
	frameMedias := make([]*entity.Media, 0, len(frames))

	for i, frame := range frames {
		frameKey := mediadto.NewFrameFileKey()
		frameObjectKey := mediadto.NewObjectKey(userID, frameKey)
		filename := baseFilename
		if i > 0 {
			filename = fmt.Sprintf("%s#frame-%02d", baseFilename, i+1)
		}

		if err := u.storage.Put(ctx, frameObjectKey, bytes.NewReader(frame), int64(len(frame)), "image/jpeg"); err != nil {
			return fmt.Errorf("failed to store frame %d: %w", i, err)
		}

		var frameMedia *entity.Media
		if i == 0 {
			sourceMedia.Key = frameKey
			sourceMedia.Filename = filename
			sourceMedia.ContentType = "image/jpeg"
			sourceMedia.Size = int64(len(frame))
			if err := (*u.mediaRepo).Update(ctx, sourceMedia); err != nil {
				return fmt.Errorf("failed to update source media as first frame: %w", err)
			}
			frameMedia = sourceMedia
		} else {
			created := entity.Media{
				ScanID:      scanID,
				UserID:      userID,
				Key:         frameKey,
				Filename:    filename,
				ContentType: "image/jpeg",
				Size:        int64(len(frame)),
				Status:      enum.MediaStatusProcessing,
				Statuses:    []enum.MediaStatus{enum.MediaStatusProcessing},
			}
			if err := (*u.mediaRepo).Create(ctx, &created); err != nil {
				return fmt.Errorf("failed to create frame media %d: %w", i, err)
			}
			frameMedia = &created
		}

		if err := u.storeFrameThumbnail(ctx, userID, frameMedia, frame); err != nil {
			return fmt.Errorf("failed to store thumbnail for frame %d: %w", i, err)
		}

		frameMedias = append(frameMedias, frameMedia)
	}

	if err := u.assertBeforePipeline(ctx, userID, scanID, contentType, size); err != nil {
		return err
	}

	for i, frameMedia := range frameMedias {
		if err := u.updateMediaStatusUseCase.Execute(ctx, frameMedia.ID, enum.MediaStatusUploaded); err != nil {
			return fmt.Errorf("failed to update frame %d status: %w", i, err)
		}

		if err := u.publishMetadataUseCase.Execute(ctx, frameMedia.ID); err != nil {
			return fmt.Errorf("failed to publish frame %d scan: %w", i, err)
		}
	}

	return nil
}

func (u *ProcessUploadedMediaUseCase) assertBeforePipeline(
	ctx context.Context,
	userID uuid.UUID,
	scanID uuid.UUID,
	contentType string,
	size int64,
) error {
	if err := u.assertUploadAllowedUseCase.Execute(ctx, subscription.AssertUploadAllowedInput{
		UserID:              userID,
		ContentType:         contentType,
		Size:                size,
		MediaAlreadyCounted: true,
	}); err != nil {
		message := err.Error()
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			message = "No active subscription found"
		}
		if failErr := u.failScanUseCase.Execute(ctx, scanID, message); failErr != nil {
			return failErr
		}
		return nil
	}
	return nil
}

func (u *ProcessUploadedMediaUseCase) storeFrameThumbnail(ctx context.Context, userID uuid.UUID, media *entity.Media, frame []byte) error {
	thumbBytes, err := u.imageThumbnailUseCase.Execute(ctx, bytes.NewReader(frame), 400)
	if err != nil {
		return err
	}

	thumbKey := mediadto.NewThumbnailFileKey(media.ID)
	if err := u.storage.PutThumbnail(
		ctx,
		mediadto.NewThumbnailObjectKey(userID, media.ID),
		bytes.NewReader(thumbBytes),
		int64(len(thumbBytes)),
		"image/jpeg",
	); err != nil {
		return err
	}

	media.Thumbnail = thumbKey
	return (*u.mediaRepo).Update(ctx, media)
}
