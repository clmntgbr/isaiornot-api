package scan

import (
	"context"
	"errors"
	"time"

	querysubscription "go-api/internal/application/query/subscription"

	"github.com/google/uuid"
)

var ErrHistoryOutsideRetention = errors.New("scan is outside history retention")

type HistoryCutoffResolver struct {
	getSubscription *querysubscription.GetCurrentSubscriptionHandler
}

func NewHistoryCutoffResolver(
	getSubscription *querysubscription.GetCurrentSubscriptionHandler,
) *HistoryCutoffResolver {
	return &HistoryCutoffResolver{getSubscription: getSubscription}
}

func (r *HistoryCutoffResolver) ForUser(ctx context.Context, userID uuid.UUID) (time.Time, error) {
	sub, err := r.getSubscription.Handle(ctx, querysubscription.GetCurrentSubscriptionQuery{UserID: userID})
	if err != nil {
		return time.Time{}, err
	}
	if sub.Plan == nil || sub.Plan.Quota == nil {
		return time.Time{}, errors.New("subscription plan quota not found")
	}

	retention := sub.Plan.Quota.HistoryRetention
	if retention <= 0 {
		return time.Time{}, nil
	}

	return time.Now().UTC().Add(-retention), nil
}
