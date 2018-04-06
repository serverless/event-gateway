package event

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	uuid "github.com/satori/go.uuid"
	"github.com/serverless/event-gateway/internal/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// Type uniquely identifies an event type.
type Type string

// TypeInvoke is a special type of event for sync function invocation.
const TypeInvoke = Type("invoke")

// TypeHTTP is a special type of event for sync http subscriptions.
const TypeHTTP = Type("http")

const (
	mimeJSON           = "application/json"
	mimeFormMultipart  = "multipart/form-data"
	mimeFormURLEncoded = "application/x-www-form-urlencoded"
)

// Event is a default event structure. All data that passes through the Event Gateway
// is formatted to a format defined CloudEvents v0.1 spec.
type Event struct {
	EventType          Type                   `json:"event-type" validate:"required"`
	EventTypeVersion   string                 `json:"event-type-version"`
	CloudEventsVersion string                 `json:"cloud-events-version" validate:"required"`
	Source             string                 `json:"source" validate:"url,required"`
	EventID            string                 `json:"event-id" validate:"required"`
	EventTime          time.Time              `json:"event-time"`
	SchemaURL          string                 `json:"schema-url"`
	ContentType        string                 `json:"content-type"`
	Extensions         zap.MapStringInterface `json:"extensions"`
	Data               interface{}            `json:"data"`
}

// New return new instance of Event.
func New(eventType Type, mime string, payload interface{}) *Event {
	event := &Event{
		EventType:          eventType,
		CloudEventsVersion: "0.1",
		Source:             "https://slsgateway.com/",
		EventID:            uuid.NewV4().String(),
		EventTime:          time.Now(),
		ContentType:        mime,
		Data:               payload,
	}

	// it's a custom event, possibly CloudEvent
	if eventType != TypeHTTP && eventType != TypeInvoke {
		cloudEvent, err := parseAsCloudEvent(eventType, mime, payload)
		if err == nil {
			event = cloudEvent
		} else {
			event.Extensions = zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": "0.1",
				},
			}
		}
	}

	// Because event.Data is []bytes here, it will be base64 encoded by default when being sent to remote function,
	// which is why we change the event.Data type to "string" for forms, so that, it is left intact.
	if eventbody, ok := event.Data.([]byte); ok && len(eventbody) > 0 {
		switch {
		case mime == mimeJSON:
			json.Unmarshal(eventbody, &event.Data)
		case strings.HasPrefix(mime, mimeFormMultipart), mime == mimeFormURLEncoded:
			event.Data = string(eventbody)
		}
	}

	return event
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (e Event) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("event-type", string(e.EventType))
	enc.AddString("event-type-version", e.EventTypeVersion)
	enc.AddString("cloud-events-version", e.CloudEventsVersion)
	enc.AddString("source", e.Source)
	enc.AddString("event-id", e.EventID)
	enc.AddString("event-time", e.EventTime.String())
	enc.AddString("schema-url", e.SchemaURL)
	enc.AddString("content-type", e.ContentType)
	e.Extensions.MarshalLogObject(enc)
	payload, _ := json.Marshal(e.Data)
	enc.AddString("data", string(payload))

	return nil
}

// IsSystem indicates if the event is a system event.
func (e Event) IsSystem() bool {
	return strings.HasPrefix(string(e.EventType), "gateway.")
}

func parseAsCloudEvent(eventType Type, mime string, payload interface{}) (*Event, error) {
	if mime != mimeJSON {
		return nil, errors.New("content type is not json")
	}
	body, ok := payload.([]byte)
	if ok {
		validate := validator.New()

		customEvent := &Event{}
		err := json.Unmarshal(body, customEvent)
		if err != nil {
			return nil, err
		}

		err = validate.Struct(customEvent)
		if err != nil {
			return nil, err
		}

		if eventType != customEvent.EventType {
			return nil, errors.New("wrong event type")
		}

		return customEvent, nil
	}

	return nil, errors.New("couldn't cast to []byte")
}
