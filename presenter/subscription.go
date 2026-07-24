package presenter

import (
	"go-api/domain/entity"
	"go-api/usecase/subscription"
	"time"
)

type CheckoutSessionResponse struct {
	URL string `json:"url"`
}

func NewCheckoutSessionResponse(url string) CheckoutSessionResponse {
	return CheckoutSessionResponse{URL: url}
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

func NewQuotaUsageResponse(usage *subscription.QuotaUsage) *QuotaUsageResponse {
	if usage == nil {
		return nil
	}
	return &QuotaUsageResponse{
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
	ID                   string              `json:"id"`
	Status               string              `json:"status"`
	StripeCustomerID     string              `json:"stripeCustomerId"`
	StripeSubscriptionID string              `json:"stripeSubscriptionId"`
	StartDate            time.Time           `json:"startDate"`
	EndDate              time.Time           `json:"endDate"`
	CancelAtPeriodEnd    bool                `json:"cancelAtPeriodEnd"`
	QuotaPeriodStart     time.Time           `json:"quotaPeriodStart"`
	Plan                 *PlanResponse       `json:"plan"`
	EffectivePlan        *PlanResponse       `json:"effectivePlan"`
	QuotaUsage           *QuotaUsageResponse `json:"quotaUsage"`
	CreatedAt            time.Time           `json:"createdAt"`
	UpdatedAt            time.Time           `json:"updatedAt"`
}

func NewSubscriptionResponse(
	sub *entity.Subscription,
	effectivePlan *entity.Plan,
	quotaUsage *subscription.QuotaUsage,
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
		QuotaUsage:           NewQuotaUsageResponse(quotaUsage),
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
