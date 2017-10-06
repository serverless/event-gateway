package router

import (
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/pathtree"
)

// Targeter is an interface for retrieving cached configuration for driving performance-sensitive routing decisions.
type Targeter interface {
	HTTPBackingFunction(method, path string) (*functions.FunctionID, pathtree.Params)
	Function(functionID functions.FunctionID) *functions.Function
	SubscribersOfEvent(path string, eventType event.Type) []functions.FunctionID
}
