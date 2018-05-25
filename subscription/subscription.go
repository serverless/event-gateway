package subscription

import (
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/internal/zap"
	"go.uber.org/zap/zapcore"
)

// Subscription maps event type to a function.
type Subscription struct {
	Space      string      `json:"space" validate:"required,min=3,space"`
	ID         ID          `json:"subscriptionId"`
	Type       Type        `json:"type" validate:"required,eq=async|eq=sync"`
	EventType  event.Type  `json:"eventType" validate:"required,eventType"`
	FunctionID function.ID `json:"functionId" validate:"required"`
	Path       string      `json:"path" validate:"required,urlPath"`
	Method     string      `json:"method" validate:"required,eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
	CORS       *CORS       `json:"cors,omitempty"`
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
	if s.CORS != nil {
		enc.AddObject("cors", s.CORS)
	}

	return nil
}

// CORS is used to configure CORS on HTTP subscriptions.
type CORS struct {
	Origins          []string `json:"origins" validate:"min=1"`
	Methods          []string `json:"methods" validate:"min=1"`
	Headers          []string `json:"headers" validate:"min=1"`
	AllowCredentials bool     `json:"allowCredentials"`
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (c CORS) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddArray("origins", zap.Strings(c.Origins))
	enc.AddArray("methods", zap.Strings(c.Methods))
	enc.AddArray("headers", zap.Strings(c.Headers))
	enc.AddBool("allowCredentials", c.AllowCredentials)

	return nil
}
