package media

import (
	"context"
	"errors"
	"go-api/domain/enum"
	"go-api/domain/repository"
	"go-api/usecase/scan"

	"github.com/google/uuid"
)

type UpdateMediaStatusUseCase struct {
	mediaRepo               *repository.MediaRepository
	updateScanStatusUseCase *scan.UpdateScanStatusUseCase
}

func NewUpdateMediaStatusUseCase(
	mediaRepo *repository.MediaRepository,
	updateScanStatusUseCase *scan.UpdateScanStatusUseCase,
) *UpdateMediaStatusUseCase {
	return &UpdateMediaStatusUseCase{
		mediaRepo:               mediaRepo,
		updateScanStatusUseCase: updateScanStatusUseCase,
	}
}

func (u *UpdateMediaStatusUseCase) Execute(ctx context.Context, mediaID uuid.UUID, status enum.MediaStatus) error {
	media, err := (*u.mediaRepo).GetByID(ctx, mediaID)
	if err != nil {
		return errors.New("failed to get media")
	}

	media.Statuses = append(media.Statuses, status)
	media.Status = media.Statuses[len(media.Statuses)-1]

	if err := (*u.mediaRepo).Update(ctx, media); err != nil {
		return err
	}

	return u.updateScanStatusUseCase.Execute(ctx, media.ScanID)
}
