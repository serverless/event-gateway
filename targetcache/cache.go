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

type functionCache struct {
	sync.RWMutex
	cache map[string]functionTypes.Function
	log   *zap.Logger
}

func newFunctionCache(log *zap.Logger) *functionCache {
	return &functionCache{
		cache: map[string]functionTypes.Function{},
		log:   log,
	}
}

func (c *functionCache) reactor() *cacheMaintainer {
	return newCacheMaintainer(
		func(k string, v []byte) {
			f := functionTypes.Function{}
			err := json.NewDecoder(bytes.NewReader(v)).Decode(&f)
			if err != nil {
				c.log.Error("Could not deserialize received state!", zap.Error(err), zap.String("key", k))
			} else {
				c.Lock()
				defer c.Unlock()
				c.cache[k] = f
			}
		},
		func(k string, v []byte) {
			c.Lock()
			defer c.Unlock()
			delete(c.cache, k)
		},
	)
}

type endpointCache struct {
	sync.RWMutex
	cache map[endpointTypes.EndpointID]endpointTypes.Endpoint
	log   *zap.Logger
}

func newEndpointCache(log *zap.Logger) *endpointCache {
	return &endpointCache{
		cache: map[endpointTypes.EndpointID]endpointTypes.Endpoint{},
		log:   log,
	}
}

func (c *endpointCache) reactor() *cacheMaintainer {
	return newCacheMaintainer(
		func(k string, v []byte) {
			e := endpointTypes.Endpoint{}
			err := json.NewDecoder(bytes.NewReader(v)).Decode(&e)
			if err != nil {
				c.log.Error("Could not deserialize received state!", zap.Error(err), zap.String("key", k))
			} else {
				c.Lock()
				defer c.Unlock()
				c.cache[endpointTypes.EndpointID(k)] = e
			}
		},
		func(k string, v []byte) {
			c.Lock()
			defer c.Unlock()
			delete(c.cache, endpointTypes.EndpointID(k))
		},
	)
}

type publisherCache struct {
	sync.RWMutex
	cache map[string]pubsubTypes.Publisher
	log   *zap.Logger
}

func newPublisherCache(log *zap.Logger) *publisherCache {
	return &publisherCache{
		cache: map[string]pubsubTypes.Publisher{},
		log:   log,
	}
}

func (c *publisherCache) reactor() *cacheMaintainer {
	return newCacheMaintainer(
		func(k string, v []byte) {
			p := pubsubTypes.Publisher{}
			err := json.NewDecoder(bytes.NewReader(v)).Decode(&p)
			if err != nil {
				c.log.Error("Could not deserialize received state!", zap.Error(err), zap.String("key", k))
			} else {
				c.Lock()
				defer c.Unlock()
				c.cache[k] = p
			}
		},
		func(k string, v []byte) {
			c.Lock()
			defer c.Unlock()
			delete(c.cache, k)
		},
	)
}

type subscriberCache struct {
	sync.RWMutex
	cache map[string]pubsubTypes.Subscriber
	log   *zap.Logger
}

func newSubscriberCache(log *zap.Logger) *subscriberCache {
	return &subscriberCache{
		cache: map[string]pubsubTypes.Subscriber{},
		log:   log,
	}
}

func (c *subscriberCache) reactor() *cacheMaintainer {
	return newCacheMaintainer(
		func(k string, v []byte) {
			s := pubsubTypes.Subscriber{}
			err := json.NewDecoder(bytes.NewReader(v)).Decode(&s)
			if err != nil {
				c.log.Error("Could not deserialize received state!", zap.Error(err), zap.String("key", k))
			} else {
				c.Lock()
				defer c.Unlock()
				c.cache[k] = s
			}
		},
		func(k string, v []byte) {
			c.Lock()
			defer c.Unlock()
			delete(c.cache, k)
		},
	)
}

type topicCache struct {
	sync.RWMutex
	cache map[pubsubTypes.TopicID]struct{}
	log   *zap.Logger
}

func newTopicCache(log *zap.Logger) *topicCache {
	return &topicCache{
		cache: map[pubsubTypes.TopicID]struct{}{},
		log:   log,
	}
}

func (c *topicCache) reactor() *cacheMaintainer {
	return newCacheMaintainer(
		func(k string, v []byte) {
			c.Lock()
			defer c.Unlock()
			c.cache[pubsubTypes.TopicID(k)] = struct{}{}
		},
		func(k string, v []byte) {
			c.Lock()
			defer c.Unlock()
			delete(c.cache, pubsubTypes.TopicID(k))
		},
	)
}
