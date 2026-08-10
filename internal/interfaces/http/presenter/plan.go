package presenter

import (
	"time"

	domainplan "go-api/internal/domain/plan"
)

type PlanResponse struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Slug            string         `json:"slug"`
	Price           float64        `json:"price"`
	Currency        string         `json:"currency"`
	BillingInterval string         `json:"billingInterval"`
	IsActive        bool           `json:"isActive"`
	Quota           *QuotaResponse `json:"quota"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

func NewPlanResponse(view *domainplan.PlanView) PlanResponse {
	return PlanResponse{
		ID:              view.ID.String(),
		Name:            view.Name,
		Description:     view.Description,
		Slug:            view.Slug,
		Price:           view.Price,
		Currency:        string(view.Currency),
		BillingInterval: string(view.BillingInterval),
		IsActive:        view.IsActive,
		Quota:           NewQuotaResponse(view.Quota),
		CreatedAt:       view.CreatedAt,
		UpdatedAt:       view.UpdatedAt,
	}
}

func NewPlanResponses(views []*domainplan.PlanView) []PlanResponse {
	responses := make([]PlanResponse, 0, len(views))
	for _, view := range views {
		if view == nil {
			continue
		}
		responses = append(responses, NewPlanResponse(view))
	}
	return responses
}
