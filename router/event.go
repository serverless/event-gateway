package router

import (
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

	body := []byte{}
	var err error
	if r.Body != nil {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, "", err
		}
	}

	event := eventpkg.New(eventType, mime, body)

	if eventType == eventpkg.TypeHTTP {
		event.Data = eventpkg.NewHTTPEvent(r, event.Data, headers)
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
