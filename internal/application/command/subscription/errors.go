package subscription

import "errors"

var (
	ErrStripeSubscriptionNotLinked = errors.New("stripe subscription not linked yet")
	ErrMissingStripeCustomer       = errors.New("user has no stripe customer id")
)
