package media

import (
	"context"
	"log"

	"go-api/domain/entity"
	mediadto "go-api/domain/media"
	"go-api/domain/port"

	"github.com/google/uuid"
)

type DeleteOriginalsUseCase struct {
	storage port.Storage
}

func NewDeleteOriginalsUseCase(storage port.Storage) *DeleteOriginalsUseCase {
	return &DeleteOriginalsUseCase{storage: storage}
}

func (u *DeleteOriginalsUseCase) Execute(ctx context.Context, userID uuid.UUID, medias []entity.Media) {
	for _, media := range medias {
		if media.Key == "" {
			continue
		}

		objectKey := mediadto.NewObjectKey(userID, media.Key)
		if err := u.storage.Delete(ctx, objectKey); err != nil {
			log.Printf("failed to delete original object %q: %v", objectKey, err)
		}
	}
}
