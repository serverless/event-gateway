package event

import (
	"encoding/json"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	uuid "github.com/satori/go.uuid"
	"github.com/serverless/event-gateway/internal/zap"
)

// Event is a default event structure. All data that passes through the Event Gateway is formatted as an Event, based on
// this schema.
type Event struct {
	Namespace          string `json:"namespace"`
	EventType          Type   `json:"event-type"`
	EventTypeVersion   string `json:"event-type-version"`
	CloudEventsVersion string `json:"cloud-events-version"`
	Source             struct {
		SourceType string `json:"source-type"`
		SourceID   string `json:"source-id"`
	} `json:"source"`
	EventID     string                 `json:"event-id"`
	EventTime   uint64                 `json:"event-time"`
	ContentType string                 `json:"content-type"`
	Extensions  zap.MapStringInterface `json:"extensions"`
	Data        interface{}            `json:"data"`
}

// New return new instance of Event.
func New(eventType Type, mime string, payload interface{}) *Event {
	return &Event{
		EventType:   eventType,
		EventID:     uuid.NewV4().String(),
		EventTime:   uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		ContentType: mime,
		Data:        payload,
	}
}

// Type uniquely identifies an event type.
type Type string

// TypeInvoke is a special type of event for sync function invocation.
const TypeInvoke = Type("invoke")

// TypeHTTP is a special type of event for sync http subscriptions.
const TypeHTTP = Type("http")

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (e Event) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if e.Namespace != "" {
		enc.AddString("namespace", string(e.Namespace))
	}
	enc.AddString("event-type", string(e.EventType))
	enc.AddString("event-type-version", string(e.EventTypeVersion))
	enc.AddString("cloud-events-version", string(e.CloudEventsVersion))
	enc.AddString("source.source-type", string(e.Source.SourceType))
	enc.AddString("source.source-id", string(e.Source.SourceID))
	enc.AddString("event-id", e.EventID)
	enc.AddUint64("event-time", e.EventTime)
	enc.AddString("content-type", e.ContentType)
	e.Extensions.MarshalLogObject(enc)
	payload, _ := json.Marshal(e.Data)
	enc.AddString("data", string(payload))

	return nil
}

// IsSystem indicates if th event is a system event.
func (e Event) IsSystem() bool {
	return strings.HasPrefix(string(e.EventType), "gateway.")
}
