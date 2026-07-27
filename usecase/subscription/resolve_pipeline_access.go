package subscription

import (
	"context"
	"errors"

	"go-api/domain/repository"

	"github.com/google/uuid"
)

type ResolvePipelineAccessUseCase struct {
	userRepo                    repository.UserRepository
	subscriptionRepo            repository.SubscriptionRepository
	resolveEffectivePlanUseCase *ResolveEffectivePlanUseCase
}

func NewResolvePipelineAccessUseCase(
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	resolveEffectivePlanUseCase *ResolveEffectivePlanUseCase,
) *ResolvePipelineAccessUseCase {
	return &ResolvePipelineAccessUseCase{
		userRepo:                    userRepo,
		subscriptionRepo:            subscriptionRepo,
		resolveEffectivePlanUseCase: resolveEffectivePlanUseCase,
	}
}

// FullPipelineForUser reports whether the user's effective plan includes the AI model stage.
func (u *ResolvePipelineAccessUseCase) FullPipelineForUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return false, errors.New("user not found")
	}
	if user.SubscriptionID == nil {
		return false, ErrSubscriptionNotFound
	}

	sub, err := u.subscriptionRepo.GetByID(ctx, *user.SubscriptionID)
	if err != nil || sub == nil {
		return false, ErrSubscriptionNotFound
	}

	plan, err := u.resolveEffectivePlanUseCase.Execute(ctx, sub)
	if err != nil {
		return false, err
	}

	return plan.Quota.FullPipeline, nil
}
