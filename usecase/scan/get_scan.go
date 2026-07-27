package scan

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GetScanUseCase struct {
	scanRepo *repository.ScanRepository
}

func NewGetScanUseCase(scanRepo *repository.ScanRepository) *GetScanUseCase {
	return &GetScanUseCase{scanRepo: scanRepo}
}

func (u *GetScanUseCase) Execute(ctx context.Context, userID uuid.UUID, scanID uuid.UUID) (*entity.Scan, error) {
	scan, err := (*u.scanRepo).GetByID(ctx, scanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("scan not found")
		}

		return nil, errors.New("failed to get scan")
	}

	if scan.UserID != userID {
		return nil, errors.New("scan not found")
	}

	return scan, nil
}
