package analysis

import (
	"context"
	"errors"
	"path/filepath"

	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/repository"
	mediadto "go-api/infrastructure/media"
	"go-api/infrastructure/storage"
	"time"

	"github.com/google/uuid"
)

var ErrUnsupportedMediaType = errors.New("unsupported media type")

type GeneratePresignedUploadUrlUseCase struct {
	storage      *storage.MinIOStorage
	analysisRepo *repository.AnalysisRepository
	mediaRepo    *repository.MediaRepository
}

func NewGeneratePresignedUploadUrlUseCase(
	storage *storage.MinIOStorage,
	analysisRepo *repository.AnalysisRepository,
	mediaRepo *repository.MediaRepository,
) *GeneratePresignedUploadUrlUseCase {
	return &GeneratePresignedUploadUrlUseCase{
		storage:      storage,
		analysisRepo: analysisRepo,
		mediaRepo:    mediaRepo,
	}
}

func (uc *GeneratePresignedUploadUrlUseCase) Execute(ctx context.Context, userID uuid.UUID, input mediadto.PresignUploadInput) (string, error) {
	if err := mediadto.ValidatePresignUploadInput(input); err != nil {
		return "", errors.Join(ErrUnsupportedMediaType, err)
	}

	fileKey := mediadto.NewFileKey(input.Filename)
	objectKey := mediadto.NewObjectKey(userID, fileKey)
	filename := filepath.Base(input.Filename)

	analysisEntity := entity.Analysis{
		UserID:   userID,
		Status:   enum.AnalysisStatusProcessing,
		Statuses: []enum.AnalysisStatus{enum.AnalysisStatusProcessing},
	}
	if err := (*uc.analysisRepo).Create(ctx, &analysisEntity); err != nil {
		return "", errors.New("failed to create analysis")
	}

	media := entity.Media{
		AnalysisID:  analysisEntity.ID,
		UserID:      userID,
		Key:         fileKey,
		Filename:    filename,
		ContentType: input.ContentType,
		Size:        0,
		Status:      enum.MediaStatusProcessing,
		Statuses:    []enum.MediaStatus{enum.MediaStatusProcessing},
	}
	if err := (*uc.mediaRepo).Create(ctx, &media); err != nil {
		return "", errors.New("failed to create media")
	}

	url, err := uc.storage.PresignedPutURL(ctx, objectKey, 15*time.Minute)
	if err != nil {
		return "", err
	}

	return url, nil
}
