package targetcache

import (
	"strings"

	"go.uber.org/zap"

	"github.com/docker/libkv/store"

	"github.com/serverless/gateway/db"
	"github.com/serverless/gateway/types"
)

type TargetCache interface {
	BackingFunctions(endpoint types.Endpoint) []types.WeightedFunction
	GetFunction(functionID types.FunctionID) types.Function
	FunctionInputToTopics(function types.FunctionID) []types.TopicID
	FunctionOutputToTopics(function types.FunctionID) []types.TopicID
	SubscribersOfTopic(topic types.TopicID) []types.FunctionID
}

type LibKVTargetCache struct {
	shutdown chan struct{}
}

func (tc *LibKVTargetCache) BackingFunctions(endpoint types.Endpoint) []types.WeightedFunction {
	return []types.WeightedFunction{}
}

func (tc *LibKVTargetCache) GetFunction(functionID types.FunctionID) types.Function {
	return types.Function{}
}

func (tc *LibKVTargetCache) FunctionInputToTopics(function types.FunctionID) []types.TopicID {
	return []types.TopicID{}
}

func (tc *LibKVTargetCache) FunctionOutputToTopics(function types.FunctionID) []types.TopicID {
	return []types.TopicID{}
}

func (tc *LibKVTargetCache) SubscribersOfTopic(topic types.TopicID) []types.FunctionID {
	return []types.FunctionID{}
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
