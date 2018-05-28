package router

import (
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/subscription"
)

// Targeter is an interface for retrieving cached configuration for driving performance-sensitive routing decisions.
type Targeter interface {
	Function(space string, id function.ID) *function.Function
	EventType(space string, name event.TypeName) *event.Type
	AsyncSubscribers(path string, eventType event.TypeName) []FunctionInfo
	SyncSubscriber(method, path string) (string, *function.ID, pathtree.Params, *subscription.CORS)
}

// FunctionInfo store info about space and function ID.
type FunctionInfo struct {
	Space string
	ID    function.ID
}
