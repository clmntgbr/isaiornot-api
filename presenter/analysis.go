package presenter

import (
	"go-api/domain/entity"
	"go-api/domain/enum"
	"time"
)

type GeneratePresignedUploadUrlDetailResponse struct {
	UploadURL  string `json:"uploadUrl,omitempty"`
	AnalysisID string `json:"analysisId"`
}

func NewGeneratePresignedUploadUrlDetailResponse(url string, analysisID string) GeneratePresignedUploadUrlDetailResponse {
	return GeneratePresignedUploadUrlDetailResponse{
		UploadURL:  url,
		AnalysisID: analysisID,
	}
}

type AnalysisListResponse struct {
	ID         string              `json:"id"`
	Status     string              `json:"status"`
	Statuses   []string            `json:"statuses"`
	Message    string              `json:"message,omitempty"`
	FinalScore float64             `json:"finalScore,omitempty"`
	Confidence string              `json:"confidence,omitempty"`
	Verdict    string              `json:"verdict,omitempty"`
	Filename   string              `json:"filename,omitempty"`
	Thumbnail  string              `json:"thumbnail,omitempty"`
	Medias     []MediaItemResponse `json:"medias"`
	CreatedAt  time.Time           `json:"createdAt"`
	UpdatedAt  time.Time           `json:"updatedAt"`
}

type AnalysisDetailResponse struct {
	ID         string              `json:"id"`
	Status     string              `json:"status"`
	Statuses   []string            `json:"statuses"`
	Message    string              `json:"message,omitempty"`
	FinalScore float64             `json:"finalScore,omitempty"`
	Confidence string              `json:"confidence,omitempty"`
	Verdict    string              `json:"verdict,omitempty"`
	Filename   string              `json:"filename,omitempty"`
	Thumbnail  string              `json:"thumbnail,omitempty"`
	Insight    *InsightResponse    `json:"insight,omitempty"`
	Medias     []MediaItemResponse `json:"medias"`
	CreatedAt  time.Time           `json:"createdAt"`
	UpdatedAt  time.Time           `json:"updatedAt"`
}

func primaryMedia(analysis *entity.Analysis) *entity.Media {
	if analysis == nil || len(analysis.Medias) == 0 {
		return nil
	}
	return &analysis.Medias[0]
}

func NewAnalysisListResponse(analysis *entity.Analysis) *AnalysisListResponse {
	response := &AnalysisListResponse{
		ID:        analysis.ID.String(),
		Status:    string(analysis.Status),
		Statuses:  analysisStatusStrings(analysis.Statuses),
		Message:   analysis.Message,
		Medias:    NewMediaItemResponses(analysis.Medias),
		CreatedAt: analysis.CreatedAt,
		UpdatedAt: analysis.UpdatedAt,
	}

	if media := primaryMedia(analysis); media != nil {
		response.Filename = media.Filename
		response.Thumbnail = thumbnailURL(*media)
	}

	if analysis.Verdict != "" {
		response.FinalScore = analysis.FinalScore
		response.Confidence = string(analysis.AnalysisConfidence)
		response.Verdict = analysis.Verdict
	}

	return response
}

func NewAnalysisDetailResponse(analysis *entity.Analysis) *AnalysisDetailResponse {
	response := &AnalysisDetailResponse{
		ID:        analysis.ID.String(),
		Status:    string(analysis.Status),
		Statuses:  analysisStatusStrings(analysis.Statuses),
		Message:   analysis.Message,
		Insight:   aggregatedInsight(analysis.Medias),
		Medias:    NewMediaItemResponses(analysis.Medias),
		CreatedAt: analysis.CreatedAt,
		UpdatedAt: analysis.UpdatedAt,
	}

	if media := primaryMedia(analysis); media != nil {
		response.Filename = media.Filename
		response.Thumbnail = thumbnailURL(*media)
	}

	if analysis.Verdict != "" {
		response.FinalScore = analysis.FinalScore
		response.Confidence = string(analysis.AnalysisConfidence)
		response.Verdict = analysis.Verdict
	}

	return response
}

func analysisStatusStrings(statuses []enum.AnalysisStatus) []string {
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

func NewAnalysisListResponses(analyses []*entity.Analysis) []*AnalysisListResponse {
	responses := make([]*AnalysisListResponse, len(analyses))
	for i, analysis := range analyses {
		responses[i] = NewAnalysisListResponse(analysis)
	}
	return responses
}
