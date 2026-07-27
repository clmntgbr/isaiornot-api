package scan

import (
	"context"
	"errors"

	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type GetMediaStatisticsUseCase struct {
	scanRepo      repository.ScanRepository
	historyCutoff *HistoryCutoffResolver
}

func NewGetMediaStatisticsUseCase(
	scanRepo repository.ScanRepository,
	historyCutoff *HistoryCutoffResolver,
) *GetMediaStatisticsUseCase {
	return &GetMediaStatisticsUseCase{
		scanRepo:      scanRepo,
		historyCutoff: historyCutoff,
	}
}

func (u *GetMediaStatisticsUseCase) Execute(ctx context.Context, userID uuid.UUID) (*entity.MediaStatistics, error) {
	since, err := u.historyCutoff.ForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats, err := u.scanRepo.GetStatisticsByUserID(ctx, userID, since)
	if err != nil {
		return nil, errors.New("failed to get scan statistics")
	}

	return stats, nil
}
