package presenter

import (
	"time"

	"go-api/domain/entity"
	"go-api/usecase/subscription"
)

type CheckoutSessionResponse struct {
	URL string `json:"url"`
}

func NewCheckoutSessionResponse(url string) CheckoutSessionResponse {
	return CheckoutSessionResponse{URL: url}
}

type ChangeSubscriptionResponse struct {
	URL     string `json:"url,omitempty"`
	Updated bool   `json:"updated"`
}

func NewChangeSubscriptionResponse(result *subscription.ChangeSubscriptionResult) ChangeSubscriptionResponse {
	if result == nil {
		return ChangeSubscriptionResponse{}
	}
	return ChangeSubscriptionResponse{
		URL:     result.URL,
		Updated: result.Updated,
	}
}

type ProrationPreviewLineResponse struct {
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Proration   bool   `json:"proration"`
}

type PlanChangePreviewResponse struct {
	RequiresCheckout bool                           `json:"requiresCheckout"`
	Currency         string                         `json:"currency"`
	AmountDue        int64                          `json:"amountDue"`
	Subtotal         int64                          `json:"subtotal"`
	Total            int64                          `json:"total"`
	ProrationDate    int64                          `json:"prorationDate,omitempty"`
	PeriodStart      time.Time                      `json:"periodStart"`
	PeriodEnd        time.Time                      `json:"periodEnd"`
	Lines            []ProrationPreviewLineResponse `json:"lines"`
	CurrentPlanID    string                         `json:"currentPlanId,omitempty"`
	CurrentPlanSlug  string                         `json:"currentPlanSlug,omitempty"`
	TargetPlanID     string                         `json:"targetPlanId"`
	TargetPlanSlug   string                         `json:"targetPlanSlug"`
	TargetPlanName   string                         `json:"targetPlanName"`
	TargetPlanPrice  float64                        `json:"targetPlanPrice"`
}

func NewPlanChangePreviewResponse(preview *subscription.PlanChangePreview) PlanChangePreviewResponse {
	if preview == nil {
		return PlanChangePreviewResponse{Lines: []ProrationPreviewLineResponse{}}
	}

	lines := make([]ProrationPreviewLineResponse, 0, len(preview.Lines))
	for _, line := range preview.Lines {
		lines = append(lines, ProrationPreviewLineResponse{
			Description: line.Description,
			Amount:      line.Amount,
			Proration:   line.Proration,
		})
	}

	return PlanChangePreviewResponse{
		RequiresCheckout: preview.RequiresCheckout,
		Currency:         preview.Currency,
		AmountDue:        preview.AmountDue,
		Subtotal:         preview.Subtotal,
		Total:            preview.Total,
		ProrationDate:    preview.ProrationDate,
		PeriodStart:      preview.PeriodStart,
		PeriodEnd:        preview.PeriodEnd,
		Lines:            lines,
		CurrentPlanID:    preview.CurrentPlanID,
		CurrentPlanSlug:  preview.CurrentPlanSlug,
		TargetPlanID:     preview.TargetPlanID,
		TargetPlanSlug:   preview.TargetPlanSlug,
		TargetPlanName:   preview.TargetPlanName,
		TargetPlanPrice:  preview.TargetPlanPrice,
	}
}

type QuotaUsageResponse struct {
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`

	ImagesUsed int64 `json:"imagesUsed"`
	ImagesMax  int   `json:"imagesMax"`
	ImagesLeft int64 `json:"imagesLeft"`
	VideosUsed int64 `json:"videosUsed"`
	VideosMax  int   `json:"videosMax"`
	VideosLeft int64 `json:"videosLeft"`

	MaxFileSizeImage int64 `json:"maxFileSizeImage"`
	MaxFileSizeVideo int64 `json:"maxFileSizeVideo"`
	FullPipeline     bool  `json:"fullPipeline"`
}

func NewQuotaUsageResponse(usage *subscription.QuotaUsage) QuotaUsageResponse {
	return QuotaUsageResponse{
		PeriodStart:      usage.PeriodStart,
		PeriodEnd:        usage.PeriodEnd,
		ImagesUsed:       usage.ImagesUsed,
		ImagesMax:        usage.ImagesMax,
		ImagesLeft:       usage.ImagesLeft,
		VideosUsed:       usage.VideosUsed,
		VideosMax:        usage.VideosMax,
		VideosLeft:       usage.VideosLeft,
		MaxFileSizeImage: usage.MaxFileSizeImage,
		MaxFileSizeVideo: usage.MaxFileSizeVideo,
		FullPipeline:     usage.FullPipeline,
	}
}

type SubscriptionResponse struct {
	ID                   string        `json:"id"`
	Status               string        `json:"status"`
	StripeCustomerID     string        `json:"stripeCustomerId"`
	StripeSubscriptionID string        `json:"stripeSubscriptionId"`
	StartDate            time.Time     `json:"startDate"`
	EndDate              time.Time     `json:"endDate"`
	CancelAtPeriodEnd    bool          `json:"cancelAtPeriodEnd"`
	QuotaPeriodStart     time.Time     `json:"quotaPeriodStart"`
	Plan                 *PlanResponse `json:"plan"`
	EffectivePlan        *PlanResponse `json:"effectivePlan"`
	CreatedAt            time.Time     `json:"createdAt"`
	UpdatedAt            time.Time     `json:"updatedAt"`
}

func NewSubscriptionResponse(
	sub *entity.Subscription,
	effectivePlan *entity.Plan,
) *SubscriptionResponse {
	response := &SubscriptionResponse{
		ID:                   sub.ID.String(),
		Status:               string(sub.SubscriptionStatus),
		StripeCustomerID:     sub.StripeCustomerID,
		StripeSubscriptionID: sub.StripeSubscriptionID,
		StartDate:            sub.SubscriptionStartDate,
		EndDate:              sub.SubscriptionEndDate,
		CancelAtPeriodEnd:    sub.CancelAtPeriodEnd,
		QuotaPeriodStart:     sub.QuotaPeriodStart,
		Plan:                 NewPlanResponse(&sub.Plan),
		CreatedAt:            sub.CreatedAt,
		UpdatedAt:            sub.UpdatedAt,
	}

	if effectivePlan != nil {
		response.EffectivePlan = NewPlanResponse(effectivePlan)
	} else {
		response.EffectivePlan = response.Plan
	}

	return response
}
