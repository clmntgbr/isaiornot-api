package scan

import (
	"context"
	"errors"

	domainmedia "go-api/internal/domain/media"
	domainscan "go-api/internal/domain/scan"
	domainsignal "go-api/internal/domain/signal"

	"github.com/google/uuid"
)

var ErrScanNotFound = errors.New("scan not found")

type GetScanByIDQuery struct {
	UserID uuid.UUID
	ScanID uuid.UUID
}

type GetScanByIDHandler struct {
	scanRepo      domainscan.ScanReadRepository
	mediaRepo     domainmedia.MediaReadRepository
	signalRepo    domainsignal.SignalReadRepository
	historyCutoff *HistoryCutoffResolver
}

func NewGetScanByIDHandler(
	scanRepo domainscan.ScanReadRepository,
	mediaRepo domainmedia.MediaReadRepository,
	signalRepo domainsignal.SignalReadRepository,
	historyCutoff *HistoryCutoffResolver,
) *GetScanByIDHandler {
	return &GetScanByIDHandler{
		scanRepo:      scanRepo,
		mediaRepo:     mediaRepo,
		signalRepo:    signalRepo,
		historyCutoff: historyCutoff,
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

	since, err := h.historyCutoff.ForUser(ctx, q.UserID)
	if err != nil {
		return nil, err
	}
	if !since.IsZero() && view.CreatedAt.Before(since) {
		return nil, ErrHistoryOutsideRetention
	}

	medias, err := h.mediaRepo.FindByScanID(ctx, view.ID)
	if err != nil {
		return nil, errors.New("failed to get scan medias")
	}
	if medias == nil {
		medias = []*domainmedia.MediaView{}
	}

	mediaIDs := make([]uuid.UUID, 0, len(medias))
	for _, media := range medias {
		if media != nil {
			mediaIDs = append(mediaIDs, media.ID)
		}
	}

	signalsByMedia, err := h.signalRepo.FindByMediaIDs(ctx, mediaIDs)
	if err != nil {
		return nil, errors.New("failed to get media signals")
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

	view.Medias = medias
	return view, nil
}
