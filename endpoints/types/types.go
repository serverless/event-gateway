package types

import (
	functionTypes "github.com/serverless/gateway/functions/types"
)

// EndpointID uniquely identifies an endpoint
type EndpointID string

// Endpoint represents single endpoint
type Endpoint struct {
	ID         EndpointID               `json:"id"`
	FunctionID functionTypes.FunctionID `json:"functionId"`
	Method     string                   `json:"method"`
	Path       string                   `json:"path"`
}
