package scan

import (
	"context"
	"errors"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/paginate"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

type ListScansQuery struct {
	UserID uuid.UUID
	Query  paginate.PaginateQuery
}

type ListScansHandler struct {
	scanRepo  domainscan.ScanReadRepository
	mediaRepo domainmedia.MediaReadRepository
}

func NewListScansHandler(
	scanRepo domainscan.ScanReadRepository,
	mediaRepo domainmedia.MediaReadRepository,
) *ListScansHandler {
	return &ListScansHandler{
		scanRepo:  scanRepo,
		mediaRepo: mediaRepo,
	}
}

func (h *ListScansHandler) Handle(ctx context.Context, q ListScansQuery) ([]*domainscan.ScanView, int64, error) {
	scans, total, err := h.scanRepo.FindByUserID(ctx, q.UserID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list scans")
	}

	scanIDs := make([]uuid.UUID, 0, len(scans))
	for _, scan := range scans {
		if scan != nil {
			scanIDs = append(scanIDs, scan.ID)
		}
	}

	mediasByScan, err := h.mediaRepo.FindByScanIDs(ctx, scanIDs)
	if err != nil {
		return nil, 0, errors.New("failed to list scan medias")
	}

	for _, scan := range scans {
		if scan == nil {
			continue
		}
		medias := mediasByScan[scan.ID]
		if medias == nil {
			medias = []*domainmedia.MediaView{}
		}
		scan.Medias = medias
	}

	return scans, total, nil
}
