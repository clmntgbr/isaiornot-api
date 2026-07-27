package scan

import (
	"context"

	"go-api/domain/entity"
	"go-api/domain/repository"
	domainScan "go-api/domain/scan"

	"github.com/google/uuid"
)

type GetScansUseCase struct {
	scanRepo      repository.ScanRepository
	historyCutoff *HistoryCutoffResolver
}

func NewGetScansUseCase(
	scanRepo repository.ScanRepository,
	historyCutoff *HistoryCutoffResolver,
) *GetScansUseCase {
	return &GetScansUseCase{
		scanRepo:      scanRepo,
		historyCutoff: historyCutoff,
	}
}

func (u *GetScansUseCase) Execute(ctx context.Context, userID uuid.UUID, query domainScan.ListQuery) ([]*entity.Scan, int64, error) {
	since, err := u.historyCutoff.ForUser(ctx, userID)
	if err != nil {
		return []*entity.Scan{}, 0, err
	}

	scans, total, err := u.scanRepo.GetByUserID(ctx, userID, query, since)
	if err != nil {
		return []*entity.Scan{}, 0, err
	}

	return scans, total, nil
}
