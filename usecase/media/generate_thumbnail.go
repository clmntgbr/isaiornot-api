package media

import (
	"bytes"
	"context"
	"errors"

	"go-api/domain/port"
	"go-api/domain/repository"
	mediadto "go-api/domain/media"
	"go-api/usecase/thumbnail"

	"github.com/google/uuid"
)

type GenerateThumbnailUseCase struct {
	storage                  port.Storage
	mediaRepo                repository.MediaRepository
	generateThumbnailUseCase *thumbnail.GenerateImageThumbnailUseCase
}

func NewGenerateThumbnailUseCase(storage port.Storage, mediaRepo repository.MediaRepository, generateThumbnailUseCase *thumbnail.GenerateImageThumbnailUseCase) *GenerateThumbnailUseCase {
	return &GenerateThumbnailUseCase{storage: storage, mediaRepo: mediaRepo, generateThumbnailUseCase: generateThumbnailUseCase}
}

func (uc *GenerateThumbnailUseCase) Execute(ctx context.Context, userID uuid.UUID, mediaID uuid.UUID) error {
	media, err := uc.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return errors.New("media not found")
	}

	original, err := uc.storage.Get(ctx, mediadto.NewObjectKey(userID, media.Key))
	if err != nil {
		return errors.New("failed to fetch original")
	}
	defer original.Close()

	var thumbBytes []byte
	contentType := mediadto.ContentTypeFromKey(media.Key, media.ContentType)
	if mediadto.IsImageContentType(contentType) {
		thumbBytes, err = uc.generateThumbnailUseCase.Execute(ctx, original, 400)
	} else {
		return errors.New("unsupported content type")
	}
	if err != nil {
		return err
	}

	thumbKey := mediadto.NewThumbnailFileKey(media.ID)
	if err := uc.storage.PutThumbnail(
		ctx,
		mediadto.NewThumbnailObjectKey(userID, media.ID),
		bytes.NewReader(thumbBytes),
		int64(len(thumbBytes)),
		"image/jpeg",
	); err != nil {
		return errors.New("failed to store thumbnail")
	}

	media.Thumbnail = thumbKey
	return uc.mediaRepo.Update(ctx, media)
}
