package kv

import (
	"net/url"
)

// EndpointID uniquely identifies an endpoint.
type EndpointID string

// Endpoint represents single endpoint. It's used only for preventing from creating conflicting HTTP subscriptions.
type Endpoint struct {
	ID EndpointID `json:"endpointId"`
}

// NewEndpoint creates an Endpoint.
func NewEndpoint(method, path string) *Endpoint {
	return &Endpoint{
		ID: NewEndpointID(method, path),
	}
}

// NewEndpointID returns Endpoint ID.
func NewEndpointID(method, path string) EndpointID {
	return EndpointID(method + "," + url.PathEscape(path))
}
