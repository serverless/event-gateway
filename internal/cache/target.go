package cache

import (
	"strings"

	"github.com/serverless/libkv/store"
	"go.uber.org/zap"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/libkv"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/subscription/cors"
)

// Target is an implementation of router.Targeter using the docker/libkv library for watching data in etcd, zookeeper, and
// consul.
type Target struct {
	log               *zap.Logger
	shutdown          chan struct{}
	eventTypeCache    *eventTypeCache
	functionCache     *functionCache
	subscriptionCache *subscriptionCache
	corsCache         *corsCache
}

// EventType takes a event type name and returns a deserialized instance of event type, if it exists
func (tc *Target) EventType(space string, name eventpkg.TypeName) *eventpkg.Type {
	tc.eventTypeCache.RLock()
	defer tc.eventTypeCache.RUnlock()
	return tc.eventTypeCache.cache[libkv.EventTypeKey{Space: space, Name: name}]
}

// SyncSubscriber returns function space and ID for handling sync subscription. It also returns matched URL
// parameters in case of HTTP subscription containing parameters in path.
func (tc *Target) SyncSubscriber(method, path string, eventType eventpkg.TypeName) *router.SyncSubscriber {
	tc.subscriptionCache.RLock()
	defer tc.subscriptionCache.RUnlock()

	root := tc.subscriptionCache.sync[method][eventType]
	if root == nil {
		return nil
	}

	value, params := root.Resolve(path)
	if value == nil {
		return nil
	}

	key := value.(libkv.FunctionKey)
	return &router.SyncSubscriber{
		Space:      key.Space,
		FunctionID: key.ID,
		Params:     params,
	}
}

// Function takes a function ID and returns a deserialized instance of that function, if it exists
func (tc *Target) Function(space string, id function.ID) *function.Function {
	tc.functionCache.RLock()
	defer tc.functionCache.RUnlock()
	return tc.functionCache.cache[libkv.FunctionKey{Space: space, ID: id}]
}

// AsyncSubscribers is used for determining which functions is async subscribed to the event.
func (tc *Target) AsyncSubscribers(method, path string, eventType eventpkg.TypeName) []router.AsyncSubscriber {
	tc.subscriptionCache.RLock()
	defer tc.subscriptionCache.RUnlock()

	keys := tc.subscriptionCache.async[method][path][eventType]
	subscribers := []router.AsyncSubscriber{}
	for _, key := range keys {
		subscribers = append(subscribers, router.AsyncSubscriber{
			Space:      key.Space,
			FunctionID: key.ID,
		})
	}
	return subscribers
}

// CORS returns CORS configuration for method and path pair
func (tc *Target) CORS(method, path string) *cors.CORS {
	tc.corsCache.RLock()
	defer tc.corsCache.RUnlock()

	root := tc.corsCache.endpoints[method]
	if root == nil {
		return nil
	}

	value, _ := root.Resolve(path)
	if value == nil {
		return nil
	}

	config := value.(cors.CORS)
	return &config
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

	eventTypePathWatcher := NewWatcher(path+"eventtypes", kvstore, log)
	functionPathWatcher := NewWatcher(path+"functions", kvstore, log)
	subscriptionPathWatcher := NewWatcher(path+"subscriptions", kvstore, log)
	corsPathWatcher := NewWatcher(path+"cors", kvstore, log)

	// serves lookups for event types
	eventTypeCache := newEventTypeCache(log)
	// serves lookups for function info
	functionCache := newFunctionCache(log)
	// serves lookups for which functions are subscribed to an event
	subscriptionCache := newSubscriptionCache(log)
	// serves lookups for cors configuration
	corsCache := newCORSCache(log)

	// start reacting to changes
	shutdown := make(chan struct{})
	eventTypePathWatcher.React(eventTypeCache, shutdown)
	functionPathWatcher.React(functionCache, shutdown)
	subscriptionPathWatcher.React(subscriptionCache, shutdown)
	corsPathWatcher.React(corsCache, shutdown)

	return &Target{
		log:               log,
		shutdown:          shutdown,
		eventTypeCache:    eventTypeCache,
		functionCache:     functionCache,
		subscriptionCache: subscriptionCache,
		corsCache:         corsCache,
	}
}
