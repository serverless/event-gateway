package event

import (
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/metadata"
	"go.uber.org/zap/zapcore"
)

const (
	// TypeHTTPRequest is a special type of event HTTP requests that are not CloudEvents.
	TypeHTTPRequest = TypeName("http.request")
)

// TypeName uniquely identifies an event type.
type TypeName string

// Type is a registered event type.
type Type struct {
	Space        string       `json:"space" validate:"required,min=3,space"`
	Name         TypeName     `json:"name" validate:"required"`
	AuthorizerID *function.ID `json:"authorizerId,omitempty"`

	Metadata *metadata.Metadata `json:"metadata,omitempty"`
}

// Types is an array of subscriptions.
type Types []*Type

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (t Type) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("space", string(t.Space))
	enc.AddString("name", string(t.Name))
	if t.AuthorizerID != nil {
		enc.AddString("authorizer", string(*t.AuthorizerID))
	}

	return nil
}
