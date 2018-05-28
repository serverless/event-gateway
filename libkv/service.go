package libkv

import (
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/libkv/store"
	"go.uber.org/zap"
)

// Service implements function.Service and subscription.Service using libkv as a backend.
type Service struct {
	EventTypeStore    store.Store
	FunctionStore     store.Store
	SubscriptionStore store.Store
	Log               *zap.Logger
}

var _ function.Service = (*Service)(nil)
var _ subscription.Service = (*Service)(nil)
