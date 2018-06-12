package libkv

import (
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/event-gateway/subscription/cors"
	"github.com/serverless/libkv/store"
	"go.uber.org/zap"
)

// Service implements function.Service and subscription.Service using libkv as a backend.
type Service struct {
	EventTypeStore    store.Store
	FunctionStore     store.Store
	SubscriptionStore store.Store
	CORSStore         store.Store
	Log               *zap.Logger
}

var _ event.Service = (*Service)(nil)
var _ function.Service = (*Service)(nil)
var _ subscription.Service = (*Service)(nil)
var _ cors.Service = (*Service)(nil)
