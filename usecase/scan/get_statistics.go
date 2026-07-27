package scan

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type GetStatisticsUseCase struct {
	scanRepo repository.ScanRepository
}

func NewGetStatisticsUseCase(scanRepo repository.ScanRepository) *GetStatisticsUseCase {
	return &GetStatisticsUseCase{scanRepo: scanRepo}
}

func (u *GetStatisticsUseCase) Execute(ctx context.Context, userID uuid.UUID) (*entity.MediaStatistics, error) {
	stats, err := u.scanRepo.GetStatisticsByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to get scan statistics")
	}

	return stats, nil
}
