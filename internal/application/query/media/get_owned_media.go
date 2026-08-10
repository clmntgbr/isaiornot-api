package media

import (
	"context"
	"errors"

	queryscan "go-api/internal/application/query/scan"
	domainmedia "go-api/internal/domain/media"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

var ErrThumbnailMediaNotFound = errors.New("media not found")

type GetOwnedMediaQuery struct {
	UserID  uuid.UUID
	MediaID uuid.UUID
}

type GetOwnedMediaHandler struct {
	mediaRepo     domainmedia.MediaReadRepository
	scanRepo      domainscan.ScanReadRepository
	historyCutoff *queryscan.HistoryCutoffResolver
}

func NewGetOwnedMediaHandler(
	mediaRepo domainmedia.MediaReadRepository,
	scanRepo domainscan.ScanReadRepository,
	historyCutoff *queryscan.HistoryCutoffResolver,
) *GetOwnedMediaHandler {
	return &GetOwnedMediaHandler{
		mediaRepo:     mediaRepo,
		scanRepo:      scanRepo,
		historyCutoff: historyCutoff,
	}
}

func (h *GetOwnedMediaHandler) Handle(ctx context.Context, query GetOwnedMediaQuery) (*domainmedia.MediaView, error) {
	mediaView, err := h.mediaRepo.FindByID(ctx, query.MediaID)
	if err != nil {
		return nil, err
	}
	if mediaView == nil {
		return nil, ErrThumbnailMediaNotFound
	}

	scanView, err := h.scanRepo.FindByID(ctx, mediaView.ScanID)
	if err != nil {
		return nil, err
	}
	if scanView == nil || scanView.UserID != query.UserID {
		return nil, ErrThumbnailMediaNotFound
	}

	since, err := h.historyCutoff.ForUser(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	if !since.IsZero() && scanView.CreatedAt.Before(since) {
		return nil, ErrThumbnailMediaNotFound
	}

	return mediaView, nil
}
