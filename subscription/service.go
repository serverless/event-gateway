package subscription

import "github.com/serverless/event-gateway/metadata"

// Service represents service for managing subscriptions.
type Service interface {
	GetSubscription(space string, id ID) (*Subscription, error)
	ListSubscriptions(space string, filters ...metadata.Filter) (Subscriptions, error)
	CreateSubscription(s *Subscription) (*Subscription, error)
	UpdateSubscription(id ID, s *Subscription) (*Subscription, error)
	DeleteSubscription(space string, id ID) error
}
