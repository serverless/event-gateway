package cache

import (
	"strings"

	"github.com/serverless/libkv/store"
	"go.uber.org/zap"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/subscription"
)

// Target is an implementation of router.Targeter using the docker/libkv library for watching data in etcd, zookeeper, and
// consul.
type Target struct {
	log               *zap.Logger
	shutdown          chan struct{}
	functionCache     *functionCache
	subscriptionCache *subscriptionCache
}

// HTTPBackingFunction returns function ID for handling HTTP sync endpoint. It also returns matched URL parameters in
// case of HTTP subscription containing parameters in path.
func (tc *Target) HTTPBackingFunction(method, path string) (*function.ID, pathtree.Params, *subscription.CORS) {
	tc.subscriptionCache.RLock()
	defer tc.subscriptionCache.RUnlock()

	root := tc.subscriptionCache.endpoints[method]
	if root == nil {
		return nil, nil, nil
	}
	return root.Resolve(path)
}

// InvokableFunction returns function ID for handling invoke sync event.
func (tc *Target) InvokableFunction(path string, functionID function.ID) bool {
	tc.subscriptionCache.RLock()
	defer tc.subscriptionCache.RUnlock()

	_, exists := tc.subscriptionCache.invokable[path][functionID]
	return exists
}

// Function takes a function ID and returns a deserialized instance of that function, if it exists
func (tc *Target) Function(functionID function.ID) *function.Function {
	tc.functionCache.RLock()
	defer tc.functionCache.RUnlock()
	return tc.functionCache.cache[functionID]
}

// SubscribersOfEvent is used for determining which functions to forward messages to.
func (tc *Target) SubscribersOfEvent(path string, eventType eventpkg.Type) []function.ID {
	tc.subscriptionCache.RLock()
	defer tc.subscriptionCache.RUnlock()

	return tc.subscriptionCache.eventToFunctions[path][eventType]
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

	functionPathWatcher := NewWatcher(path+"functions", kvstore, log)
	subscriptionPathWatcher := NewWatcher(path+"subscriptions", kvstore, log)

	// serves lookups for function info
	functionCache := newFunctionCache(log)
	// serves lookups for which functions are subscribed to an event
	subscriptionCache := newSubscriptionCache(log)

	// start reacting to changes
	shutdown := make(chan struct{})
	functionPathWatcher.React(functionCache, shutdown)
	subscriptionPathWatcher.React(subscriptionCache, shutdown)

	return &Target{
		log:               log,
		shutdown:          shutdown,
		functionCache:     functionCache,
		subscriptionCache: subscriptionCache,
	}
}
