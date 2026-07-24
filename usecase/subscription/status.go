package subscription

import "go-api/domain/entity"

func MapStripeStatus(stripeStatus string) entity.SubscriptionStatus {
	switch stripeStatus {
	case "active", "trialing":
		return entity.SubscriptionStatusActive
	case "past_due":
		return entity.SubscriptionStatusPastDue
	case "unpaid":
		return entity.SubscriptionStatusPastDue
	case "canceled":
		return entity.SubscriptionStatusCancelled
	case "incomplete", "incomplete_expired", "paused":
		return entity.SubscriptionStatusInactive
	default:
		return entity.SubscriptionStatusInactive
	}
}
