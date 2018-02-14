package subscription

// Service represents service for managing subscriptions.
type Service interface {
	CreateSubscription(s *Subscription) (*Subscription, error)
	DeleteSubscription(space string, id ID) error
	GetSubscriptions(space string) (Subscriptions, error)
}
