package subscription

import (
	"context"
	"go-api/domain/repository"
	"go-api/infrastructure/centrifugo"
	"log"
	"time"

	"github.com/google/uuid"
)

// PaymentNotificationDelay is how long we wait before pushing payment
// success/failure notifications, so the client has time to land on the
// redirect page before receiving the realtime event.
const PaymentNotificationDelay = 10 * time.Second

// Notifier publishes subscription changes to the owning user over Centrifugo.
// It is best-effort: failures are logged and never abort the caller.
type Notifier struct {
	userRepo         *repository.UserRepository
	subscriptionRepo *repository.SubscriptionRepository
	publisher        *centrifugo.Publisher
}

func NewNotifier(
	userRepo *repository.UserRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	publisher *centrifugo.Publisher,
) *Notifier {
	return &Notifier{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		publisher:        publisher,
	}
}

func (n *Notifier) Notify(ctx context.Context, subscriptionID uuid.UUID) {
	n.notify(ctx, subscriptionID, centrifugo.EventSubscriptionUpdated)
}

// NotifyPaymentSucceededAfter publishes the payment success event after the
// configured delay, in a background goroutine. It never blocks the caller.
func (n *Notifier) NotifyPaymentSucceededAfter(subscriptionID uuid.UUID) {
	n.notifyAfter(subscriptionID, centrifugo.EventPaymentSucceeded, PaymentNotificationDelay)
}

// NotifyPaymentFailedAfter publishes the payment failure event after the
// configured delay, in a background goroutine. It never blocks the caller.
func (n *Notifier) NotifyPaymentFailedAfter(subscriptionID uuid.UUID) {
	n.notifyAfter(subscriptionID, centrifugo.EventPaymentFailed, PaymentNotificationDelay)
}

func (n *Notifier) notifyAfter(subscriptionID uuid.UUID, eventType string, delay time.Duration) {
	go func() {
		time.Sleep(delay)

		log.Printf("subscription notifier: sending delayed %s for subscription %s (after %s)", eventType, subscriptionID, delay)

		// The request context is already cancelled at this point, so use a
		// fresh, bounded background context for the delayed publish.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		n.notify(ctx, subscriptionID, eventType)
	}()
}

func (n *Notifier) notify(ctx context.Context, subscriptionID uuid.UUID, eventType string) {
	subscription, err := (*n.subscriptionRepo).GetByID(ctx, subscriptionID)
	if err != nil || subscription == nil {
		log.Printf("subscription notifier: failed to reload subscription %s: %v", subscriptionID, err)
		return
	}

	user, err := (*n.userRepo).GetBySubscriptionID(ctx, subscriptionID)
	if err != nil || user == nil {
		log.Printf("subscription notifier: no user for subscription %s: %v", subscriptionID, err)
		return
	}

	event := centrifugo.NewSubscriptionEventWithType(subscription, user.ID, eventType)
	if err := n.publisher.PublishSubscriptionToUser(ctx, user.ID, event); err != nil {
		log.Printf("subscription notifier: failed to publish for user %s: %v", user.ID, err)
	}
}
