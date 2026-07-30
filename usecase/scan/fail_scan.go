package scan

import (
	"context"
	"errors"

	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/port"
	"go-api/domain/realtime"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type OriginalsDeleter interface {
	Execute(ctx context.Context, userID uuid.UUID, medias []entity.Media)
}

type FailScanUseCase struct {
	scanRepo            repository.ScanRepository
	mediaRepo           repository.MediaRepository
	centrifugoPublisher port.RealtimePublisher
	deleteOriginals     OriginalsDeleter
}

func NewFailScanUseCase(
	scanRepo repository.ScanRepository,
	mediaRepo repository.MediaRepository,
	centrifugoPublisher port.RealtimePublisher,
	deleteOriginals OriginalsDeleter,
) *FailScanUseCase {
	return &FailScanUseCase{
		scanRepo:            scanRepo,
		mediaRepo:           mediaRepo,
		centrifugoPublisher: centrifugoPublisher,
		deleteOriginals:     deleteOriginals,
	}
}

func (u *FailScanUseCase) Execute(ctx context.Context, scanID uuid.UUID, message string) error {
	scan, err := u.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return errors.New("failed to get scan")
	}

	for i := range scan.Medias {
		media := &scan.Medias[i]
		if media.Status == enum.MediaStatusFailed {
			continue
		}
		media.Statuses = append(media.Statuses, enum.MediaStatusFailed)
		media.Status = enum.MediaStatusFailed
		if err := u.mediaRepo.Update(ctx, media); err != nil {
			return errors.New("failed to update media status")
		}
	}

	scan.Message = message
	if scan.Status != enum.ScanStatusFailed {
		scan.Statuses = append(scan.Statuses, enum.ScanStatusFailed)
		scan.Status = enum.ScanStatusFailed
	}
	scan.FreezeDuration()

	if err := u.scanRepo.Update(ctx, scan); err != nil {
		return errors.New("failed to update scan")
	}

	if u.deleteOriginals != nil {
		u.deleteOriginals.Execute(ctx, scan.UserID, scan.Medias)
	}

	scan, err = u.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return errors.New("failed to reload scan")
	}

	event, err := realtime.NewScanFailedEvent(scan)
	if err != nil {
		return errors.New("failed to build scan failed event")
	}

	if err := u.centrifugoPublisher.PublishToUser(ctx, scan.UserID, event); err != nil {
		return errors.New("failed to publish scan failed event")
	}

	return nil
}
