package event

import (
	"io/ioutil"
	"mime"
	"net/http"

	"errors"
	"time"
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

	switch mimeType {
	case mimeCloudEventsJSON:
		event, err = parseAsCloudEvent(eventType, mimeType, body)
		// reject if HTTP event and header indicates CloudEvent but the body is not valid CloudEvent
		if err != nil && eventType == TypeHTTP {
			return event, errors.New("payload is not a valid CloudEvent")
		}
	case mimeJSON:
		event, err = parseAsCloudEvent(eventType, mimeType, body)
		if err == nil {
			break
		}
		fallthrough
	default:
		event = mapHeadersToEvent(New(eventType, mimeType, body), r.Header)
	}

	//router.log.Debug("Event received.", zap.String("path", path), zap.Object("event", event))
	//err = router.emitSystemEventReceived(path, *event, r.Header)
	//if err != nil {
	//	router.log.Debug("Event processing stopped because sync plugin subscription returned an error.",
	//		zap.Object("event", event),
	//		zap.Error(err))
	//	return nil, "", err
	//}

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
	// TBD
	//for key, allVals := range headers {
	//	event.Extensions = map[string]interface{}{}
	//	if strings.HasPrefix(key, "CE-X-") {
	//		for _, val := range allVals {
	//
	//		}
	//	}
	//}

	return event
}
