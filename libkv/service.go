package libkv

import (
	"github.com/serverless/libkv/store"
	"go.uber.org/zap"
)

// Service implements function.Service and subscription.Service using libkv as a backend.
type Service struct {
	SubscriptionStore store.Store
	FunctionStore     store.Store
	EndpointStore     store.Store
	Log               *zap.Logger
}
