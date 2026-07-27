package media

import (
	"context"
	"fmt"

	"go-api/domain/entity"
	"go-api/domain/enum"

	"github.com/google/uuid"
)

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

	allowed, err := u.assertBeforePipeline(ctx, userID, media.ScanID, contentType, size)
	if err != nil {
		return err
	}
	if !allowed {
		return nil
	}

	if err := u.updateMediaStatusUseCase.Execute(ctx, media.ID, enum.MediaStatusUploaded); err != nil {
		return fmt.Errorf("failed to update media status: %w", err)
	}

	if err := u.publishMetadataUseCase.Execute(ctx, media.ID); err != nil {
		return fmt.Errorf("failed to publish metadata: %w", err)
	}

	return nil
}
