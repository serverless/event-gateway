package cache

import (
	"strings"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/kv"
	"github.com/serverless/event-gateway/subscriptions"
)

// Targeter is an interface for retrieving cached configuration for driving performance-sensitive routing decisions.
type Targeter interface {
	BackingFunction(endpoint subscriptions.EndpointID) *functions.FunctionID
	Function(functionID functions.FunctionID) *functions.Function
	SubscribersOfTopic(topic subscriptions.TopicID) []functions.FunctionID
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

// BackingFunction returns functions and their weights, along with the group ID if this was a Group function target, so
// we can submit events to topics that are fed by both.
func (tc *Target) BackingFunction(endpointID subscriptions.EndpointID) *functions.FunctionID {
	// try to get the endpoint from our cache
	tc.endpointCache.RLock()
	defer tc.endpointCache.RUnlock()
	endpoint := tc.endpointCache.cache[endpointID]
	if endpoint == nil {
		return nil
	}
	return &endpoint.FunctionID
}

// Function takes a function ID and returns a deserialized instance of that function, if it exists
func (tc *Target) Function(functionID functions.FunctionID) *functions.Function {
	tc.functionCache.RLock()
	defer tc.functionCache.RUnlock()
	return tc.functionCache.cache[functionID]
}

// SubscribersOfTopic is used for determining which functions to forward messages in a topic to.
func (tc *Target) SubscribersOfTopic(topic subscriptions.TopicID) []functions.FunctionID {
	tc.subscriptionCache.RLock()
	fnSet, exists := tc.subscriptionCache.topicToFns[topic]
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
func NewTarget(path string, kvstore store.Store, log *zap.Logger, debug ...bool) *Target {
	// make sure we have a trailing slash for trimming future updates
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	functionPathWatcher := kv.NewWatcher(path+"functions", kvstore, log)
	endpointPathWatcher := kv.NewWatcher(path+"endpoints", kvstore, log)
	subscriptionPathWatcher := kv.NewWatcher(path+"subscriptions", kvstore, log)

	if len(debug) == 1 && debug[0] {
		debugReconciliation := func(w ...*kv.Watcher) {
			for _, w := range w {
				w.ReconciliationJitter = 0
				w.ReconciliationBaseDelay = 3
			}
		}
		debugReconciliation(functionPathWatcher, endpointPathWatcher, subscriptionPathWatcher)
	}

	// updates dynamic routing information for endpoints when config changes are detected.
	endpointCache := newEndpointCache(log)
	// serves lookups for function info
	functionCache := newFunctionCache(log)
	// serves lookups for which functions are subscribed to a topic
	subscriptionCache := newSubscriptionCache(log)

	// start reacting to changes
	shutdown := make(chan struct{})
	functionPathWatcher.React(newCacheMaintainer(functionCache), shutdown)
	endpointPathWatcher.React(newCacheMaintainer(endpointCache), shutdown)
	subscriptionPathWatcher.React(newCacheMaintainer(subscriptionCache), shutdown)

	return &Target{
		log:               log,
		shutdown:          shutdown,
		functionCache:     functionCache,
		endpointCache:     endpointCache,
		subscriptionCache: subscriptionCache,
	}
}
