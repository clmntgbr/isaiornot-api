package scan

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/port"
	"go-api/domain/repository"
	mediadto "go-api/domain/media"

	"github.com/google/uuid"
)

var ErrUnsupportedMediaType = errors.New("unsupported media type")

type PresignUploadResult struct {
	URL    string
	ScanID uuid.UUID
}

type GeneratePresignedUploadUrlUseCase struct {
	storage   port.Storage
	scanRepo  repository.ScanRepository
	mediaRepo repository.MediaRepository
}

func NewGeneratePresignedUploadUrlUseCase(
	storage port.Storage,
	scanRepo repository.ScanRepository,
	mediaRepo repository.MediaRepository,
) *GeneratePresignedUploadUrlUseCase {
	return &GeneratePresignedUploadUrlUseCase{
		storage:   storage,
		scanRepo:  scanRepo,
		mediaRepo: mediaRepo,
	}
}

func (uc *GeneratePresignedUploadUrlUseCase) Execute(ctx context.Context, userID uuid.UUID, input mediadto.PresignUploadInput) (*PresignUploadResult, error) {
	if err := mediadto.ValidatePresignUploadInput(input); err != nil {
		return nil, errors.Join(ErrUnsupportedMediaType, err)
	}

	contentType := input.ContentType
	if contentType == "" {
		contentType = mediadto.ContentTypeFromKey(input.Filename, "")
	}

	fileKey := mediadto.NewFileKey(input.Filename)
	objectKey := mediadto.NewObjectKey(userID, fileKey)
	filename := filepath.Base(input.Filename)

	scanEntity := entity.Scan{
		UserID:   userID,
		Status:   enum.ScanStatusUploaded,
		Statuses: []enum.ScanStatus{enum.ScanStatusUploaded},
	}
	if err := uc.scanRepo.Create(ctx, &scanEntity); err != nil {
		return nil, errors.New("failed to create scan")
	}

	media := entity.Media{
		ScanID:      scanEntity.ID,
		UserID:      userID,
		Key:         fileKey,
		Filename:    filename,
		ContentType: contentType,
		Size:        0,
		Status:      enum.MediaStatusProcessing,
		Statuses:    []enum.MediaStatus{enum.MediaStatusProcessing},
	}
	if err := uc.mediaRepo.Create(ctx, &media); err != nil {
		return nil, errors.New("failed to create media")
	}

	url, err := uc.storage.PresignedPutURL(ctx, objectKey, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	return &PresignUploadResult{
		URL:    url,
		ScanID: scanEntity.ID,
	}, nil
}
