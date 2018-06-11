package subscription

import (
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"go.uber.org/zap/zapcore"
)

// Subscription maps event type to a function.
type Subscription struct {
	Space      string         `json:"space" validate:"required,min=3,space"`
	ID         ID             `json:"subscriptionId"`
	Type       Type           `json:"type" validate:"required,eq=async|eq=sync"`
	EventType  event.TypeName `json:"eventType" validate:"required,eventType"`
	FunctionID function.ID    `json:"functionId" validate:"required"`
	Path       string         `json:"path" validate:"required,urlPath"`
	Method     string         `json:"method" validate:"required,eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
}

// ID uniquely identifies a subscription.
type ID string

// Type of subscription.
type Type string

const (
	// TypeSync causes that function invocation result will be returned to the caller.
	TypeSync = Type("sync")
	// TypeAsync causes asynchronous event processing.
	TypeAsync = Type("async")
)

// Subscriptions is an array of subscriptions.
type Subscriptions []*Subscription

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (s Subscription) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("space", string(s.Space))
	enc.AddString("subscriptionId", string(s.ID))
	enc.AddString("type", string(s.Type))
	enc.AddString("eventType", string(s.EventType))
	enc.AddString("functionId", string(s.FunctionID))
	if s.Method != "" {
		enc.AddString("method", string(s.Method))
	}
	if s.Path != "" {
		enc.AddString("path", string(s.Path))
	}

	return nil
}
