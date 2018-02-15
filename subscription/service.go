package subscription

// Service represents service for managing subscriptions.
type Service interface {
	CreateSubscription(s *Subscription) (*Subscription, error)
	GetSubscription(space string, id ID) (*Subscription, error)
	GetSubscriptions(space string) (Subscriptions, error)
	DeleteSubscription(space string, id ID) error
}
