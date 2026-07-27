package scan

import (
	"context"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"go-api/infrastructure/paginate"

	"github.com/google/uuid"
)

type GetScansUseCase struct {
	scanRepo *repository.ScanRepository
}

func NewGetScansUseCase(scanRepo *repository.ScanRepository) *GetScansUseCase {
	return &GetScansUseCase{scanRepo: scanRepo}
}

func (u *GetScansUseCase) Execute(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery) ([]*entity.Scan, int64, error) {
	scans, total, err := (*u.scanRepo).GetByUserID(ctx, userID, query)
	if err != nil {
		return []*entity.Scan{}, 0, err
	}

	return scans, total, nil
}
