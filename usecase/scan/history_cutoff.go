package scan

import (
	"context"
	"errors"
	"time"

	"go-api/domain/repository"
	"go-api/usecase/subscription"

	"github.com/google/uuid"
)

var ErrHistoryOutsideRetention = errors.New("scan is outside history retention")

type HistoryCutoffResolver struct {
	userRepo                    repository.UserRepository
	subscriptionRepo            repository.SubscriptionRepository
	resolveEffectivePlanUseCase *subscription.ResolveEffectivePlanUseCase
}

func NewHistoryCutoffResolver(
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	resolveEffectivePlanUseCase *subscription.ResolveEffectivePlanUseCase,
) *HistoryCutoffResolver {
	return &HistoryCutoffResolver{
		userRepo:                    userRepo,
		subscriptionRepo:            subscriptionRepo,
		resolveEffectivePlanUseCase: resolveEffectivePlanUseCase,
	}
}

// ForUser returns the earliest created_at included in history/statistics for the user's effective plan.
// A zero time means no retention limit.
func (r *HistoryCutoffResolver) ForUser(ctx context.Context, userID uuid.UUID) (time.Time, error) {
	user, err := r.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return time.Time{}, errors.New("user not found")
	}
	if user.SubscriptionID == nil {
		return time.Time{}, subscription.ErrSubscriptionNotFound
	}

	sub, err := r.subscriptionRepo.GetByID(ctx, *user.SubscriptionID)
	if err != nil || sub == nil {
		return time.Time{}, subscription.ErrSubscriptionNotFound
	}

	plan, err := r.resolveEffectivePlanUseCase.Execute(ctx, sub)
	if err != nil {
		return time.Time{}, err
	}

	retention := plan.Quota.HistoryRetention
	if retention <= 0 {
		return time.Time{}, nil
	}

	return time.Now().UTC().Add(-retention), nil
}
