package subscription

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"time"
)

type GetQuotaUsageUseCase struct {
	mediaRepo *repository.MediaRepository
}

func NewGetQuotaUsageUseCase(
	mediaRepo *repository.MediaRepository,
) *GetQuotaUsageUseCase {
	return &GetQuotaUsageUseCase{
		mediaRepo: mediaRepo,
	}
}

type QuotaUsage struct {
	PeriodStart time.Time
	PeriodEnd   time.Time

	ImagesUsed int64
	ImagesMax  int
	ImagesLeft int64
	VideosUsed int64
	VideosMax  int
	VideosLeft int64

	MaxFileSizeImage int64
	MaxFileSizeVideo int64
	FullPipeline     bool
}

func (u *GetQuotaUsageUseCase) Execute(
	ctx context.Context,
	user *entity.User,
	subscription *entity.Subscription,
	effectivePlan *entity.Plan,
) (*QuotaUsage, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}
	if subscription == nil {
		return nil, errors.New("subscription is required")
	}
	if effectivePlan == nil {
		return nil, errors.New("effective plan is required")
	}

	anchor := subscription.QuotaPeriodStart
	if anchor.IsZero() {
		anchor = subscription.SubscriptionStartDate
	}
	if anchor.IsZero() {
		anchor = subscription.CreatedAt
	}

	periodStart, periodEnd := CurrentQuotaPeriod(anchor, time.Now().UTC())

	counts, err := (*u.mediaRepo).CountUsageInPeriod(ctx, user.ID, periodStart, periodEnd)
	if err != nil {
		return nil, errors.New("failed to count media usage")
	}

	quota := effectivePlan.Quota
	imagesLeft := int64(quota.MaxImagesPerMonth) - counts.Images
	if imagesLeft < 0 {
		imagesLeft = 0
	}
	videosLeft := int64(quota.MaxVideosPerMonth) - counts.Videos
	if videosLeft < 0 {
		videosLeft = 0
	}

	return &QuotaUsage{
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
