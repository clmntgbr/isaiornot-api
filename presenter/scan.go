package presenter

import (
	"go-api/domain/entity"
	"go-api/domain/enum"
	"time"
)

type GeneratePresignedUploadUrlDetailResponse struct {
	UploadURL string `json:"uploadUrl"`
	ScanID    string `json:"scanId"`
}

func NewGeneratePresignedUploadUrlDetailResponse(url string, scanID string) GeneratePresignedUploadUrlDetailResponse {
	return GeneratePresignedUploadUrlDetailResponse{
		UploadURL: url,
		ScanID:    scanID,
	}
}

type ScanListResponse struct {
	ID         string              `json:"id"`
	Status     string              `json:"status"`
	Statuses   []string            `json:"statuses"`
	Message    string              `json:"message"`
	FinalScore float64             `json:"finalScore"`
	Confidence string              `json:"confidence"`
	Verdict    string              `json:"verdict"`
	Duration   int                 `json:"duration"`
	Filename   string              `json:"filename"`
	Thumbnail  string              `json:"thumbnail"`
	RetryCount int                 `json:"retryCount"`
	Medias     []MediaItemResponse `json:"medias"`
	CreatedAt  time.Time           `json:"createdAt"`
	UpdatedAt  time.Time           `json:"updatedAt"`
}

type ScanDetailResponse struct {
	ID         string              `json:"id"`
	Status     string              `json:"status"`
	Statuses   []string            `json:"statuses"`
	Message    string              `json:"message"`
	FinalScore float64             `json:"finalScore"`
	Confidence string              `json:"confidence"`
	Verdict    string              `json:"verdict"`
	Duration   int                 `json:"duration"`
	Filename   string              `json:"filename"`
	Thumbnail  string              `json:"thumbnail"`
	RetryCount int                 `json:"retryCount"`
	Insight    *InsightResponse    `json:"insight"`
	Medias     []MediaItemResponse `json:"medias"`
	CreatedAt  time.Time           `json:"createdAt"`
	UpdatedAt  time.Time           `json:"updatedAt"`
}

func primaryMedia(scan *entity.Scan) *entity.Media {
	if scan == nil || len(scan.Medias) == 0 {
		return nil
	}
	return &scan.Medias[0]
}

func NewScanListResponse(scan *entity.Scan) *ScanListResponse {
	response := &ScanListResponse{
		ID:         scan.ID.String(),
		Status:     string(scan.Status),
		Statuses:   scanStatusStrings(scan.Statuses),
		Message:    scan.Message,
		Medias:     NewMediaItemResponses(scan.Medias),
		RetryCount: scan.RetryCount,
		CreatedAt:  scan.CreatedAt,
		UpdatedAt:  scan.UpdatedAt,
		Duration:   scan.Duration,
	}

	if media := primaryMedia(scan); media != nil {
		response.Filename = media.Filename
		response.Thumbnail = thumbnailURL(*media)
	}

	if scan.Verdict != "" {
		response.FinalScore = scan.FinalScore
		response.Confidence = string(scan.Confidence)
		response.Verdict = scan.Verdict
	}

	return response
}

func NewScanDetailResponse(scan *entity.Scan) *ScanDetailResponse {
	response := &ScanDetailResponse{
		ID:         scan.ID.String(),
		Status:     string(scan.Status),
		Statuses:   scanStatusStrings(scan.Statuses),
		Message:    scan.Message,
		Insight:    aggregatedInsight(scan.Medias),
		Medias:     NewMediaItemResponses(scan.Medias),
		CreatedAt:  scan.CreatedAt,
		UpdatedAt:  scan.UpdatedAt,
		Duration:   scan.Duration,
		RetryCount: scan.RetryCount,
	}

	if media := primaryMedia(scan); media != nil {
		response.Filename = media.Filename
		response.Thumbnail = thumbnailURL(*media)
	}

	if scan.Verdict != "" {
		response.FinalScore = scan.FinalScore
		response.Confidence = string(scan.Confidence)
		response.Verdict = scan.Verdict
	}
	return response
}

func scanStatusStrings(statuses []enum.ScanStatus) []string {
	result := make([]string, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, string(status))
	}
	return result
}

func aggregatedInsight(medias []entity.Media) *InsightResponse {
	var (
		count       int
		noise       float64
		compression float64
		frequency   float64
		histogram   float64
	)

	for _, media := range medias {
		if media.Insight == nil {
			continue
		}

		count++
		noise += media.Insight.Noise
		compression += media.Insight.Compression
		frequency += media.Insight.Frequency
		histogram += media.Insight.Histogram
	}

	if count == 0 {
		return nil
	}

	n := float64(count)
	return &InsightResponse{
		Noise:       noise / n,
		Compression: compression / n,
		Frequency:   frequency / n,
		Histogram:   histogram / n,
	}
}

func NewScanListResponses(scans []*entity.Scan) []*ScanListResponse {
	responses := make([]*ScanListResponse, len(scans))
	for i, scan := range scans {
		responses[i] = NewScanListResponse(scan)
	}
	return responses
}
