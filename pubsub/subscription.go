package pubsub

import (
	"fmt"

	"github.com/serverless/event-gateway/functions"
)

// SubscriptionID uniquely identifies a subscription
type SubscriptionID string

// Subscription maps from Topic to Function
type Subscription struct {
	ID         SubscriptionID       `json:"subscriptionId"`
	TopicID    TopicID              `json:"event" validate:"required"`
	FunctionID functions.FunctionID `json:"functionId" validate:"required"`
	Method     string               `json:"method,omitempty" validate:"omitempty,eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
	Path       string               `json:"path,omitempty"`
}

func newSubscriptionID(s *Subscription) SubscriptionID {
	if s.Method == "" && s.Path == "" {
		return SubscriptionID(string(s.TopicID) + "-" + string(s.FunctionID))
	}
	return SubscriptionID(string(s.TopicID) + "-" + string(s.FunctionID) + "-" + s.Method + "-" + s.Path)
}

// ErrorSubscriptionAlreadyExists occurs when subscription with the same ID already exists.
type ErrorSubscriptionAlreadyExists struct {
	ID SubscriptionID
}

func (e ErrorSubscriptionAlreadyExists) Error() string {
	return fmt.Sprintf("Subscription %q already exits.", e.ID)
}

// ErrorSubscriptionValidation occurs when subscription payload doesn't validate.
type ErrorSubscriptionValidation struct {
	original error
}

func (e ErrorSubscriptionValidation) Error() string {
	return fmt.Sprintf("Subscription doesn't validate. Validation error: %q", e.original)
}

// ErrorSubscriptionNotFound occurs when subscription cannot be found.
type ErrorSubscriptionNotFound struct {
	ID SubscriptionID
}

func (e ErrorSubscriptionNotFound) Error() string {
	return fmt.Sprintf("Subscription %q not found.", e.ID)
}

// ErrorFunctionNotFound occurs when subscription cannot be created because backing function doesn't exist.
type ErrorFunctionNotFound struct {
	functionID string
}

func (e ErrorFunctionNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", e.functionID)
}
