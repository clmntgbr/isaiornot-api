package scan

import (
	"context"
	"errors"

	domainmedia "go-api/internal/domain/media"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

var ErrScanNotFound = errors.New("scan not found")

type GetScanByIDQuery struct {
	UserID uuid.UUID
	ScanID uuid.UUID
}

type GetScanByIDHandler struct {
	scanRepo  domainscan.ScanReadRepository
	mediaRepo domainmedia.MediaReadRepository
}

func NewGetScanByIDHandler(
	scanRepo domainscan.ScanReadRepository,
	mediaRepo domainmedia.MediaReadRepository,
) *GetScanByIDHandler {
	return &GetScanByIDHandler{
		scanRepo:  scanRepo,
		mediaRepo: mediaRepo,
	}
}

func (h *GetScanByIDHandler) Handle(ctx context.Context, q GetScanByIDQuery) (*domainscan.ScanView, error) {
	view, err := h.scanRepo.FindByID(ctx, q.ScanID)
	if err != nil {
		return nil, errors.New("failed to get scan")
	}
	if view == nil || view.UserID != q.UserID {
		return nil, ErrScanNotFound
	}

	medias, err := h.mediaRepo.FindByScanID(ctx, view.ID)
	if err != nil {
		return nil, errors.New("failed to get scan medias")
	}
	if medias == nil {
		medias = []*domainmedia.MediaView{}
	}
	view.Medias = medias

	return view, nil
}
