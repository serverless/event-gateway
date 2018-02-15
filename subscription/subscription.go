package subscription

import (
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
)

// Subscription maps event type to a function.
type Subscription struct {
	Space      string      `json:"space" validate:"required,min=3,space"`
	ID         ID          `json:"subscriptionId"`
	Event      event.Type  `json:"event" validate:"required,eventtype"`
	FunctionID function.ID `json:"functionId" validate:"required"`
	Method     string      `json:"method,omitempty" validate:"omitempty,eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
	Path       string      `json:"path,omitempty" validate:"omitempty,urlpath"`
	CORS       *CORS       `json:"cors,omitempty"`
}

// Subscriptions is an array of subscriptions.
type Subscriptions []*Subscription

// ID uniquely identifies a subscription.
type ID string

// CORS is used to configure CORS on HTTP subscriptions.
type CORS struct {
	Origins          []string `json:"origins" validate:"min=1"`
	Methods          []string `json:"methods" validate:"min=1"`
	Headers          []string `json:"headers" validate:"min=1"`
	AllowCredentials bool     `json:"allowCredentials"`
}
