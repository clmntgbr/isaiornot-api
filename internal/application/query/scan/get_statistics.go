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
	scanRepo      domainscan.ScanReadRepository
	historyCutoff *HistoryCutoffResolver
}

func NewGetStatisticsHandler(
	scanRepo domainscan.ScanReadRepository,
	historyCutoff *HistoryCutoffResolver,
) *GetStatisticsHandler {
	return &GetStatisticsHandler{
		scanRepo:      scanRepo,
		historyCutoff: historyCutoff,
	}
}

func (h *GetStatisticsHandler) Handle(ctx context.Context, q GetStatisticsQuery) (*domainscan.StatisticsView, error) {
	since, err := h.historyCutoff.ForUser(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	stats, err := h.scanRepo.GetStatisticsByUserID(ctx, q.UserID, since)
	if err != nil {
		return nil, errors.New("failed to get scan statistics")
	}
	return stats, nil
}
