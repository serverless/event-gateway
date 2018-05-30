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
	AsyncSubscribers(method, path string, eventType event.TypeName) []AsyncSubscriber
	SyncSubscriber(method, path string, eventType event.TypeName) *SyncSubscriber
}

// AsyncSubscriber store info about space and function ID.
type AsyncSubscriber struct {
	Space      string
	FunctionID function.ID
}

// SyncSubscriber store info about space, function ID, path params and CORS configuration for sync subscriptions.
type SyncSubscriber struct {
	Space      string
	FunctionID function.ID
	Params     pathtree.Params
	CORS       *subscription.CORS
}
