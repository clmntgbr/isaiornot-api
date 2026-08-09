package scan

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

type ListScansQuery struct {
	UserID uuid.UUID
	Query  paginate.PaginateQuery
}

type ListScansHandler struct {
	readRepo domainscan.ScanReadRepository
}

func NewListScansHandler(readRepo domainscan.ScanReadRepository) *ListScansHandler {
	return &ListScansHandler{readRepo: readRepo}
}

func (h *ListScansHandler) Handle(ctx context.Context, q ListScansQuery) ([]*domainscan.ScanView, int64, error) {
	scans, total, err := h.readRepo.FindByUserID(ctx, q.UserID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list scans")
	}
	return scans, total, nil
}
