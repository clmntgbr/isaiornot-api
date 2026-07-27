package entity

type MediaStatistics struct {
	ScansCount     int64   `json:"scans_count"`
	RealImageCount int64   `json:"real_image_count"`
	AIImageCount   int64   `json:"ai_image_count"`
	AverageScore   float64 `json:"average_score"`
}
