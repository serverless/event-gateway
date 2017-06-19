package targetcache

import (
	"bytes"
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	endpointTypes "github.com/serverless/gateway/endpoints/types"
	functionTypes "github.com/serverless/gateway/functions/types"
	pubsubTypes "github.com/serverless/gateway/pubsub/types"
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
	cache map[string]functionTypes.Function
	log   *zap.Logger
}

func newFunctionCache(log *zap.Logger) *functionCache {
	return &functionCache{
		cache: map[string]functionTypes.Function{},
		log:   log,
	}
}

func (c *functionCache) Set(k string, v []byte) {
	f := functionTypes.Function{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&f)
	if err != nil {
		c.log.Error("Could not deserialize Function state!", zap.Error(err), zap.String("key", k))
	} else {
		c.Lock()
		defer c.Unlock()
		c.cache[k] = f
	}
}

func (c *functionCache) Del(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, k)
}

type endpointCache struct {
	sync.RWMutex
	// cache maps from EndpointID to Endpoint
	cache map[string]endpointTypes.Endpoint
	log   *zap.Logger
}

func newEndpointCache(log *zap.Logger) *endpointCache {
	return &endpointCache{
		cache: map[string]endpointTypes.Endpoint{},
		log:   log,
	}
}

func (c *endpointCache) Set(k string, v []byte) {
	e := endpointTypes.Endpoint{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&e)
	if err != nil {
		c.log.Error("Could not deserialize Endpoint state!", zap.Error(err), zap.String("key", k))
	} else {
		c.Lock()
		defer c.Unlock()
		c.cache[k] = e
	}
}

func (c *endpointCache) Del(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, k)
}

type publisherCache struct {
	sync.RWMutex
	// cache maps from PublisherID to Publisher
	cache map[string]pubsubTypes.Publisher
	log   *zap.Logger
}

func newPublisherCache(log *zap.Logger) *publisherCache {
	return &publisherCache{
		cache: map[string]pubsubTypes.Publisher{},
		log:   log,
	}
}

func (c *publisherCache) Set(k string, v []byte) {
	p := pubsubTypes.Publisher{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&p)
	if err != nil {
		c.log.Error("Could not deserialize Publisher state!", zap.Error(err), zap.String("key", k))
	} else {
		c.Lock()
		defer c.Unlock()
		c.cache[k] = p
	}
}

func (c *publisherCache) Del(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, k)
}

type subscriberCache struct {
	sync.RWMutex
	// cache maps from SubscriberID to Subscriber
	cache map[string]pubsubTypes.Subscriber
	log   *zap.Logger
}

func newSubscriberCache(log *zap.Logger) *subscriberCache {
	return &subscriberCache{
		cache: map[string]pubsubTypes.Subscriber{},
		log:   log,
	}
}

func (c *subscriberCache) Set(k string, v []byte) {
	s := pubsubTypes.Subscriber{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&s)
	if err != nil {
		c.log.Error("Could not deserialize Subscriber state!", zap.Error(err), zap.String("key", k))
	} else {
		c.Lock()
		defer c.Unlock()
		c.cache[k] = s
	}
}

func (c *subscriberCache) Del(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, k)
}

type topicCache struct {
	sync.RWMutex
	cache map[pubsubTypes.TopicID]struct{}
	log   *zap.Logger
}

func newTopicCache(log *zap.Logger) *topicCache {
	return &topicCache{
		// cache is a set of all TopicID's
		cache: map[pubsubTypes.TopicID]struct{}{},
		log:   log,
	}
}

func (c *topicCache) Set(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	c.cache[pubsubTypes.TopicID(k)] = struct{}{}
}

func (c *topicCache) Del(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, pubsubTypes.TopicID(k))
}
