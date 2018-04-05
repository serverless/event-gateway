package event

import (
	internalhttp "github.com/serverless/event-gateway/internal/http"
	"net/http"
)

// HTTPEvent is a event schema used for sending events to HTTP subscriptions.
type HTTPEvent struct {
	Headers map[string]string   `json:"headers"`
	Query   map[string][]string `json:"query"`
	Body    interface{}         `json:"body"`
	Host    string              `json:"host"`
	Path    string              `json:"path"`
	Method  string              `json:"method"`
	Params  map[string]string   `json:"params"`
}

// NewHTTPEvent returns a new instance of HTTPEvent
func NewHTTPEvent(r *http.Request, eventData interface{}) *HTTPEvent {
	headers := internalhttp.TransformHeaders(r.Header)
	
	return &HTTPEvent{
		Headers: headers,
		Query:   r.URL.Query(),
		Body:    eventData,
		Host:    r.Host,
		Path:    r.URL.Path,
		Method:  r.Method,
	}
}
