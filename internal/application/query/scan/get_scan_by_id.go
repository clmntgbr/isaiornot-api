package scan

import (
	"context"
	"errors"

	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

var ErrScanNotFound = errors.New("scan not found")

type GetScanByIDQuery struct {
	UserID uuid.UUID
	ScanID uuid.UUID
}

type GetScanByIDHandler struct {
	readRepo domainscan.ScanReadRepository
}

func NewGetScanByIDHandler(readRepo domainscan.ScanReadRepository) *GetScanByIDHandler {
	return &GetScanByIDHandler{readRepo: readRepo}
}

func (h *GetScanByIDHandler) Handle(ctx context.Context, q GetScanByIDQuery) (*domainscan.ScanView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ScanID)
	if err != nil {
		return nil, errors.New("failed to get scan")
	}
	if view == nil || view.UserID != q.UserID {
		return nil, ErrScanNotFound
	}
	return view, nil
}
