package router

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	eventpkg "github.com/serverless/event-gateway/event"
	"go.uber.org/zap"
)

// HTTPResponse is a response schema returned by subscribed function in case of HTTP event.
type HTTPResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

const (
	mimeJSON = "application/json"
	mimeFormMultipart = "multipart/form-data"
	mimeFormURLEncoded = "application/x-www-form-urlencoded"
)

func isHTTPEvent(r *http.Request) bool {
	// is request with custom event
	if r.Header.Get("event") != "" {
		return false
	}

	// is pre-flight CORS request with "event" header
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		corsReqHeaders := r.Header.Get("Access-Control-Request-Headers")
		headers := strings.Split(corsReqHeaders, ",")
		for _, header := range headers {
			if header == "event" {
				return false
			}
		}
	}

	return true
}

func (router *Router) eventFromRequest(r *http.Request) (*eventpkg.Event, string, error) {
	path := extractPath(r.Host, r.URL.Path)
	eventType := extractEventType(r)
	headers := transformHeaders(r.Header)

	mime := r.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	var body []byte
	var err error
	if r.Body != nil {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, "", err
		}
	}

	event := eventpkg.New(eventType, mime, body)
	if customevent, err := router.parseCustomEventAsCloudEvent(eventType, event.ContentType, body); err == nil {
		event = customevent
	}

	// Because event.Data is []bytes here, it will be base64 encoded by default when being sent to remote function,
	// which is why we change the event.Data type to "string" for forms, so that, it is left intact.
	if eventbody, ok := event.Data.([]byte); ok && len(eventbody) > 0 {
		switch {
		case mime == mimeJSON:
			err := json.Unmarshal(eventbody, &event.Data)
			if err != nil {
				return nil, "", errors.New("malformed JSON body")
			}
		case strings.HasPrefix(mime, mimeFormMultipart), mime == mimeFormURLEncoded:
			event.Data = string(eventbody)
		}
	}

	if eventType == eventpkg.TypeHTTP {
		event.Data = &eventpkg.HTTPEvent{
			Headers: headers,
			Query:   r.URL.Query(),
			Body:    event.Data,
			Host:    r.Host,
			Path:    r.URL.Path,
			Method:  r.Method,
		}
	}

	router.log.Debug("Event received.", zap.String("path", path), zap.Object("event", event))
	err = router.emitSystemEventReceived(path, *event, headers)
	if err != nil {
		router.log.Debug("Event processing stopped because sync plugin subscription returned an error.",
			zap.Object("event", event),
			zap.Error(err))
		return nil, "", err
	}

	return event, path, nil
}

func (router *Router) parseCustomEventAsCloudEvent(originalEventType eventpkg.Type, contentType string, body []byte) (*eventpkg.Event, error) {
	var event = &eventpkg.Event{}
	if originalEventType == eventpkg.TypeHTTP || originalEventType == eventpkg.TypeInvoke {
		return event, errors.New("not a custom event")
	}

	err := json.Unmarshal(body, event)
	if err != nil {
		return event, err
	}
	if len(event.EventType) < 0 ||
		len(event.CloudEventsVersion) < 0 ||
		len(event.EventID) < 0 ||
		event.Data == nil {
			return event, errors.New("payload is not in valid CloudEvent format")
	}
	if originalEventType != event.EventType {
		return event, errors.New("event-type from header is not same as payload CloudEvent field")
	}
	return event, err
}

func extractPath(host, path string) string {
	extracted := path
	rxp, _ := regexp.Compile(hostedDomain)
	if rxp.MatchString(host) {
		subdomain := strings.Split(host, ".")[0]
		extracted = "/" + subdomain + path
	}
	return extracted
}

func extractEventType(r *http.Request) eventpkg.Type {
	eventType := eventpkg.Type(r.Header.Get("event"))
	if eventType == "" {
		eventType = eventpkg.TypeHTTP
	}
	return eventType
}

// transformHeaders takes http.Header and flatten value array (map[string][]string -> map[string]string) so it's easier
// to access headers by user.
func transformHeaders(req http.Header) map[string]string {
	headers := map[string]string{}
	for key, header := range req {
		headers[key] = header[0]
		if len(header) > 1 {
			headers[key] = strings.Join(header, ", ")
		}
	}

	return headers
}
