package subscriptions

import (
	"net/url"

	"github.com/serverless/event-gateway/functions"
	istrings "github.com/serverless/event-gateway/internal/strings"
)

// EndpointID uniquely identifies an endpoint.
type EndpointID string

type EndpointType string

const (
	EndpointTypeSync  = EndpointType("sync")
	EndpointTypeAsync = EndpointType("async")
)

// Endpoint represents single endpoint. It's used for preventing from creating
type Endpoint struct {
	ID         EndpointID           `json:"endpointId"`
	FunctionID functions.FunctionID `json:"functionId"`
	Method     string               `json:"method"`
	Path       string               `json:"path"`
}

// NewEndpoint creates an Endpoint.
func NewEndpoint(functionID functions.FunctionID, space, method, path string) *Endpoint {
	return &Endpoint{
		ID:         NewEndpointID(space, method, path),
		FunctionID: functionID,
		Method:     method,
		Path:       path,
	}
}

// NewEndpointID returns Endpoint ID.
func NewEndpointID(space, method, path string) EndpointID {
	return EndpointID(method + "," + url.PathEscape(istrings.EnsurePrefix(space+path, "/")))
}
