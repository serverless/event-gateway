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

// NewEndpoint creates an Endpoint.
func NewEndpoint(functionID functions.FunctionID, method, path string) *Endpoint {
	return &Endpoint{
		ID:         NewEndpointID(method, path),
		FunctionID: functionID,
		Method:     method,
		Path:       path,
	}
}

// NewEndpointID returns Endpoint ID.
func NewEndpointID(method, path string) EndpointID {
	return EndpointID(method + "," + url.PathEscape(path))
}
