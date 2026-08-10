package subscription

import (
	"context"
	"errors"
	"time"

	domainmedia "go-api/internal/domain/media"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type GetQuotaUsageQuery struct {
	UserID uuid.UUID
}

type QuotaUsageView struct {
	PeriodStart      time.Time
	PeriodEnd        time.Time
	ImagesUsed       int64
	ImagesMax        int
	ImagesLeft       int64
	VideosUsed       int64
	VideosMax        int
	VideosLeft       int64
	MaxFileSizeImage int64
	MaxFileSizeVideo int64
	FullPipeline     bool
}

type GetQuotaUsageHandler struct {
	userRepo         domainuser.UserReadRepository
	subscriptionRepo domainsubscription.SubscriptionReadRepository
	mediaRepo        domainmedia.MediaReadRepository
}

func NewGetQuotaUsageHandler(
	userRepo domainuser.UserReadRepository,
	subscriptionRepo domainsubscription.SubscriptionReadRepository,
	mediaRepo domainmedia.MediaReadRepository,
) *GetQuotaUsageHandler {
	return &GetQuotaUsageHandler{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		mediaRepo:        mediaRepo,
	}
}

func (h *GetQuotaUsageHandler) Handle(ctx context.Context, q GetQuotaUsageQuery) (*QuotaUsageView, error) {
	user, err := h.userRepo.FindByID(ctx, q.UserID)
	if err != nil {
		return nil, errors.New("failed to get user")
	}
	if user == nil || user.SubscriptionID == nil {
		return nil, ErrSubscriptionNotFound
	}

	sub, err := h.subscriptionRepo.FindByID(ctx, *user.SubscriptionID)
	if err != nil {
		return nil, errors.New("failed to get subscription")
	}
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	if sub.Plan == nil || sub.Plan.Quota == nil {
		return nil, errors.New("subscription plan quota not found")
	}

	anchor := sub.QuotaPeriodStart
	if anchor.IsZero() {
		anchor = sub.StartDate
	}
	if anchor.IsZero() {
		anchor = sub.CreatedAt
	}

	periodStart, periodEnd := domainsubscription.CurrentQuotaPeriod(anchor, time.Now().UTC())

	counts, err := h.mediaRepo.CountUsageInPeriod(ctx, q.UserID, periodStart, periodEnd)
	if err != nil {
		return nil, errors.New("failed to count media usage")
	}

	quota := sub.Plan.Quota
	imagesLeft := int64(quota.MaxImagesPerMonth) - counts.Images
	if imagesLeft < 0 {
		imagesLeft = 0
	}
	videosLeft := int64(quota.MaxVideosPerMonth) - counts.Videos
	if videosLeft < 0 {
		videosLeft = 0
	}

	return &QuotaUsageView{
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		ImagesUsed:       counts.Images,
		ImagesMax:        quota.MaxImagesPerMonth,
		ImagesLeft:       imagesLeft,
		VideosUsed:       counts.Videos,
		VideosMax:        quota.MaxVideosPerMonth,
		VideosLeft:       videosLeft,
		MaxFileSizeImage: quota.MaxFileSizeImage,
		MaxFileSizeVideo: quota.MaxFileSizeVideo,
		FullPipeline:     quota.FullPipeline,
	}, nil
}
