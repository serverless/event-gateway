package cache

import (
	"bytes"
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/subscriptions"
)

type functionCache struct {
	sync.RWMutex
	cache map[functions.FunctionID]*functions.Function
	log   *zap.Logger
}

func newFunctionCache(log *zap.Logger) *functionCache {
	return &functionCache{
		cache: map[functions.FunctionID]*functions.Function{},
		log:   log,
	}
}

func (c *functionCache) Modified(k string, v []byte) {
	c.log.Debug("Function local cache received value update.", zap.String("key", k), zap.String("value", string(v)))

	f := &functions.Function{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(f)
	if err != nil {
		c.log.Error("Could not deserialize Function state!", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
	} else {
		c.Lock()
		defer c.Unlock()
		c.cache[functions.FunctionID(k)] = f
	}
}

func (c *functionCache) Deleted(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, functions.FunctionID(k))
}

type endpointCache struct {
	sync.RWMutex
	// paths maps HTTP method to internal/pathtree.Tree struct which is used for resolving HTTP requests paths
	paths map[string]*pathtree.Node
	log   *zap.Logger
}

func newEndpointCache(log *zap.Logger) *endpointCache {
	return &endpointCache{
		paths: map[string]*pathtree.Node{},
		log:   log,
	}
}

func (c *endpointCache) Modified(k string, v []byte) {
	c.log.Debug("Endpoint local cache received value update.", zap.String("key", k), zap.String("value", string(v)))

	e := &subscriptions.Endpoint{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(e)
	if err != nil {
		c.log.Error("Could not deserialize Endpoint state!", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
	} else {
		c.Lock()
		defer c.Unlock()

		root := c.paths[e.Method]
		if root == nil {
			root = pathtree.NewNode()
			c.paths[e.Method] = root
		}
		root.AddRoute(e.Path, e.FunctionID)
	}
}

func (c *endpointCache) Deleted(k string, v []byte) {
	e := &subscriptions.Endpoint{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(e)

	if err != nil {
		c.log.Error("Could not deserialize Endpoint state!", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
	} else {
		c.Lock()
		defer c.Unlock()

		root := c.paths[e.Method]
		if root == nil {
			return
		}
		root.DeleteRoute(e.Path)
	}
}

type subscriptionCache struct {
	sync.RWMutex
	eventToFunctions map[event.Type]map[functions.FunctionID]struct{}
	log              *zap.Logger
}

func newSubscriptionCache(log *zap.Logger) *subscriptionCache {
	return &subscriptionCache{
		eventToFunctions: map[event.Type]map[functions.FunctionID]struct{}{},
		log:              log,
	}
}

func (c *subscriptionCache) Modified(k string, v []byte) {
	c.log.Debug("Subscription local cache received value update.", zap.String("key", k), zap.String("value", string(v)))

	s := subscriptions.Subscription{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&s)
	if err != nil {
		c.log.Error("Could not deserialize Subscription state!", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
		return
	}

	c.Lock()
	defer c.Unlock()

	fnSet, exists := c.eventToFunctions[s.Event]
	if exists {
		fnSet[s.FunctionID] = struct{}{}
	} else {
		fnSet := map[functions.FunctionID]struct{}{}
		fnSet[s.FunctionID] = struct{}{}
		c.eventToFunctions[s.Event] = fnSet
	}
}

func (c *subscriptionCache) Deleted(k string, v []byte) {
	c.Lock()
	defer c.Unlock()

	oldSub := subscriptions.Subscription{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&oldSub)
	if err != nil {
		c.log.Error("Could not deserialize Subscription state during deletion!", zap.Error(err), zap.String("key", k))
		return
	}

	fnSet, exists := c.eventToFunctions[oldSub.Event]
	if exists {
		delete(fnSet, oldSub.FunctionID)

		if len(fnSet) == 0 {
			delete(c.eventToFunctions, oldSub.Event)
		}
	}
}
