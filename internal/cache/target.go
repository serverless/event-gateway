package cache

import (
	"strings"

	"github.com/serverless/libkv/store"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/kv"
	"github.com/serverless/event-gateway/internal/pathtree"
)

// Targeter is an interface for retrieving cached configuration for driving performance-sensitive routing decisions.
type Targeter interface {
	HTTPBackingFunction(method, path string) (*functions.FunctionID, pathtree.Params)
	Function(functionID functions.FunctionID) *functions.Function
	SubscribersOfEvent(eventType event.Type) []functions.FunctionID
}

// Target is an implementation of Targeter using the docker/libkv library for watching data in etcd, zookeeper, and
// consul.
type Target struct {
	log               *zap.Logger
	shutdown          chan struct{}
	functionCache     *functionCache
	endpointCache     *endpointCache
	subscriptionCache *subscriptionCache
}

// HTTPBackingFunction returns function ID for handling HTTP sync endpoint. It also returns matched URL parameters in
// case of HTTP subscription containing parameters in path.
func (tc *Target) HTTPBackingFunction(method, path string) (*functions.FunctionID, pathtree.Params) {
	tc.endpointCache.RLock()
	defer tc.endpointCache.RUnlock()

	root := tc.endpointCache.paths[method]
	if root == nil {
		return nil, nil
	}
	return root.Resolve(path)
}

// Function takes a function ID and returns a deserialized instance of that function, if it exists
func (tc *Target) Function(functionID functions.FunctionID) *functions.Function {
	tc.functionCache.RLock()
	defer tc.functionCache.RUnlock()
	return tc.functionCache.cache[functionID]
}

// SubscribersOfEvent is used for determining which functions to forward messages to.
func (tc *Target) SubscribersOfEvent(eventType event.Type) []functions.FunctionID {
	tc.subscriptionCache.RLock()
	fnSet, exists := tc.subscriptionCache.eventToFunctions[eventType]
	tc.subscriptionCache.RUnlock()

	if !exists {
		return []functions.FunctionID{}
	}

	res := []functions.FunctionID{}
	for fid := range fnSet {
		res = append(res, fid)
	}

	return res
}

// Shutdown causes all state watchers to clean up their state.
func (tc *Target) Shutdown() {
	close(tc.shutdown)
}

// NewTarget instantiates a new Target, rooted at a particular location.
func NewTarget(path string, kvstore store.Store, log *zap.Logger) *Target {
	// make sure we have a trailing slash for trimming future updates
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	functionPathWatcher := kv.NewWatcher(path+"functions", kvstore, log)
	endpointPathWatcher := kv.NewWatcher(path+"endpoints", kvstore, log)
	subscriptionPathWatcher := kv.NewWatcher(path+"subscriptions", kvstore, log)

	// updates dynamic routing information for endpoints when config changes are detected.
	endpointCache := newEndpointCache(log)
	// serves lookups for function info
	functionCache := newFunctionCache(log)
	// serves lookups for which functions are subscribed to an event
	subscriptionCache := newSubscriptionCache(log)

	// start reacting to changes
	shutdown := make(chan struct{})
	functionPathWatcher.React(functionCache, shutdown)
	endpointPathWatcher.React(endpointCache, shutdown)
	subscriptionPathWatcher.React(subscriptionCache, shutdown)

	return &Target{
		log:               log,
		shutdown:          shutdown,
		functionCache:     functionCache,
		endpointCache:     endpointCache,
		subscriptionCache: subscriptionCache,
	}
}
