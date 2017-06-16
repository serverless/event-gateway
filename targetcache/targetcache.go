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

type TargetCache interface {
	BackingFunctions(endpoint endpointTypes.EndpointID) ([]functionTypes.WeightedFunction, error)
	GetFunction(functionID functionTypes.FunctionID) (functionTypes.Function, error)
	FunctionInputToTopics(function functionTypes.FunctionID) ([]pubsubTypes.TopicID, error)
	FunctionOutputToTopics(function functionTypes.FunctionID) ([]pubsubTypes.TopicID, error)
	SubscribersOfTopic(topic pubsubTypes.TopicID) ([]functionTypes.FunctionID, error)
}

type LibKVTargetCache struct {
	shutdown        chan struct{}
	functionCache   *functionCache
	endpointCache   *endpointCache
	publisherCache  *publisherCache
	subscriberCache *subscriberCache
	topicCache      *topicCache
}

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

func (tc *LibKVTargetCache) GetFunction(functionID functionTypes.FunctionID) (functionTypes.Function, error) {
	return functionTypes.Function{}, nil
}

func (tc *LibKVTargetCache) FunctionInputToTopics(function functionTypes.FunctionID) ([]pubsubTypes.TopicID, error) {
	return []pubsubTypes.TopicID{}, nil
}

func (tc *LibKVTargetCache) FunctionOutputToTopics(function functionTypes.FunctionID) ([]pubsubTypes.TopicID, error) {
	return []pubsubTypes.TopicID{}, nil
}

func (tc *LibKVTargetCache) SubscribersOfTopic(topic pubsubTypes.TopicID) ([]functionTypes.FunctionID, error) {
	return []functionTypes.FunctionID{}, nil
}

func (tc *LibKVTargetCache) Shutdown() {
	close(tc.shutdown)
}

func New(path string, kv store.Store, log *zap.Logger) TargetCache {
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
