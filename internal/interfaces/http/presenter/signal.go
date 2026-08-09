package presenter

import (
	"time"

	domainsignal "go-api/internal/domain/signal"
)

type SignalResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Score      int       `json:"score"`
	Confidence string    `json:"confidence"`
	Details    []string  `json:"details"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func NewSignalResponse(view *domainsignal.SignalView) SignalResponse {
	details := view.Details
	if details == nil {
		details = []string{}
	}

	return SignalResponse{
		ID:         view.ID.String(),
		Name:       view.Name,
		Score:      view.Score,
		Confidence: string(view.Confidence),
		Details:    details,
		CreatedAt:  view.CreatedAt,
		UpdatedAt:  view.UpdatedAt,
	}
}

func NewSignalResponses(views []*domainsignal.SignalView) []SignalResponse {
	responses := make([]SignalResponse, 0, len(views))
	for _, view := range views {
		if view == nil {
			continue
		}
		responses = append(responses, NewSignalResponse(view))
	}
	return responses
}
