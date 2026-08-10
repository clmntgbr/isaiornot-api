package subscription

import (
	"context"
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	domainmedia "go-api/internal/domain/media"

	"github.com/google/uuid"
)

var (
	ErrImageQuotaExceeded = errors.New("monthly image quota exceeded for your current plan")
	ErrVideoQuotaExceeded = errors.New("monthly video quota exceeded for your current plan")
	ErrImagesNotAllowed   = errors.New("your current plan does not allow image uploads")
	ErrVideosNotAllowed   = errors.New("your current plan does not allow video uploads")
	ErrImageFileTooLarge  = errors.New("image file exceeds the size limit of your current plan")
	ErrVideoFileTooLarge  = errors.New("video file exceeds the size limit of your current plan")
)

type AssertUploadAllowedCommand struct {
	UserID              uuid.UUID
	ContentType         string
	Size                int64
	MediaAlreadyCounted bool
}

type AssertUploadAllowedHandler struct {
	getQuotaUsage *querysubscription.GetQuotaUsageHandler
}

func NewAssertUploadAllowedHandler(
	getQuotaUsage *querysubscription.GetQuotaUsageHandler,
) *AssertUploadAllowedHandler {
	return &AssertUploadAllowedHandler{getQuotaUsage: getQuotaUsage}
}

func (h *AssertUploadAllowedHandler) Handle(ctx context.Context, cmd AssertUploadAllowedCommand) error {
	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{UserID: cmd.UserID})
	if err != nil {
		return err
	}

	switch {
	case domainmedia.IsVideoContentType(cmd.ContentType):
		return assertVideoAllowed(usage, cmd)
	case domainmedia.IsImageContentType(cmd.ContentType):
		return assertImageAllowed(usage, cmd)
	default:
		return nil
	}
}

func assertVideoAllowed(usage *querysubscription.QuotaUsageView, cmd AssertUploadAllowedCommand) error {
	if usage.VideosMax <= 0 {
		return ErrVideosNotAllowed
	}
	if cmd.MediaAlreadyCounted {
		if usage.VideosUsed > int64(usage.VideosMax) {
			return ErrVideoQuotaExceeded
		}
	} else if usage.VideosLeft <= 0 {
		return ErrVideoQuotaExceeded
	}
	if cmd.Size > 0 && usage.MaxFileSizeVideo > 0 && cmd.Size > usage.MaxFileSizeVideo {
		return ErrVideoFileTooLarge
	}
	return nil
}

func assertImageAllowed(usage *querysubscription.QuotaUsageView, cmd AssertUploadAllowedCommand) error {
	if usage.ImagesMax <= 0 {
		return ErrImagesNotAllowed
	}
	if cmd.MediaAlreadyCounted {
		if usage.ImagesUsed > int64(usage.ImagesMax) {
			return ErrImageQuotaExceeded
		}
	} else if usage.ImagesLeft <= 0 {
		return ErrImageQuotaExceeded
	}
	if cmd.Size > 0 && usage.MaxFileSizeImage > 0 && cmd.Size > usage.MaxFileSizeImage {
		return ErrImageFileTooLarge
	}
	return nil
}
