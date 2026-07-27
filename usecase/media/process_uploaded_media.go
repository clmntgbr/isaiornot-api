package media

import (
	"context"
	"errors"

	mediadto "go-api/domain/media"
	"go-api/domain/port"
	"go-api/domain/repository"
	"go-api/usecase/scan"
	"go-api/usecase/subscription"
	"go-api/usecase/thumbnail"

	"github.com/google/uuid"
)

type ProcessUploadedMediaUseCase struct {
	storage                    port.Storage
	mediaRepo                  repository.MediaRepository
	createMediaUseCase         *CreateMediaUseCase
	generateThumbnailUseCase   *GenerateThumbnailUseCase
	updateMediaStatusUseCase   *UpdateMediaStatusUseCase
	publishMetadataUseCase     *PublishMetadataUseCase
	assertUploadAllowedUseCase *subscription.AssertUploadAllowedUseCase
	failScanUseCase            *scan.FailScanUseCase
	frameExtractor             port.FrameExtractor
	imageThumbnailUseCase      *thumbnail.GenerateImageThumbnailUseCase
}

func NewProcessUploadedMediaUseCase(
	storage port.Storage,
	mediaRepo repository.MediaRepository,
	createMediaUseCase *CreateMediaUseCase,
	generateThumbnailUseCase *GenerateThumbnailUseCase,
	updateMediaStatusUseCase *UpdateMediaStatusUseCase,
	publishMetadataUseCase *PublishMetadataUseCase,
	assertUploadAllowedUseCase *subscription.AssertUploadAllowedUseCase,
	failScanUseCase *scan.FailScanUseCase,
	frameExtractor port.FrameExtractor,
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

func (u *ProcessUploadedMediaUseCase) assertBeforePipeline(
	ctx context.Context,
	userID uuid.UUID,
	scanID uuid.UUID,
	contentType string,
	size int64,
) (bool, error) {
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
			return false, failErr
		}
		return false, nil
	}
	return true, nil
}
