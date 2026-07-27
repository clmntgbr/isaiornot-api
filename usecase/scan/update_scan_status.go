package scan

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/enum"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type UpdateScanStatusUseCase struct {
	scanRepo repository.ScanRepository
}

func NewUpdateScanStatusUseCase(
	scanRepo repository.ScanRepository,
) *UpdateScanStatusUseCase {
	return &UpdateScanStatusUseCase{
		scanRepo: scanRepo,
	}
}

func (u *UpdateScanStatusUseCase) Execute(ctx context.Context, scanID uuid.UUID) error {
	scan, err := u.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return errors.New("failed to get scan")
	}

	if scan.Status == enum.ScanStatusFailed {
		return nil
	}

	status := ResolveStatus(scan.Medias)
	if scan.Status == status {
		return nil
	}

	scan.Statuses = append(scan.Statuses, status)
	scan.Status = status
	scan.Message = ""

	if err := u.scanRepo.Update(ctx, scan); err != nil {
		return errors.New("failed to update scan status")
	}

	return nil
}

func ResolveStatus(medias []entity.Media) enum.ScanStatus {
	if len(medias) == 0 {
		return enum.ScanStatusUploaded
	}

	allAnalyzed := true
	hasUploaded := false
	for _, media := range medias {
		if media.Status != enum.MediaStatusAnalyzed {
			allAnalyzed = false
		}
		if media.Status == enum.MediaStatusUploaded || media.Status == enum.MediaStatusAnalyzed {
			hasUploaded = true
		}
	}

	if allAnalyzed {
		return enum.ScanStatusAnalyzed
	}
	if hasUploaded {
		return enum.ScanStatusProcessing
	}

	return enum.ScanStatusUploaded
}
