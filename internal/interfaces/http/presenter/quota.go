package presenter

import (
	"time"

	domainquota "go-api/internal/domain/quota"
)

type QuotaResponse struct {
	ID                string    `json:"id"`
	MaxImagesPerMonth int       `json:"maxImagesPerMonth"`
	MaxVideosPerMonth int       `json:"maxVideosPerMonth"`
	MaxFileSizeImage  int64     `json:"maxFileSizeImage"`
	MaxFileSizeVideo  int64     `json:"maxFileSizeVideo"`
	FullPipeline      bool      `json:"fullPipeline"`
	HistoryRetention  int64     `json:"historyRetention"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func NewQuotaResponse(view *domainquota.QuotaView) *QuotaResponse {
	if view == nil {
		return nil
	}
	return &QuotaResponse{
		ID:                view.ID.String(),
		MaxImagesPerMonth: view.MaxImagesPerMonth,
		MaxVideosPerMonth: view.MaxVideosPerMonth,
		MaxFileSizeImage:  view.MaxFileSizeImage,
		MaxFileSizeVideo:  view.MaxFileSizeVideo,
		FullPipeline:      view.FullPipeline,
		HistoryRetention:  view.HistoryRetention.Nanoseconds(),
		CreatedAt:         view.CreatedAt,
		UpdatedAt:         view.UpdatedAt,
	}
}
