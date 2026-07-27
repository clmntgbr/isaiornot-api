package media

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type CreateMediaUseCase struct {
	scanRepo  *repository.ScanRepository
	mediaRepo *repository.MediaRepository
}

func NewCreateMediaUseCase(
	scanRepo *repository.ScanRepository,
	mediaRepo *repository.MediaRepository,
) *CreateMediaUseCase {
	return &CreateMediaUseCase{
		scanRepo:  scanRepo,
		mediaRepo: mediaRepo,
	}
}

func (u *CreateMediaUseCase) Execute(ctx context.Context, userID uuid.UUID, key string, contentType string, size int64) (*entity.Media, error) {
	existing, err := (*u.mediaRepo).GetByKey(ctx, key)
	if err == nil {
		existing.ContentType = contentType
		existing.Size = size
		if err := (*u.mediaRepo).Update(ctx, existing); err != nil {
			return nil, errors.New("failed to update media")
		}
		return existing, nil
	}

	scan := entity.Scan{
		UserID:   userID,
		Status:   enum.ScanStatusUploaded,
		Statuses: []enum.ScanStatus{enum.ScanStatusUploaded},
	}
	if err := (*u.scanRepo).Create(ctx, &scan); err != nil {
		return nil, errors.New("failed to create scan")
	}

	media := entity.Media{
		ScanID:      scan.ID,
		UserID:      userID,
		Key:         key,
		ContentType: contentType,
		Size:        size,
		Status:      enum.MediaStatusProcessing,
		Statuses:    []enum.MediaStatus{enum.MediaStatusProcessing},
	}
	if err := (*u.mediaRepo).Create(ctx, &media); err != nil {
		return nil, errors.New("failed to create media")
	}

	return &media, nil
}
