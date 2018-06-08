package event

import (
	"net/http"

	ihttp "github.com/serverless/event-gateway/internal/http"
)

// HTTPRequestData is a event schema used for sending events to HTTP subscriptions.
type HTTPRequestData struct {
	Headers map[string]string   `json:"headers"`
	Query   map[string][]string `json:"query"`
	Body    interface{}         `json:"body"`
	Host    string              `json:"host"`
	Path    string              `json:"path"`
	Method  string              `json:"method"`
	Params  map[string]string   `json:"params"`
}

// NewHTTPRequestData returns a new instance of HTTPRequestData
func NewHTTPRequestData(r *http.Request, eventData interface{}) *HTTPRequestData {
	req := &HTTPRequestData{
		Headers: ihttp.FlattenHeader(r.Header),
		Query:   r.URL.Query(),
		Body:    eventData,
		Host:    r.Host,
		Path:    r.URL.Path,
		Method:  r.Method,
	}

	req.Body = normalizePayload(req.Body, r.Header.Get("content-type"))
	return req
}
