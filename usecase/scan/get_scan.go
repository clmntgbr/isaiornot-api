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
	scanRepo      repository.ScanRepository
	historyCutoff *HistoryCutoffResolver
}

func NewGetScanUseCase(
	scanRepo repository.ScanRepository,
	historyCutoff *HistoryCutoffResolver,
) *GetScanUseCase {
	return &GetScanUseCase{
		scanRepo:      scanRepo,
		historyCutoff: historyCutoff,
	}
}

func (u *GetScanUseCase) Execute(ctx context.Context, userID uuid.UUID, scanID uuid.UUID) (*entity.Scan, error) {
	scan, err := u.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("scan not found")
		}

		return nil, errors.New("failed to get scan")
	}

	if scan.UserID != userID {
		return nil, errors.New("scan not found")
	}

	since, err := u.historyCutoff.ForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !since.IsZero() && scan.CreatedAt.Before(since) {
		return nil, ErrHistoryOutsideRetention
	}

	return scan, nil
}
