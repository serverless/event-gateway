package router

import "github.com/serverless/event-gateway/event"

// AuthorizerPayload is a object that authorizer function is invoked with
type AuthorizerPayload struct {
	Event   event.Event           `json:"event"`
	Request event.HTTPRequestData `json:"request"`
}

// AuthorizerResponse is a expected result from authorizer function
type AuthorizerResponse struct {
	Authorization      *Authorization      `json:"authorization"`
	AuthorizationError *AuthorizationError `json:"error"`
}

// Authorization is an object containing authorization information
type Authorization struct {
	PrincipalID string            `json:"principalId"`
	Context     map[string]string `json:"context"`
}

// AuthorizationError represents error during authorization process
type AuthorizationError struct {
	Message string `json:"message"`
}
