package presenter

import (
	"go-api/domain/entity"
	"time"
)

type SignalResponse struct {
	Name       string   `json:"name"`
	Score      int      `json:"score"`
	Confidence string   `json:"confidence"`
	Details    []string `json:"details"`
}

type InsightResponse struct {
	Noise       float64 `json:"noise"`
	Compression float64 `json:"compression"`
	Frequency   float64 `json:"frequency"`
	Histogram   float64 `json:"histogram"`
}

type MediaItemResponse struct {
	ID          string           `json:"id"`
	Key         string           `json:"key"`
	Filename    string           `json:"filename"`
	Thumbnail   string           `json:"thumbnail"`
	ContentType string           `json:"contentType"`
	Status      string           `json:"status"`
	Signals     []SignalResponse `json:"signals"`
	Insight     *InsightResponse `json:"insight"`
	Size        int64            `json:"size"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

func NewMediaItemResponse(media entity.Media) MediaItemResponse {
	item := MediaItemResponse{
		ID:          media.ID.String(),
		Key:         media.Key,
		Filename:    media.Filename,
		Thumbnail:   thumbnailURL(media),
		ContentType: media.ContentType,
		Status:      string(media.Status),
		Signals:     NewSignalResponses(media.Signals),
		Size:        media.Size,
		CreatedAt:   media.CreatedAt,
		UpdatedAt:   media.UpdatedAt,
	}

	if media.Insight != nil {
		insight := NewInsightResponse(media.Insight)
		item.Insight = &insight
	}

	return item
}

func NewMediaItemResponses(medias []entity.Media) []MediaItemResponse {
	items := make([]MediaItemResponse, 0, len(medias))
	for _, media := range medias {
		items = append(items, NewMediaItemResponse(media))
	}
	return items
}

func NewSignalResponses(signals []entity.Signal) []SignalResponse {
	responses := make([]SignalResponse, 0, len(signals))
	for _, signal := range signals {
		responses = append(responses, SignalResponse{
			Name:       signal.Name,
			Score:      signal.Score,
			Confidence: string(signal.Confidence),
			Details:    signal.Details,
		})
	}

	return responses
}

func NewInsightResponse(insight *entity.Insight) InsightResponse {
	return InsightResponse{
		Noise:       insight.Noise,
		Compression: insight.Compression,
		Frequency:   insight.Frequency,
		Histogram:   insight.Histogram,
	}
}

func thumbnailURL(media entity.Media) string {
	if media.Thumbnail == "" {
		return ""
	}

	return "/api/medias/" + media.ID.String() + "/thumbnail"
}
