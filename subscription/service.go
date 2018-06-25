package subscription

// Service represents service for managing subscriptions.
type Service interface {
	GetSubscription(space string, id ID) (*Subscription, error)
	ListSubscriptions(space string) (Subscriptions, error)
	CreateSubscription(s *Subscription) (*Subscription, error)
	UpdateSubscription(id ID, s *Subscription) (*Subscription, error)
	DeleteSubscription(space string, id ID) error
}
