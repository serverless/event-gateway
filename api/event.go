package api

import (
	"encoding/json"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	uuid "github.com/satori/go.uuid"
)

// Event is a default event structure. All data that passes through the Event Gateway is formatted as an Event, based on
// this schema.
type Event struct {
	Type       EventType   `json:"event"`
	ID         string      `json:"id"`
	ReceivedAt uint64      `json:"receivedAt"`
	Data       interface{} `json:"data"`
	DataType   string      `json:"dataType"`
}

// EventType uniquely identifies an event type.
type EventType string

// EventTypeInvoke is a special type of event for sync function invocation.
const EventTypeInvoke = EventType("invoke")

// EventTypeHTTP is a special type of event for sync http subscriptions.
const EventTypeHTTP = EventType("http")

// NewEvent return new instance of Event.
func NewEvent(eventType EventType, mime string, payload interface{}) *Event {
	return &Event{
		Type:       eventType,
		ID:         uuid.NewV4().String(),
		ReceivedAt: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		DataType:   mime,
		Data:       payload,
	}
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (e Event) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", string(e.Type))
	enc.AddString("id", e.ID)
	enc.AddUint64("receivedAt", e.ReceivedAt)
	payload, _ := json.Marshal(e.Data)
	enc.AddString("data", string(payload))
	enc.AddString("dataType", e.DataType)

	return nil
}

// IsSystem indicates if th event is a system event.
func (e Event) IsSystem() bool {
	return strings.HasPrefix(string(e.Type), "gateway.")
}

// HTTPEvent is a event schema used for sending events to HTTP subscriptions.
type HTTPEvent struct {
	Headers map[string][]string `json:"headers"`
	Query   map[string][]string `json:"query"`
	Body    interface{}         `json:"body"`
	Host    string              `json:"host"`
	Path    string              `json:"path"`
	Method  string              `json:"method"`
	Params  map[string]string   `json:"params"`
}
