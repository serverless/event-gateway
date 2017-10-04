package event

import (
	"encoding/json"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	uuid "github.com/satori/go.uuid"
)

// Event is a default event structure. All data that passes through the Event Gateway is formatted as an Event, based on this schema.
type Event struct {
	Type       Type        `json:"event"`
	ID         string      `json:"id"`
	ReceivedAt uint64      `json:"receivedAt"`
	Data       interface{} `json:"data"`
	DataType   string      `json:"dataType"`
}

// NewEvent return new instance of Event.
func NewEvent(eventType Type, mime string, payload interface{}) *Event {
	return &Event{
		Type:       eventType,
		ID:         uuid.NewV4().String(),
		ReceivedAt: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		DataType:   mime,
		Data:       payload,
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
	enc.AddString("type", string(e.Type))
	enc.AddString("id", e.ID)
	enc.AddUint64("receivedAt", e.ReceivedAt)
	payload, _ := json.Marshal(e.Data)
	enc.AddString("data", string(payload))
	enc.AddString("dataType", e.DataType)

	return nil
}

// IsSystem indicates if system is a system event.
func (e Event) IsSystem() bool {
	return strings.HasPrefix(string(e.Type), "gateway.")
}
