package scan

import (
	"context"
	"errors"

	"go-api/domain/enum"
	"go-api/domain/port"
	"go-api/domain/realtime"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type FailScanUseCase struct {
	scanRepo            repository.ScanRepository
	centrifugoPublisher port.RealtimePublisher
}

func NewFailScanUseCase(
	scanRepo repository.ScanRepository,
	centrifugoPublisher port.RealtimePublisher,
) *FailScanUseCase {
	return &FailScanUseCase{
		scanRepo:            scanRepo,
		centrifugoPublisher: centrifugoPublisher,
	}
}

func (u *FailScanUseCase) Execute(ctx context.Context, scanID uuid.UUID, message string) error {
	scan, err := u.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return errors.New("failed to get scan")
	}

	scan.Message = message
	if scan.Status != enum.ScanStatusFailed {
		scan.Statuses = append(scan.Statuses, enum.ScanStatusFailed)
		scan.Status = enum.ScanStatusFailed
	}

	if err := u.scanRepo.Update(ctx, scan); err != nil {
		return errors.New("failed to update scan")
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
