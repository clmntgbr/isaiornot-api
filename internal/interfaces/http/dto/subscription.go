package dto

type PreviewSubscriptionRequest struct {
	PlanID string `json:"planId"`
}

type CreateSubscriptionRequest struct {
	PlanID        string `json:"planId"`
	ProrationDate *int64 `json:"prorationDate"`
}
