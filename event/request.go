package event

import (
	"io/ioutil"
	"mime"
	"net/http"

	"time"

	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

const (
	mimeCloudEventsJSON = "application/cloudevents+json"
)

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
		if err != nil {
			event = New(eventType, mimeType, body)
		}
	} else {
		event = parseAsCloudEventBinary(New(eventType, mimeType, body), r.Header)
	}

	if eventType == TypeHTTP {
		event.Data = NewHTTPRequestData(r, event.Data)
	}

	return event, nil
}

func extractEventType(r *http.Request) Type {
	eventType := Type(r.Header.Get("event"))
	if eventType == "" {
		eventType = TypeHTTP
	}
	return eventType
}

func parseAsCloudEventBinary(event *Event, headers http.Header) *Event {
	he := Event{
		EventType:          Type(headers.Get("CE-EventType")),
		EventTypeVersion:   headers.Get("CE-EventTypeVersion"),
		CloudEventsVersion: headers.Get("CE-CloudEventsVersion"),
		Source:             headers.Get("CE-Source"),
		EventID:            headers.Get("CE-EventID"),
	}
	err := validator.New().Struct(he)
	if err != nil {
		return event
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
	for key, allVals := range headers {
		he.Extensions = map[string]interface{}{}
		if strings.HasPrefix(key, "CE-X-") {
			for _, val := range allVals {
				he.Extensions[key[5:]] = val
			}
		}
	}

	return &he
}
