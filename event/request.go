package event

import (
	"io/ioutil"
	"mime"
	"net/http"

	"time"

	"strings"
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

	if isJSONContent(mimeType) {
		event, err = parseAsCloudEvent(eventType, mimeType, body)
		if err != nil {
			event = New(eventType, mimeType, body)
		}
	} else {
		event = mapHeadersToEvent(New(eventType, mimeType, body), r.Header)
	}

	if eventType == TypeHTTP {
		event.Data = NewHTTPEvent(r, event.Data)
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

func mapHeadersToEvent(event *Event, headers http.Header) *Event {
	if len(headers.Get("CE-EventType")) < 1 ||
		len(headers.Get("CE-CloudEventsVersion")) < 1 ||
		len(headers.Get("CE-Source")) < 1 ||
		len(headers.Get("CE-EventID")) < 1 ||
		len(headers.Get("CE-EventTypeVersion")) < 1 {
		return event
	}
	if val := headers.Get("CE-EventType"); len(val) > 0 {
		event.EventType = Type(val)
	}
	if val := headers.Get("CE-EventTypeVersion"); len(val) > 0 {
		event.EventTypeVersion = val
	}
	if val := headers.Get("CE-CloudEventsVersion"); len(val) > 0 {
		event.CloudEventsVersion = val
	}
	if val := headers.Get("CE-Source"); len(val) > 0 {
		event.Source = val
	}
	if val := headers.Get("CE-EventID"); len(val) > 0 {
		event.EventID = val
	}
	if val, err := time.Parse(time.RFC3339, headers.Get("CE-EventTime")); err == nil {
		event.EventTime = val
	}
	if val := headers.Get("CE-SchemaURL"); len(val) > 0 {
		event.SchemaURL = val
	}
	if val := headers.Get("CE-ContentType"); len(val) > 0 {
		event.ContentType = val
	}
	for key, allVals := range headers {
		event.Extensions = map[string]interface{}{}
		if strings.HasPrefix(key, "CE-X-") {
			for _, val := range allVals {
				event.Extensions[key[5:]] = val
			}
		}
	}

	return event
}
