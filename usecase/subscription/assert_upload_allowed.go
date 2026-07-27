package subscription

import (
	"context"
	"errors"
	"go-api/domain/repository"
	mediadto "go-api/domain/media"

	"github.com/google/uuid"
)

var (
	ErrImageQuotaExceeded = errors.New("Monthly image quota exceeded for your current plan")
	ErrVideoQuotaExceeded = errors.New("Monthly video quota exceeded for your current plan")
	ErrImagesNotAllowed   = errors.New("Your current plan does not allow image uploads")
	ErrVideosNotAllowed   = errors.New("Your current plan does not allow video uploads")
	ErrImageFileTooLarge  = errors.New("Image file exceeds the size limit of your current plan")
	ErrVideoFileTooLarge  = errors.New("Video file exceeds the size limit of your current plan")
)

type AssertUploadAllowedUseCase struct {
	userRepo                    repository.UserRepository
	subscriptionRepo            repository.SubscriptionRepository
	resolveEffectivePlanUseCase *ResolveEffectivePlanUseCase
	getQuotaUsageUseCase        *GetQuotaUsageUseCase
}

func NewAssertUploadAllowedUseCase(
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	resolveEffectivePlanUseCase *ResolveEffectivePlanUseCase,
	getQuotaUsageUseCase *GetQuotaUsageUseCase,
) *AssertUploadAllowedUseCase {
	return &AssertUploadAllowedUseCase{
		userRepo:                    userRepo,
		subscriptionRepo:            subscriptionRepo,
		resolveEffectivePlanUseCase: resolveEffectivePlanUseCase,
		getQuotaUsageUseCase:        getQuotaUsageUseCase,
	}
}

type AssertUploadAllowedInput struct {
	UserID              uuid.UUID
	ContentType         string
	Size                int64
	MediaAlreadyCounted bool
}

func (u *AssertUploadAllowedUseCase) Execute(ctx context.Context, input AssertUploadAllowedInput) error {
	user, err := u.userRepo.GetByID(ctx, input.UserID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}
	if user.SubscriptionID == nil {
		return ErrSubscriptionNotFound
	}

	subscription, err := u.subscriptionRepo.GetByID(ctx, *user.SubscriptionID)
	if err != nil || subscription == nil {
		return ErrSubscriptionNotFound
	}

	effectivePlan, err := u.resolveEffectivePlanUseCase.Execute(ctx, subscription)
	if err != nil {
		return err
	}

	usage, err := u.getQuotaUsageUseCase.Execute(ctx, user, subscription, effectivePlan)
	if err != nil {
		return err
	}

	switch {
	case mediadto.IsVideoContentType(input.ContentType):
		return assertVideoAllowed(usage, input)
	case mediadto.IsImageContentType(input.ContentType):
		return assertImageAllowed(usage, input)
	default:
		return nil
	}
}

func assertVideoAllowed(usage *QuotaUsage, input AssertUploadAllowedInput) error {
	if usage.VideosMax <= 0 {
		return ErrVideosNotAllowed
	}
	if input.MediaAlreadyCounted {
		if usage.VideosUsed > int64(usage.VideosMax) {
			return ErrVideoQuotaExceeded
		}
	} else if usage.VideosLeft <= 0 {
		return ErrVideoQuotaExceeded
	}
	if input.Size > 0 && usage.MaxFileSizeVideo > 0 && input.Size > usage.MaxFileSizeVideo {
		return ErrVideoFileTooLarge
	}
	return nil
}

func assertImageAllowed(usage *QuotaUsage, input AssertUploadAllowedInput) error {
	if usage.ImagesMax <= 0 {
		return ErrImagesNotAllowed
	}
	if input.MediaAlreadyCounted {
		if usage.ImagesUsed > int64(usage.ImagesMax) {
			return ErrImageQuotaExceeded
		}
	} else if usage.ImagesLeft <= 0 {
		return ErrImageQuotaExceeded
	}
	if input.Size > 0 && usage.MaxFileSizeImage > 0 && input.Size > usage.MaxFileSizeImage {
		return ErrImageFileTooLarge
	}
	return nil
}
