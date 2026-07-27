package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"go-api/domain/entity"
	"go-api/domain/enum"
	mediadto "go-api/domain/media"

	"github.com/google/uuid"
)

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
			if err := u.mediaRepo.Update(ctx, sourceMedia); err != nil {
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
			if err := u.mediaRepo.Create(ctx, &created); err != nil {
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
	return u.mediaRepo.Update(ctx, media)
}
