package scan

import (
	"context"
	"errors"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/paginate"
	domainscan "go-api/internal/domain/scan"
	domainsignal "go-api/internal/domain/signal"

	"github.com/google/uuid"
)

type ListScansQuery struct {
	UserID uuid.UUID
	Query  paginate.PaginateQuery
}

type ListScansHandler struct {
	scanRepo      domainscan.ScanReadRepository
	mediaRepo     domainmedia.MediaReadRepository
	signalRepo    domainsignal.SignalReadRepository
	historyCutoff *HistoryCutoffResolver
}

func NewListScansHandler(
	scanRepo domainscan.ScanReadRepository,
	mediaRepo domainmedia.MediaReadRepository,
	signalRepo domainsignal.SignalReadRepository,
	historyCutoff *HistoryCutoffResolver,
) *ListScansHandler {
	return &ListScansHandler{
		scanRepo:      scanRepo,
		mediaRepo:     mediaRepo,
		signalRepo:    signalRepo,
		historyCutoff: historyCutoff,
	}
}

func (h *ListScansHandler) Handle(ctx context.Context, q ListScansQuery) ([]*domainscan.ScanView, int64, error) {
	since, err := h.historyCutoff.ForUser(ctx, q.UserID)
	if err != nil {
		return nil, 0, err
	}

	scans, total, err := h.scanRepo.FindByUserID(ctx, q.UserID, q.Query, since)
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

	mediaIDs := make([]uuid.UUID, 0)
	for _, medias := range mediasByScan {
		for _, media := range medias {
			if media != nil {
				mediaIDs = append(mediaIDs, media.ID)
			}
		}
	}

	signalsByMedia, err := h.signalRepo.FindByMediaIDs(ctx, mediaIDs)
	if err != nil {
		return nil, 0, errors.New("failed to list media signals")
	}

	for _, scan := range scans {
		if scan == nil {
			continue
		}
		medias := mediasByScan[scan.ID]
		if medias == nil {
			medias = []*domainmedia.MediaView{}
		}
		for _, media := range medias {
			if media == nil {
				continue
			}
			signals := signalsByMedia[media.ID]
			if signals == nil {
				signals = []*domainsignal.SignalView{}
			}
			media.Signals = signals
		}
		scan.Medias = medias
	}

	return scans, total, nil
}
