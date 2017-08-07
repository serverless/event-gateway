package targetcache

import (
	"bytes"
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/pubsub"
)

// Cache is a simplification of the db.Reactive interface, which doesn't care about
// the distinction between Created and Modified, reducing them to Set.
type Cache interface {
	Set(string, []byte)
	Del(string, []byte)
}

type cacheMaintainer struct {
	cache Cache
}

func newCacheMaintainer(cache Cache) *cacheMaintainer {
	return &cacheMaintainer{
		cache: cache,
	}
}

// Created is called when a new endpoint is detected in the config.
func (c *cacheMaintainer) Created(key string, value []byte) {
	c.cache.Set(key, value)
}

// Modified is called when an existing endpoint is modified in the config.
func (c *cacheMaintainer) Modified(key string, newValue []byte) {
	c.cache.Set(key, newValue)
}

// Deleted is called when a endpoint is deleted in the config.
func (c *cacheMaintainer) Deleted(key string, lastKnownValue []byte) {
	c.cache.Del(key, lastKnownValue)
}

type functionCache struct {
	sync.RWMutex
	// cache maps from FunctionID to Function
	cache map[functions.FunctionID]*functions.Function
	log   *zap.Logger
}

func newFunctionCache(log *zap.Logger) *functionCache {
	return &functionCache{
		cache: map[functions.FunctionID]*functions.Function{},
		log:   log,
	}
}

func (c *functionCache) Set(k string, v []byte) {
	f := &functions.Function{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(f)
	c.log.Debug("Function local cache received value update.", zap.String("key", k), zap.String("value", string(v)))
	if err != nil {
		c.log.Error("Could not deserialize Function state!", zap.Error(err), zap.String("key", k))
	} else {
		c.Lock()
		defer c.Unlock()
		c.cache[functions.FunctionID(k)] = f
	}
}

func (c *functionCache) Del(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, functions.FunctionID(k))
}

type endpointCache struct {
	sync.RWMutex
	// cache maps from EndpointID to Endpoint
	cache map[pubsub.EndpointID]*pubsub.Endpoint
	log   *zap.Logger
}

func newEndpointCache(log *zap.Logger) *endpointCache {
	return &endpointCache{
		cache: map[pubsub.EndpointID]*pubsub.Endpoint{},
		log:   log,
	}
}

func (c *endpointCache) Set(k string, v []byte) {
	e := &pubsub.Endpoint{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(e)
	c.log.Debug("Endpoint local cache received value update.", zap.String("key", k), zap.String("value", string(v)))
	if err != nil {
		c.log.Error("Could not deserialize Endpoint state!", zap.Error(err), zap.String("key", k))
	} else {
		c.Lock()
		defer c.Unlock()
		c.cache[pubsub.EndpointID(k)] = e
	}
}

func (c *endpointCache) Del(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, pubsub.EndpointID(k))
}

type subscriptionCache struct {
	sync.RWMutex
	// topicToSub maps from a TopicID to a set of subscribing FunctionID's
	topicToFns map[pubsub.TopicID]map[functions.FunctionID]struct{}
	log        *zap.Logger
}

func newSubscriptionCache(log *zap.Logger) *subscriptionCache {
	return &subscriptionCache{
		// topicToFns is a map from TopicID to a set of FunctionID's
		topicToFns: map[pubsub.TopicID]map[functions.FunctionID]struct{}{},
		log:        log,
	}
}

func (c *subscriptionCache) Set(k string, v []byte) {
	s := pubsub.Subscription{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&s)
	c.log.Debug("Subscription local cache received value update.", zap.String("key", k), zap.String("value", string(v)))
	if err != nil {
		c.log.Error("Could not deserialize Subscription state!", zap.Error(err), zap.String("key", k))
		return
	}

	c.Lock()
	defer c.Unlock()

	// set FunctionID as destination in topicToSub
	fnSet, exists := c.topicToFns[s.Event]
	if exists {
		fnSet[s.FunctionID] = struct{}{}
	} else {
		fnSet := map[functions.FunctionID]struct{}{}
		fnSet[s.FunctionID] = struct{}{}
		c.topicToFns[s.Event] = fnSet
	}
}

func (c *subscriptionCache) Del(k string, v []byte) {
	c.Lock()
	defer c.Unlock()

	oldSub := pubsub.Subscription{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&oldSub)
	if err != nil {
		c.log.Error("Could not deserialize Subscription state during deletion!", zap.Error(err), zap.String("key", k))
		return
	}

	fnSet, exists := c.topicToFns[oldSub.Event]
	if exists {
		delete(fnSet, oldSub.FunctionID)

		if len(fnSet) == 0 {
			delete(c.topicToFns, oldSub.Event)
		}
	}
}
