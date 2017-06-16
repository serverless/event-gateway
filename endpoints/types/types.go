package types

// EndpointID uniquely identifies an endpoint
type EndpointID string

// Endpoint represents single endpoint
type Endpoint struct {
	ID         string `json:"id"`
	FunctionID string `json:"functionId"`
	Method     string `json:"method"`
	Path       string `json:"path"`
}
