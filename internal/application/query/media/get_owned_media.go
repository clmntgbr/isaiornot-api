package media

import (
	"context"
	"errors"

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
	mediaRepo domainmedia.MediaReadRepository
	scanRepo  domainscan.ScanReadRepository
}

func NewGetOwnedMediaHandler(
	mediaRepo domainmedia.MediaReadRepository,
	scanRepo domainscan.ScanReadRepository,
) *GetOwnedMediaHandler {
	return &GetOwnedMediaHandler{
		mediaRepo: mediaRepo,
		scanRepo:  scanRepo,
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

	return mediaView, nil
}
