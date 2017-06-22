package targetcache

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/docker/libkv/store"

	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/endpoints"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/pubsub"
)

// TargetCache is an interface for retrieving cached configuration
// for driving performance-sensitive routing decisions.
type TargetCache interface {
	BackingFunctions(endpoint endpoints.EndpointID) (functions.WeightedFunctions, *functions.FunctionID, error)
	Function(functionID functions.FunctionID) (functions.Function, error)
	FunctionInputToTopics(function functions.FunctionID) ([]pubsub.TopicID, error)
	FunctionOutputToTopics(function functions.FunctionID) ([]pubsub.TopicID, error)
	SubscribersOfTopic(topic pubsub.TopicID) ([]functions.FunctionID, error)
}

// LibKVTargetCache is an implementation of TargetCache using the docker/libkv
// library for watching data in etcd, zookeeper, and consul.
type LibKVTargetCache struct {
	shutdown        chan struct{}
	functionCache   *functionCache
	endpointCache   *endpointCache
	publisherCache  *publisherCache
	subscriberCache *subscriberCache
}

// BackingFunctions returns functions and their weights, along with the
// group ID if this was a Group function target, so we can submit
// events to topics that are fed by both.
func (tc *LibKVTargetCache) BackingFunctions(endpointID endpoints.EndpointID) (
	functions.WeightedFunctions, *functions.FunctionID, error,
) {

	// try to get the endpoint from our cache
	tc.endpointCache.RLock()
	endpoint, exists := tc.endpointCache.cache[endpointID]
	tc.endpointCache.RUnlock()
	if !exists {
		return functions.WeightedFunctions{}, nil, errors.New("endpoint not found")
	}

	// try to get the function from our cache
	fid := endpoint.FunctionID
	tc.functionCache.RLock()
	function, exists := tc.functionCache.cache[fid]
	tc.functionCache.RUnlock()
	if !exists {
		errMsg := fmt.Sprintf("Function %s not found in function cache. Is it configured?", fid)
		return functions.WeightedFunctions{}, nil, errors.New(errMsg)
	}

	// if function is a group, get weights, otherwise, just return the ID
	if function.Group == nil {
		res := functions.WeightedFunctions{
			{
				FunctionID: function.ID,
				Weight:     1,
			},
		}
		return res, nil, nil
	}

	return functions.WeightedFunctions(function.Group.Functions), &function.ID, nil
}

// Function takes a function ID and returns a deserialized instance of that function, if it exists
func (tc *LibKVTargetCache) Function(functionID functions.FunctionID) (functions.Function, error) {
	tc.functionCache.RLock()
	function, exists := tc.functionCache.cache[functionID]
	tc.functionCache.RUnlock()
	if !exists {
		errMsg := fmt.Sprintf("Function %s not found in function cache. Is it configured?", functionID)
		return function, errors.New(errMsg)
	}

	return function, nil
}

// FunctionInputToTopics is used for determining the topics to forward inputs to a function to.
func (tc *LibKVTargetCache) FunctionInputToTopics(function functions.FunctionID) ([]pubsub.TopicID, error) {
	tc.publisherCache.RLock()
	topicSet, exists := tc.publisherCache.fnInToTopic[function]
	tc.publisherCache.RUnlock()

	if !exists {
		return []pubsub.TopicID{}, nil
	}

	res := []pubsub.TopicID{}
	for tid := range topicSet {
		res = append(res, tid)
	}
	return res, nil
}

// FunctionOutputToTopics is used for determining the topics to forward outputs of a function to.
func (tc *LibKVTargetCache) FunctionOutputToTopics(function functions.FunctionID) ([]pubsub.TopicID, error) {
	tc.publisherCache.RLock()
	topicSet, exists := tc.publisherCache.fnOutToTopic[function]
	tc.publisherCache.RUnlock()

	if !exists {
		return []pubsub.TopicID{}, nil
	}

	res := []pubsub.TopicID{}
	for tid := range topicSet {
		res = append(res, tid)
	}
	return res, nil
}

// SubscribersOfTopic is used for determining which functions to forward messages in a topic to.
func (tc *LibKVTargetCache) SubscribersOfTopic(topic pubsub.TopicID) ([]functions.FunctionID, error) {
	tc.subscriberCache.RLock()
	fnSet, exists := tc.subscriberCache.topicToFns[topic]
	tc.subscriberCache.RUnlock()

	if !exists {
		return []functions.FunctionID{}, nil
	}

	res := []functions.FunctionID{}
	for fid := range fnSet {
		res = append(res, fid)
	}

	return res, nil
}

// Shutdown causes all state watchers to clean up their state.
func (tc *LibKVTargetCache) Shutdown() {
	close(tc.shutdown)
}

// New instantiates a new LibKVTargetCache, rooted at a particular location.
func New(path string, kv store.Store, log *zap.Logger) *LibKVTargetCache {
	// make sure we have a trailing slash for trimming future updates
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// path watchers
	functionPathWatcher := db.NewPathWatcher(path+"functions", kv, log)
	endpointPathWatcher := db.NewPathWatcher(path+"endpoints", kv, log)
	subscriberPathWatcher := db.NewPathWatcher(path+"subscribers", kv, log)
	publisherPathWatcher := db.NewPathWatcher(path+"publishers", kv, log)

	// updates dynamic routing information for endpoints when config changes are detected.
	endpointCache := newEndpointCache(log)
	// serves lookups for function info
	functionCache := newFunctionCache(log)
	// serves lookups for which functions are subscribed to a topic
	subscriberCache := newSubscriberCache(log)
	// serves lookups for which topics a function's input or output are published to
	publisherCache := newPublisherCache(log)

	// start reacting to changes
	shutdown := make(chan struct{})
	functionPathWatcher.React(newCacheMaintainer(functionCache), shutdown)
	endpointPathWatcher.React(newCacheMaintainer(endpointCache), shutdown)
	subscriberPathWatcher.React(newCacheMaintainer(subscriberCache), shutdown)
	publisherPathWatcher.React(newCacheMaintainer(publisherCache), shutdown)

	return &LibKVTargetCache{
		shutdown:        shutdown,
		functionCache:   functionCache,
		endpointCache:   endpointCache,
		publisherCache:  publisherCache,
		subscriberCache: subscriberCache,
	}
}
