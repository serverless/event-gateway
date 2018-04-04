package event

import "net/http"

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

func NewHTTPEvent(r *http.Request, eventData interface{}, headers map[string]string) *HTTPEvent {
	return &HTTPEvent{
		Headers: headers,
		Query:   r.URL.Query(),
		Body:    eventData,
		Host:    r.Host,
		Path:    r.URL.Path,
		Method:  r.Method,
	}
}
