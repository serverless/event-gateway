package event

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/satori/go.uuid"
	ihttp "github.com/serverless/event-gateway/internal/http"
	"github.com/serverless/event-gateway/internal/zap"
	"gopkg.in/go-playground/validator.v9"
)

// Type uniquely identifies an event type.
type Type string

const (
	// TypeInvoke is a special type of event for sync function invocation.
	TypeInvoke = Type("invoke")
	// TypeHTTP is a special type of event for sync http subscriptions.
	TypeHTTP = Type("http")
)

// TransformationVersion is indicative of the revision of how Event Gateway transforms a request into CloudEvents format.
const (
	TransformationVersion = "0.1"
)

const (
	mimeJSON           = "application/json"
	mimeFormMultipart  = "multipart/form-data"
	mimeFormURLEncoded = "application/x-www-form-urlencoded"
)

// Event is a default event structure. All data that passes through the Event Gateway
// is formatted to a format defined CloudEvents v0.1 spec.
type Event struct {
	EventType          Type                   `json:"eventType" validate:"required"`
	EventTypeVersion   string                 `json:"eventTypeVersion,omitempty"`
	CloudEventsVersion string                 `json:"cloudEventsVersion" validate:"required"`
	Source             string                 `json:"source" validate:"uri,required"`
	EventID            string                 `json:"eventID" validate:"required"`
	EventTime          time.Time              `json:"eventTime,omitempty"`
	SchemaURL          string                 `json:"schemaURL,omitempty"`
	Extensions         zap.MapStringInterface `json:"extensions,omitempty"`
	ContentType        string                 `json:"contentType,omitempty"`
	Data               interface{}            `json:"data"`
}

// New return new instance of Event.
func New(eventType Type, mimeType string, payload interface{}) *Event {
	event := &Event{
		EventType:          eventType,
		CloudEventsVersion: "0.1",
		Source:             "https://serverless.com/event-gateway/#transformationVersion=" + TransformationVersion,
		EventID:            uuid.NewV4().String(),
		EventTime:          time.Now(),
		ContentType:        mimeType,
		Data:               payload,
	}

	// Because event.Data is []bytes here, it will be base64 encoded by default when being sent to remote function,
	// which is why we change the event.Data type to "string" for forms, so that, it is left intact.
	if eventBody, ok := event.Data.([]byte); ok && len(eventBody) > 0 {
		switch {
		case isJSONMimeType(event.ContentType):
			json.Unmarshal(eventBody, &event.Data)
		case strings.HasPrefix(mimeType, mimeFormMultipart), mimeType == mimeFormURLEncoded:
			event.Data = string(eventBody)
		}
	}

	event.Extensions = zap.MapStringInterface{
		"eventgateway": map[string]interface{}{
			"transformed":            true,
			"transformation-version": TransformationVersion,
		},
	}

	return event
}

// FromRequest takes an HTTP request and returns an Event along with path
func FromRequest(r *http.Request) (*Event, error) {
	eventType := extractEventType(r)

	mimeType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		if err.Error() != "mime: no media type" {
			return nil, err
		}
		mimeType = "application/octet-stream"
	}

	body := []byte{}
	if r.Body != nil {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
	}

	var event *Event

	if isJSONMimeType(mimeType) {
		event, err = parseAsCloudEvent(mimeType, body)
	} else {
		event, err = parseAsCloudEventBinary(r.Header, body)
	}
	if err != nil {
		event = New(eventType, mimeType, body)
	}

	// Because event.Data is []bytes here, it will be base64 encoded by default when being sent to remote function,
	// which is why we change the event.Data type to "string" for forms, so that, it is left intact.
	if eventBody, ok := event.Data.([]byte); ok && len(eventBody) > 0 {
		switch {
		case isJSONMimeType(event.ContentType):
			json.Unmarshal(eventBody, &event.Data)
		case strings.HasPrefix(mimeType, mimeFormMultipart), mimeType == mimeFormURLEncoded:
			event.Data = string(eventBody)
		}
	}

	if eventType == TypeHTTP {
		event.ContentType = mimeJSON
		event.Data = NewHTTPRequestData(r, event.Data)
	}

	return event, nil
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (e Event) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("eventType", string(e.EventType))
	if e.EventTypeVersion != "" {
		enc.AddString("eventTypeVersion", e.EventTypeVersion)
	}
	enc.AddString("cloudEventsVersion", e.CloudEventsVersion)
	enc.AddString("source", e.Source)
	enc.AddString("eventID", e.EventID)
	enc.AddString("eventTime", e.EventTime.String())
	if e.SchemaURL != "" {
		enc.AddString("schemaURL", e.SchemaURL)
	}
	if e.ContentType != "" {
		enc.AddString("contentType", e.ContentType)
	}
	if e.Extensions != nil {
		e.Extensions.MarshalLogObject(enc)
	}

	payload, _ := json.Marshal(e.Data)
	enc.AddString("data", string(payload))

	return nil
}

func extractEventType(r *http.Request) Type {
	eventType := Type(r.Header.Get("event"))
	if eventType == "" {
		eventType = TypeHTTP
	}
	return eventType
}

func parseAsCloudEventBinary(headers http.Header, payload interface{}) (*Event, error) {
	he := Event{
		EventType:          Type(headers.Get("CE-EventType")),
		EventTypeVersion:   headers.Get("CE-EventTypeVersion"),
		CloudEventsVersion: headers.Get("CE-CloudEventsVersion"),
		Source:             headers.Get("CE-Source"),
		EventID:            headers.Get("CE-EventID"),
		Data:               payload,
	}
	err := validator.New().Struct(he)
	if err != nil {
		return nil, err
	}
	if val, err := time.Parse(time.RFC3339, headers.Get("CE-EventTime")); err == nil {
		he.EventTime = val
	}
	if val := headers.Get("CE-SchemaURL"); len(val) > 0 {
		he.SchemaURL = val
	}
	if val := headers.Get("CE-ContentType"); len(val) > 0 {
		he.ContentType = val
	}
	he.Extensions = map[string]interface{}{}
	for key, val := range ihttp.FlattenHeader(headers) {
		if strings.HasPrefix(key, "ce-x-") {
			he.Extensions[strings.TrimLeft(key, "ce-x-")] = val
		}
	}

	return &he, nil
}

// IsSystem indicates if the event is a system event.
func (e Event) IsSystem() bool {
	return strings.HasPrefix(string(e.EventType), "gateway.")
}

func parseAsCloudEvent(mime string, payload interface{}) (*Event, error) {
	if !isJSONMimeType(mime) {
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

		return customEvent, nil
	}

	return nil, errors.New("couldn't cast to []byte")
}

func isJSONMimeType(mime string) bool {
	return mime == mimeJSON || strings.HasSuffix(mime, "+json")
}
