package scan

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type GetMediaStatisticsUseCase struct {
	scanRepo repository.ScanRepository
}

func NewGetMediaStatisticsUseCase(scanRepo repository.ScanRepository) *GetMediaStatisticsUseCase {
	return &GetMediaStatisticsUseCase{scanRepo: scanRepo}
}

func (u *GetMediaStatisticsUseCase) Execute(ctx context.Context, userID uuid.UUID) (*entity.MediaStatistics, error) {
	stats, err := u.scanRepo.GetStatisticsByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to get scan statistics")
	}

	return stats, nil
}
