package presenter

import (
	"time"

	domainscan "go-api/internal/domain/scan"
)

type ScanResponse struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	Statuses   []string  `json:"statuses"`
	Message    string    `json:"message"`
	FinalScore float64   `json:"finalScore"`
	Confidence string    `json:"confidence"`
	Verdict    string    `json:"verdict"`
	Duration   int       `json:"duration"`
	RetryCount int       `json:"retryCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func NewScanResponse(view *domainscan.ScanView) ScanResponse {
	statuses := make([]string, 0, len(view.Statuses))
	for _, status := range view.Statuses {
		statuses = append(statuses, string(status))
	}

	return ScanResponse{
		ID:         view.ID.String(),
		Status:     string(view.Status),
		Statuses:   statuses,
		Message:    view.Message,
		FinalScore: view.FinalScore,
		Confidence: string(view.Confidence),
		Verdict:    view.Verdict,
		Duration:   view.Duration,
		RetryCount: view.RetryCount,
		CreatedAt:  view.CreatedAt,
		UpdatedAt:  view.UpdatedAt,
	}
}

func NewScanResponses(views []*domainscan.ScanView) []ScanResponse {
	responses := make([]ScanResponse, 0, len(views))
	for _, view := range views {
		if view == nil {
			continue
		}
		responses = append(responses, NewScanResponse(view))
	}
	return responses
}
