package api

// Subscription maps event type to a function.
type Subscription struct {
	ID         SubscriptionID `json:"subscriptionId"`
	Event      EventType      `json:"event" validate:"required,eventtype"`
	FunctionID FunctionID     `json:"functionId" validate:"required"`
	Method     string         `json:"method,omitempty" validate:"omitempty,eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
	Path       string         `json:"path,omitempty" validate:"omitempty,urlpath"`
	CORS       *CORS          `json:"cors,omitempty"`
}

// SubscriptionID uniquely identifies a subscription.
type SubscriptionID string

// CORS is used to configure CORS on HTTP subscriptions.
type CORS struct {
	Origins          []string `json:"origins" validate:"min=1"`
	Methods          []string `json:"methods" validate:"min=1"`
	Headers          []string `json:"headers" validate:"min=1"`
	AllowCredentials bool     `json:"allowCredentials"`
}

// SubscriptionService represents service for managing subscriptions.
type SubscriptionService interface {
	CreateSubscription(s *Subscription) (*Subscription, error)
	DeleteSubscription(id SubscriptionID) error
	GetAllSubscriptions() ([]*Subscription, error)
}
