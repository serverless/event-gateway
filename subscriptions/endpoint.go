package subscriptions

import (
	"net/url"

	"github.com/serverless/event-gateway/functions"
)

// EndpointID uniquely identifies an endpoint.
type EndpointID string

// Endpoint represents single endpoint.
type Endpoint struct {
	ID         EndpointID           `json:"endpointId"`
	FunctionID functions.FunctionID `json:"functionId"`
	Method     string               `json:"method"`
	Path       string               `json:"path"`
}

// NewEndpointID returns Endpoint ID.
func NewEndpointID(method, path string) EndpointID {
	return EndpointID(method + "-" + url.PathEscape(path))
}
