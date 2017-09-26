package subscriptions

import (
	"net/url"
	"path"
	"regexp"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	validator "gopkg.in/go-playground/validator.v9"
)

// SubscriptionID uniquely identifies a subscription
type SubscriptionID string

// Subscription maps from Event type to Function
type Subscription struct {
	ID         SubscriptionID       `json:"subscriptionId"`
	Event      event.Type           `json:"event" validate:"required,eventtype"`
	FunctionID functions.FunctionID `json:"functionId" validate:"required"`
	Method     string               `json:"method,omitempty" validate:"omitempty,eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
	Path       string               `json:"path,omitempty" validate:"omitempty,urlpath"`
}

func newSubscriptionID(s *Subscription) SubscriptionID {
	if s.Event == event.TypeHTTP {
		return SubscriptionID(string(s.Event) + "," + s.Method + "," + url.PathEscape(s.Path))
	}
	return SubscriptionID(string(s.Event) + "," + string(s.FunctionID) + "," + url.PathEscape(s.Path))
}

// urlPathValidator validates if field contains URL path
func urlPathValidator(fl validator.FieldLevel) bool {
	return path.IsAbs(fl.Field().String())
}

// eventTypeValidator validates if field contains event name
func eventTypeValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`).MatchString(fl.Field().String())
}
