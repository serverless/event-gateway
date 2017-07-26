package pubsub

import (
	"fmt"
	"net/url"
	"path"

	"github.com/serverless/event-gateway/functions"
	validator "gopkg.in/go-playground/validator.v9"
)

// SubscriptionID uniquely identifies a subscription
type SubscriptionID string

// EventHTTP represents sync HTTP subscription
const EventHTTP = "http"

// Subscription maps from Topic to Function
type Subscription struct {
	ID         SubscriptionID       `json:"subscriptionId"`
	Event      TopicID              `json:"event" validate:"required,alphanum"`
	FunctionID functions.FunctionID `json:"functionId" validate:"required"`
	Method     string               `json:"method,omitempty" validate:"omitempty,eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
	Path       string               `json:"path,omitempty" validate:"omitempty,urlpath"`
}

func newSubscriptionID(s *Subscription) SubscriptionID {
	if s.Method == "" && s.Path == "" {
		return SubscriptionID(string(s.Event) + "-" + string(s.FunctionID))
	}
	return SubscriptionID(string(s.Event) + "-" + s.Method + "-" + url.PathEscape(s.Path))
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
	original string
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

// urlPathValidator validates if field contains URL path
func urlPathValidator(fl validator.FieldLevel) bool {
	return path.IsAbs(fl.Field().String())
}
