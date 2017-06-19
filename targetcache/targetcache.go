package targetcache

import (
	"errors"
	"strings"

	"go.uber.org/zap"

	"github.com/docker/libkv/store"

	"github.com/serverless/gateway/db"
	endpointTypes "github.com/serverless/gateway/endpoints/types"
	functionTypes "github.com/serverless/gateway/functions/types"
	pubsubTypes "github.com/serverless/gateway/pubsub/types"
)

// TargetCache is an interface for retrieving cached configuration
// for driving performance-sensitive routing decisions.
type TargetCache interface {
	BackingFunctions(endpoint endpointTypes.EndpointID) ([]functionTypes.WeightedFunction, error)
	GetFunction(functionID functionTypes.FunctionID) (functionTypes.Function, error)
	FunctionInputToTopics(function functionTypes.FunctionID) ([]pubsubTypes.TopicID, error)
	FunctionOutputToTopics(function functionTypes.FunctionID) ([]pubsubTypes.TopicID, error)
	SubscribersOfTopic(topic pubsubTypes.TopicID) ([]functionTypes.FunctionID, error)
}

// LibKVTargetCache is an implementation of TargetCache using the docker/libkv
// library for watching data in etcd, zookeeper, and consul.
type LibKVTargetCache struct {
	shutdown        chan struct{}
	functionCache   *functionCache
	endpointCache   *endpointCache
	publisherCache  *publisherCache
	subscriberCache *subscriberCache
	topicCache      *topicCache
}

// BackingFunctions returns the weighted functions and ID's for an endpoint
func (tc *LibKVTargetCache) BackingFunctions(endpointID endpointTypes.EndpointID) ([]functionTypes.WeightedFunction, error) {
	tc.endpointCache.RLock()
	_, exists := tc.endpointCache.cache[endpointID]
	tc.endpointCache.RUnlock()
	if !exists {
		return []functionTypes.WeightedFunction{}, errors.New("endpoint not found")
	}
	res := []functionTypes.WeightedFunction{}
	return res, nil
}

// GetFunction takes a function ID and returns a deserialized instance of that function, if it exists
func (tc *LibKVTargetCache) GetFunction(functionID functionTypes.FunctionID) (functionTypes.Function, error) {
	return functionTypes.Function{}, nil
}

// FunctionInputToTopics is used for determining the topics to forward inputs to a function to.
func (tc *LibKVTargetCache) FunctionInputToTopics(function functionTypes.FunctionID) ([]pubsubTypes.TopicID, error) {
	return []pubsubTypes.TopicID{}, nil
}

// FunctionOutputToTopics is used for determining the topics to forward outputs of a function to.
func (tc *LibKVTargetCache) FunctionOutputToTopics(function functionTypes.FunctionID) ([]pubsubTypes.TopicID, error) {
	return []pubsubTypes.TopicID{}, nil
}

// SubscribersOfTopic is used for determining which functions to forward messages in a topic to.
func (tc *LibKVTargetCache) SubscribersOfTopic(topic pubsubTypes.TopicID) ([]functionTypes.FunctionID, error) {
	return []functionTypes.FunctionID{}, nil
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
	// maintains list of known topics
	topicCache := newTopicCache(log)

	// start reacting to changes
	shutdown := make(chan struct{})
	functionPathWatcher.React(functionCache.reactor(), shutdown)
	endpointPathWatcher.React(endpointCache.reactor(), shutdown)
	subscriberPathWatcher.React(subscriberCache.reactor(), shutdown)
	publisherPathWatcher.React(publisherCache.reactor(), shutdown)

	return &LibKVTargetCache{
		shutdown:        shutdown,
		functionCache:   functionCache,
		endpointCache:   endpointCache,
		publisherCache:  publisherCache,
		subscriberCache: subscriberCache,
		topicCache:      topicCache,
	}
}
