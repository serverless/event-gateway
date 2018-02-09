package router

import (
	"net/http"
	"strings"

	"github.com/serverless/event-gateway/api"
)

// HTTPResponse is a response schema returned by subscribed function in case of HTTP event.
type HTTPResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

const (
	mimeJSON       = "application/json"
	mimeOctetStrem = "application/octet-stream"
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

type backlogEvent struct {
	path  string
	event api.Event
}
