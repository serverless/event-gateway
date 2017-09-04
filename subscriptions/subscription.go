package subscriptions

import (
	"fmt"
	"net/url"
	"path"
	"regexp"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	validator "gopkg.in/go-playground/validator.v9"
)

// SubscriptionID uniquely identifies a subscription
type SubscriptionID string

// SubscriptionHTTP is a special type of subscription. It represents sync HTTP subscription.
const SubscriptionHTTP = "http"

// Subscription maps from Event type to Function
type Subscription struct {
	ID         SubscriptionID       `json:"subscriptionId"`
	Event      event.Type           `json:"event" validate:"required,eventtype"`
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

// ErrSubscriptionAlreadyExists occurs when subscription with the same ID already exists.
type ErrSubscriptionAlreadyExists struct {
	ID SubscriptionID
}

func (e ErrSubscriptionAlreadyExists) Error() string {
	return fmt.Sprintf("Subscription %q already exits.", e.ID)
}

// ErrSubscriptionValidation occurs when subscription payload doesn't validate.
type ErrSubscriptionValidation struct {
	original string
}

func (e ErrSubscriptionValidation) Error() string {
	return fmt.Sprintf("Subscription doesn't validate. Validation error: %q", e.original)
}

// ErrSubscriptionNotFound occurs when subscription cannot be found.
type ErrSubscriptionNotFound struct {
	ID SubscriptionID
}

func (e ErrSubscriptionNotFound) Error() string {
	return fmt.Sprintf("Subscription %q not found.", e.ID)
}

// ErrFunctionNotFound occurs when subscription cannot be created because backing function doesn't exist.
type ErrFunctionNotFound struct {
	functionID string
}

func (e ErrFunctionNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", e.functionID)
}

// urlPathValidator validates if field contains URL path
func urlPathValidator(fl validator.FieldLevel) bool {
	return path.IsAbs(fl.Field().String())
}

// eventTypeValidator validates if field contains event name
func eventTypeValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`).MatchString(fl.Field().String())
}
