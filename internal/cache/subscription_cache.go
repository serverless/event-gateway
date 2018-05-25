package cache

import (
	"bytes"
	"encoding/json"
	"sync"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/libkv"
	"github.com/serverless/event-gateway/subscription"
	"go.uber.org/zap"
)

type subscriptionCache struct {
	sync.RWMutex
	// eventToFunctions maps path and event type to function key (space + function ID)
	eventToFunctions map[string]map[eventpkg.Type][]libkv.FunctionKey
	// endpoints maps HTTP method to internal/pathtree. Tree struct which is used for resolving HTTP requests paths.
	endpoints map[string]*pathtree.Node
	log       *zap.Logger
}

func newSubscriptionCache(log *zap.Logger) *subscriptionCache {
	return &subscriptionCache{
		eventToFunctions: map[string]map[eventpkg.Type][]libkv.FunctionKey{},
		endpoints:        map[string]*pathtree.Node{},
		log:              log,
	}
}

func (c *subscriptionCache) Modified(k string, v []byte) {
	s := subscription.Subscription{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&s)
	if err != nil {
		c.log.Error("Could not deserialize Subscription state.", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
		return
	}

	c.log.Debug("Subscription local cache received value update.", zap.String("key", k), zap.Object("value", s))

	c.Lock()
	defer c.Unlock()
	key := libkv.FunctionKey{Space: s.Space, ID: s.FunctionID}

	if s.Type == subscription.TypeSync {
		root := c.endpoints[s.Method]
		if root == nil {
			root = pathtree.NewNode()
			c.endpoints[s.Method] = root
		}
		err := root.AddRoute(s.Path, s.Space, s.FunctionID, s.CORS)
		if err != nil {
			c.log.Error("Could not add path to the tree.", zap.Error(err), zap.String("path", s.Path), zap.String("method", s.Method))
		}
	} else {
		c.createPath(s.Path)
		ids, exists := c.eventToFunctions[s.Path][s.EventType]
		if exists {
			ids = append(ids, key)
		} else {
			ids = []libkv.FunctionKey{key}
		}
		c.eventToFunctions[s.Path][s.EventType] = ids
	}
}

func (c *subscriptionCache) Deleted(k string, v []byte) {
	c.Lock()
	defer c.Unlock()

	oldSub := subscription.Subscription{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&oldSub)
	if err != nil {
		c.log.Error("Could not deserialize Subscription state during deletion.", zap.Error(err), zap.String("key", k))
		return
	}

	if oldSub.Type == subscription.TypeSync {
		c.deleteEndpoint(oldSub)
	} else {
		c.deleteSubscription(oldSub)
	}
}

func (c *subscriptionCache) createPath(path string) {
	_, exists := c.eventToFunctions[path]
	if !exists {
		c.eventToFunctions[path] = map[eventpkg.Type][]libkv.FunctionKey{}
	}
}

func (c *subscriptionCache) deleteEndpoint(sub subscription.Subscription) {
	root := c.endpoints[sub.Method]
	if root == nil {
		return
	}
	err := root.DeleteRoute(sub.Path)
	if err != nil {
		c.log.Error("Could not delete path from the tree.", zap.Error(err), zap.String("path", sub.Path), zap.String("method", sub.Method))
	}
}

func (c *subscriptionCache) deleteSubscription(sub subscription.Subscription) {
	ids, exists := c.eventToFunctions[sub.Path][sub.EventType]
	if exists {
		for i, id := range ids {
			key := libkv.FunctionKey{Space: sub.Space, ID: sub.FunctionID}
			if id == key {
				ids = append(ids[:i], ids[i+1:]...)
				break
			}
		}
		c.eventToFunctions[sub.Path][sub.EventType] = ids

		if len(ids) == 0 {
			delete(c.eventToFunctions[sub.Path], sub.EventType)
		}
	}
}
