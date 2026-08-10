package scan

import (
	"context"
	"errors"

	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

type GetStatisticsQuery struct {
	UserID uuid.UUID
}

type GetStatisticsHandler struct {
	scanRepo domainscan.ScanReadRepository
}

func NewGetStatisticsHandler(scanRepo domainscan.ScanReadRepository) *GetStatisticsHandler {
	return &GetStatisticsHandler{scanRepo: scanRepo}
}

func (h *GetStatisticsHandler) Handle(ctx context.Context, q GetStatisticsQuery) (*domainscan.StatisticsView, error) {
	stats, err := h.scanRepo.GetStatisticsByUserID(ctx, q.UserID)
	if err != nil {
		return nil, errors.New("failed to get scan statistics")
	}
	return stats, nil
}
