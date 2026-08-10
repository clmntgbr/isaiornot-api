package presenter

import (
	"time"

	querysubscription "go-api/internal/application/query/subscription"
)

type QuotaUsageResponse struct {
	PeriodStart      time.Time `json:"periodStart"`
	PeriodEnd        time.Time `json:"periodEnd"`
	ImagesUsed       int64     `json:"imagesUsed"`
	ImagesMax        int       `json:"imagesMax"`
	ImagesLeft       int64     `json:"imagesLeft"`
	VideosUsed       int64     `json:"videosUsed"`
	VideosMax        int       `json:"videosMax"`
	VideosLeft       int64     `json:"videosLeft"`
	MaxFileSizeImage int64     `json:"maxFileSizeImage"`
	MaxFileSizeVideo int64     `json:"maxFileSizeVideo"`
	FullPipeline     bool      `json:"fullPipeline"`
}

func NewQuotaUsageResponse(usage *querysubscription.QuotaUsageView) QuotaUsageResponse {
	return QuotaUsageResponse{
		PeriodStart:      usage.PeriodStart,
		PeriodEnd:        usage.PeriodEnd,
		ImagesUsed:       usage.ImagesUsed,
		ImagesMax:        usage.ImagesMax,
		ImagesLeft:       usage.ImagesLeft,
		VideosUsed:       usage.VideosUsed,
		VideosMax:        usage.VideosMax,
		VideosLeft:       usage.VideosLeft,
		MaxFileSizeImage: usage.MaxFileSizeImage,
		MaxFileSizeVideo: usage.MaxFileSizeVideo,
		FullPipeline:     usage.FullPipeline,
	}
}
