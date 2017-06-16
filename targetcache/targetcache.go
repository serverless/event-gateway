package targetcache

import (
	"strings"

	"go.uber.org/zap"

	"github.com/docker/libkv/store"

	"github.com/serverless/gateway/db"
	endpointTypes "github.com/serverless/gateway/endpoints/types"
	functionTypes "github.com/serverless/gateway/functions/types"
	pubsubTypes "github.com/serverless/gateway/pubsub/types"
)

type TargetCache interface {
	BackingFunctions(endpoint endpointTypes.Endpoint) []functionTypes.WeightedFunction
	GetFunction(functionID functionTypes.FunctionID) functionTypes.Function
	FunctionInputToTopics(function functionTypes.FunctionID) []pubsubTypes.TopicID
	FunctionOutputToTopics(function functionTypes.FunctionID) []pubsubTypes.TopicID
	SubscribersOfTopic(topic pubsubTypes.TopicID) []functionTypes.FunctionID
}

type LibKVTargetCache struct {
	shutdown chan struct{}
}

func (tc *LibKVTargetCache) BackingFunctions(endpoint endpointTypes.Endpoint) []functionTypes.WeightedFunction {
	return []functionTypes.WeightedFunction{}
}

func (tc *LibKVTargetCache) GetFunction(functionID functionTypes.FunctionID) functionTypes.Function {
	return functionTypes.Function{}
}

func (tc *LibKVTargetCache) FunctionInputToTopics(function functionTypes.FunctionID) []pubsubTypes.TopicID {
	return []pubsubTypes.TopicID{}
}

func (tc *LibKVTargetCache) FunctionOutputToTopics(function functionTypes.FunctionID) []pubsubTypes.TopicID {
	return []pubsubTypes.TopicID{}
}

func (tc *LibKVTargetCache) SubscribersOfTopic(topic pubsubTypes.TopicID) []functionTypes.FunctionID {
	return []functionTypes.FunctionID{}
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
	endpointCache := NewEndpointCache(log)
	// serves lookups for function info
	functionCache := NewFunctionCache(log)
	// serves lookups for which functions are subscribed to a topic
	subscriberCache := NewSubscriberCache(log)
	// serves lookups for which topics a function's input or output are published to
	publisherCache := NewPublisherCache(log)

	shutdown := make(chan struct{})

	// start reacting to changes
	functionPathWatcher.React(functionCache, shutdown)
	endpointPathWatcher.React(endpointCache, shutdown)
	subscriberPathWatcher.React(subscriberCache, shutdown)
	publisherPathWatcher.React(publisherCache, shutdown)

	return &LibKVTargetCache{
		shutdown: shutdown,
	}
}
