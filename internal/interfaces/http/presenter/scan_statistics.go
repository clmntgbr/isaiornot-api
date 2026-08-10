package presenter

import domainscan "go-api/internal/domain/scan"

type ScanStatisticsResponse struct {
	ScansCount     int64   `json:"scansCount"`
	RealImageCount int64   `json:"realImageCount"`
	AIImageCount   int64   `json:"aiImageCount"`
	AverageScore   float64 `json:"averageScore"`
}

func NewScanStatisticsResponse(stats *domainscan.StatisticsView) ScanStatisticsResponse {
	return ScanStatisticsResponse{
		ScansCount:     stats.ScansCount,
		RealImageCount: stats.RealImageCount,
		AIImageCount:   stats.AIImageCount,
		AverageScore:   stats.AverageScore,
	}
}
