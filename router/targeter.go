package router

import (
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/cors"
	"github.com/serverless/event-gateway/internal/pathtree"
)

// Targeter is an interface for retrieving cached configuration for driving performance-sensitive routing decisions.
type Targeter interface {
	HTTPBackingFunction(method, path string) (*functions.FunctionID, pathtree.Params, *cors.CORS)
	InvokableFunction(path string, functionID functions.FunctionID) bool
	Function(functionID functions.FunctionID) *functions.Function
	SubscribersOfEvent(path string, eventType event.Type) []functions.FunctionID
}
