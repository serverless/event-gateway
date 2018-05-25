package router

import (
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/subscription"
)

// Targeter is an interface for retrieving cached configuration for driving performance-sensitive routing decisions.
type Targeter interface {
	HTTPBackingFunction(method, path string) (string, *function.ID, pathtree.Params, *subscription.CORS)
	Function(space string, id function.ID) *function.Function
	SubscribersOfEvent(path string, eventType event.Type) []FunctionInfo
}

// FunctionInfo store info about space and function ID.
type FunctionInfo struct {
	Space string
	ID    function.ID
}
