package subscription

import (
	"context"
	"log"
	"time"

	"go-api/domain/port"
	"go-api/domain/realtime"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

const PaymentNotificationDelay = 10 * time.Second

type Notifier struct {
	userRepo         repository.UserRepository
	subscriptionRepo repository.SubscriptionRepository
	publisher        port.RealtimePublisher
}

func NewNotifier(
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
	publisher port.RealtimePublisher,
) *Notifier {
	return &Notifier{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		publisher:        publisher,
	}
}

func (n *Notifier) Notify(ctx context.Context, subscriptionID uuid.UUID) {
	n.notify(ctx, subscriptionID, realtime.EventSubscriptionUpdated)
}

func (n *Notifier) NotifyPaymentSucceededAfter(subscriptionID uuid.UUID) {
	n.notifyAfter(subscriptionID, realtime.EventPaymentSucceeded, PaymentNotificationDelay)
}

func (n *Notifier) NotifyPaymentFailedAfter(subscriptionID uuid.UUID) {
	n.notifyAfter(subscriptionID, realtime.EventPaymentFailed, PaymentNotificationDelay)
}

func (n *Notifier) notifyAfter(subscriptionID uuid.UUID, eventType string, delay time.Duration) {
	go func() {
		time.Sleep(delay)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		n.notify(ctx, subscriptionID, eventType)
	}()
}

func (n *Notifier) notify(ctx context.Context, subscriptionID uuid.UUID, eventType string) {
	subscription, err := n.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil || subscription == nil {
		log.Printf("subscription notifier: failed to reload subscription %s: %v", subscriptionID, err)
		return
	}

	user, err := n.userRepo.GetBySubscriptionID(ctx, subscriptionID)
	if err != nil || user == nil {
		log.Printf("subscription notifier: no user for subscription %s: %v", subscriptionID, err)
		return
	}

	event := realtime.NewSubscriptionEventWithType(subscription, user.ID, eventType)
	if err := n.publisher.PublishSubscriptionToUser(ctx, user.ID, event); err != nil {
		log.Printf("subscription notifier: failed to publish for user %s: %v", user.ID, err)
	}
}
