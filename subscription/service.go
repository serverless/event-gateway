package subscription

// SubscriptionService represents service for managing subscriptions.
type Service interface {
	CreateSubscription(s *Subscription) (*Subscription, error)
	DeleteSubscription(id ID) error
	GetAllSubscriptions() ([]*Subscription, error)
}
