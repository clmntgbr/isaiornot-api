package entity

type MediaStatistics struct {
	ScansCount     int64   `json:"scansCount"`
	RealImageCount int64   `json:"realImageCount"`
	AIImageCount   int64   `json:"aiImageCount"`
	AverageScore   float64 `json:"averageScore"`
}
