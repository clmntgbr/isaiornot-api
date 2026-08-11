package presenter

import (
	"time"

	domainscan "go-api/internal/domain/scan"
)

type ScanResponse struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Status     string          `json:"status"`
	Statuses   []string        `json:"statuses"`
	Message    *string         `json:"message"`
	FinalScore float64         `json:"finalScore"`
	Confidence *string         `json:"confidence"`
	Verdict    *string         `json:"verdict"`
	Duration   int             `json:"duration"`
	RetryCount int             `json:"retryCount"`
	Medias     []MediaResponse `json:"medias"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

func NewScanResponse(view *domainscan.ScanView) ScanResponse {
	statuses := make([]string, 0, len(view.Statuses))
	for _, status := range view.Statuses {
		statuses = append(statuses, string(status))
	}

	medias := NewMediaResponses(view.Medias)
	if medias == nil {
		medias = []MediaResponse{}
	}

	return ScanResponse{
		ID:         view.ID.String(),
		Type:       string(view.Type),
		Status:     string(view.Status),
		Statuses:   statuses,
		Message:    optionalString(view.Message),
		FinalScore: view.FinalScore,
		Confidence: optionalString(string(view.Confidence)),
		Verdict:    optionalString(view.Verdict),
		Duration:   view.Duration,
		RetryCount: view.RetryCount,
		Medias:     medias,
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

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
