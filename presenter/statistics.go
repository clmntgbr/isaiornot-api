package presenter

import "go-api/domain/entity"

type MediaStatisticsResponse struct {
	ScansCount     int64   `json:"scansCount"`
	RealImageCount int64   `json:"realImageCount"`
	AIImageCount   int64   `json:"aiImageCount"`
	AverageScore   float64 `json:"averageScore"`
}

func NewMediaStatisticsResponse(stats *entity.MediaStatistics) MediaStatisticsResponse {
	return MediaStatisticsResponse{
		ScansCount:     stats.ScansCount,
		RealImageCount: stats.RealImageCount,
		AIImageCount:   stats.AIImageCount,
		AverageScore:   stats.AverageScore,
	}
}
