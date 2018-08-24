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

const (
	// TransformationVersion is indicative of the revision of how Event Gateway transforms a request into CloudEvents format.
	TransformationVersion = "0.1"

	// CloudEventsVersion currently supported by Event Gateway
	CloudEventsVersion = "0.1"
)

// Event is a default event structure. All data that passes through the Event Gateway
// is formatted to a format defined CloudEvents v0.1 spec.
type Event struct {
	EventType          TypeName               `json:"eventType" validate:"required"`
	EventTypeVersion   string                 `json:"eventTypeVersion,omitempty"`
	CloudEventsVersion string                 `json:"cloudEventsVersion" validate:"required"`
	Source             string                 `json:"source" validate:"uri,required"`
	EventID            string                 `json:"eventID" validate:"required"`
	EventTime          *time.Time             `json:"eventTime,omitempty"`
	SchemaURL          string                 `json:"schemaURL,omitempty"`
	Extensions         zap.MapStringInterface `json:"extensions,omitempty"`
	ContentType        string                 `json:"contentType,omitempty"`
	Data               interface{}            `json:"data"`
}

// New return new instance of Event.
func New(eventType TypeName, mimeType string, payload interface{}) *Event {
	now := time.Now()

	event := &Event{
		EventType:          eventType,
		CloudEventsVersion: CloudEventsVersion,
		Source:             "https://serverless.com/event-gateway/#transformationVersion=" + TransformationVersion,
		EventID:            uuid.NewV4().String(),
		EventTime:          &now,
		ContentType:        mimeType,
		Data:               payload,
		Extensions: map[string]interface{}{
			"eventgateway": map[string]interface{}{
				"transformed":            "true",
				"transformation-version": TransformationVersion,
			},
		},
	}

	event.Data = normalizePayload(event.Data, event.ContentType)
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
				return New(TypeName(r.Header.Get("event")), mimeType, body), nil
			}
			return event, err
		}

		return New(TypeName(r.Header.Get("event")), mimeType, body), nil
	}

	return New(TypeHTTPRequest, mimeJSON, NewHTTPRequestData(r, body)), nil
}

// Validate Event struct
func (e *Event) Validate() error {
	validate := validator.New()
	err := validate.Struct(e)
	if err != nil {
		return &ErrParsingCloudEvent{Message: err.Error()}
	}
	return nil
}

// IsSystem indicates if the event is a system event.
func (e *Event) IsSystem() bool {
	return strings.HasPrefix(string(e.EventType), "eventgateway.")
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
	if e.EventTime != nil {
		enc.AddString("eventTime", e.EventTime.String())
	}
	if e.SchemaURL != "" {
		enc.AddString("schemaURL", e.SchemaURL)
	}
	if e.ContentType != "" {
		enc.AddString("contentType", e.ContentType)
	}
	if e.Extensions != nil {
		enc.AddObject("extensions", e.Extensions)
	}

	payload, _ := json.Marshal(e.Data)
	enc.AddString("data", string(payload))

	return nil
}

func isLegacyMode(headers http.Header) bool {
	return headers.Get("Event") != ""
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
		EventType:          TypeName(headers.Get("CE-EventType")),
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

	if headers.Get("CE-EventTime") != "" {
		val, err := time.Parse(time.RFC3339, headers.Get("CE-EventTime"))
		if err != nil {
			return nil, err
		}
		event.EventTime = &val
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

	event.Data = normalizePayload(event.Data, event.ContentType)
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

		event.Data = normalizePayload(event.Data, event.ContentType)
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

// normalizePayload takes anything, checks if it's []byte array and depending on provided mime
// type converts it to either string or map[string]interface to avoid having base64 string after
// JSON marshaling.
func normalizePayload(payload interface{}, mime string) interface{} {
	if bytePayload, ok := payload.([]byte); ok && len(bytePayload) > 0 {
		switch {
		case mime == mimeJSON || strings.HasSuffix(mime, "+json"):
			var result map[string]interface{}
			err := json.Unmarshal(bytePayload, &result)
			if err != nil {
				return payload
			}
			return result
		case strings.HasPrefix(mime, mimeFormMultipart), mime == mimeFormURLEncoded:
			return string(bytePayload)
		}
	}

	return payload
}
