package dto

type CreateSubscriptionRequest struct {
	PlanID        string `json:"planId"`
	ProrationDate *int64 `json:"prorationDate"`
}

type PreviewSubscriptionRequest struct {
	PlanID string `json:"planId"`
}
