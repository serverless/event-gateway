package router

import (
	"github.com/serverless/event-gateway/api"
	"github.com/serverless/event-gateway/internal/pathtree"
)

// Targeter is an interface for retrieving cached configuration for driving performance-sensitive routing decisions.
type Targeter interface {
	HTTPBackingFunction(method, path string) (*api.FunctionID, pathtree.Params, *api.CORS)
	InvokableFunction(path string, functionID api.FunctionID) bool
	Function(functionID api.FunctionID) *api.Function
	SubscribersOfEvent(path string, eventType api.EventType) []api.FunctionID
}
