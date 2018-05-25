package event

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap/zapcore"

	"github.com/satori/go.uuid"
	ihttp "github.com/serverless/event-gateway/internal/http"
	"github.com/serverless/event-gateway/internal/zap"
	"gopkg.in/go-playground/validator.v9"
)

// Type uniquely identifies an event type.
type Type string

const (
	// TypeHTTPRequest is a special type of event for sync http subscriptions.
	TypeHTTPRequest = Type("http.request")

	// TransformationVersion is indicative of the revision of how Event Gateway transforms a request into CloudEvents format.
	TransformationVersion = "0.1"

	// CloudEventsVersion currently supported by Event Gateway
	CloudEventsVersion = "0.1"
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
		CloudEventsVersion: CloudEventsVersion,
		Source:             "https://serverless.com/event-gateway/#transformationVersion=" + TransformationVersion,
		EventID:            uuid.NewV4().String(),
		EventTime:          time.Now(),
		ContentType:        mimeType,
		Data:               payload,
		Extensions: map[string]interface{}{
			"eventgateway": map[string]interface{}{
				"transformed":            true,
				"transformation-version": TransformationVersion,
			},
		},
	}

	event.enhanceEventData()
	return event
}

// FromRequest takes an HTTP request and returns an Event along with path. Most of the implementation
// is based on https://github.com/cloudevents/spec/blob/master/http-transport-binding.md.
// This function also supports legacy mode where event type is sent in Event header.
func FromRequest(r *http.Request) (*Event, error) {
	contentType := r.Header.Get("Content-Type")
	mimeType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		if err.Error() != "mime: no media type" {
			return nil, err
		}
		mimeType = "application/octet-stream"
	}
	// Read request body
	body := []byte{}
	if r.Body != nil {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
	}

	var event *Event
	if mimeType == mimeCloudEventsJSON { // CloudEvents Structured Content Mode
		return parseAsCloudEvent(mimeType, body)
	} else if isCloudEventsBinaryContentMode(r.Header) { // CloudEvents Binary Content Mode
		return parseAsCloudEventBinary(r.Header, body)
	} else if isLegacyMode(r.Header) {
		if mimeType == mimeJSON { // CloudEvent in Legacy Mode
			event, err = parseAsCloudEvent(mimeType, body)
			if err != nil {
				return New(Type(r.Header.Get("event")), mimeType, body), nil
			}
			return event, err
		}

		return New(Type(r.Header.Get("event")), mimeType, body), nil
	}

	return New(TypeHTTPRequest, mimeCloudEventsJSON, NewHTTPRequestData(r, body)), nil
}

// Validate Event struct
func (e *Event) Validate() error {
	validate := validator.New()
	err := validate.Struct(e)
	if err != nil {
		return &ErrParsingCloudEvent{err.Error()}
	}
	return nil
}

// IsSystem indicates if the event is a system event.
func (e *Event) IsSystem() bool {
	return strings.HasPrefix(string(e.EventType), "gateway.")
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

func isLegacyMode(headers http.Header) bool {
	if headers.Get("Event") != "" {
		return true
	}

	return false
}

func isCloudEventsBinaryContentMode(headers http.Header) bool {
	if headers.Get("CE-EventType") != "" &&
		headers.Get("CE-CloudEventsVersion") != "" &&
		headers.Get("CE-Source") != "" &&
		headers.Get("CE-EventID") != "" {
		return true
	}

	return false
}

func parseAsCloudEventBinary(headers http.Header, payload interface{}) (*Event, error) {
	event := &Event{
		EventType:          Type(headers.Get("CE-EventType")),
		EventTypeVersion:   headers.Get("CE-EventTypeVersion"),
		CloudEventsVersion: headers.Get("CE-CloudEventsVersion"),
		Source:             headers.Get("CE-Source"),
		EventID:            headers.Get("CE-EventID"),
		ContentType:        headers.Get("Content-Type"),
		Data:               payload,
	}

	err := event.Validate()
	if err != nil {
		return nil, err
	}

	if val, err := time.Parse(time.RFC3339, headers.Get("CE-EventTime")); err == nil {
		event.EventTime = val
	}

	if val := headers.Get("CE-SchemaURL"); len(val) > 0 {
		event.SchemaURL = val
	}

	event.Extensions = map[string]interface{}{}
	for key, val := range ihttp.FlattenHeader(headers) {
		if strings.HasPrefix(key, "Ce-X-") {
			key = strings.TrimLeft(key, "Ce-X-")
			// Make first character lowercase
			runes := []rune(key)
			runes[0] = unicode.ToLower(runes[0])
			event.Extensions[string(runes)] = val
		}
	}

	event.enhanceEventData()
	return event, nil
}

func parseAsCloudEvent(mime string, payload interface{}) (*Event, error) {
	body, ok := payload.([]byte)
	if ok {
		event := &Event{}
		err := json.Unmarshal(body, event)
		if err != nil {
			return nil, err
		}

		err = event.Validate()
		if err != nil {
			return nil, err
		}

		event.enhanceEventData()
		return event, nil
	}

	return nil, errors.New("couldn't cast to []byte")
}

const (
	mimeJSON            = "application/json"
	mimeFormMultipart   = "multipart/form-data"
	mimeFormURLEncoded  = "application/x-www-form-urlencoded"
	mimeCloudEventsJSON = "application/cloudevents+json"
)

// Because event.Data is []byte, it will be base64 encoded by default when being sent to remote function,
// which is why we change the event.Data type to "string" for forms or to map[string]interface{} for JSON
// so that, it is left intact.
func (e *Event) enhanceEventData() {
	contentType := e.ContentType
	if eventBody, ok := e.Data.([]byte); ok && len(eventBody) > 0 {
		switch {
		case contentType == mimeJSON || strings.HasSuffix(contentType, "+json"):
			json.Unmarshal(eventBody, &e.Data)
		case strings.HasPrefix(contentType, mimeFormMultipart), contentType == mimeFormURLEncoded:
			e.Data = string(eventBody)
		}
	}
}
