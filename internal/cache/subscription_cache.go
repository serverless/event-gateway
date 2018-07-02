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
	// async maps method, path and event type to function key (space + function ID) (async subscriptions)
	async map[string]map[string]map[eventpkg.TypeName][]libkv.FunctionKey
	// sync maps method and event type to internal/pathtree (sync subscriptions)
	sync map[string]map[eventpkg.TypeName]*pathtree.Node
	log  *zap.Logger
}

func newSubscriptionCache(log *zap.Logger) *subscriptionCache {
	return &subscriptionCache{
		async: map[string]map[string]map[eventpkg.TypeName][]libkv.FunctionKey{},
		sync:  map[string]map[eventpkg.TypeName]*pathtree.Node{},
		log:   log,
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
		c.ensureSyncMethod(s.Method)
		root := c.sync[s.Method][s.EventType]
		if root == nil {
			root = pathtree.NewNode()
			c.sync[s.Method][s.EventType] = root
		}
		err := root.AddRoute(s.Path, libkv.FunctionKey{Space: s.Space, ID: s.FunctionID})
		if err != nil {
			c.log.Error("Could not add path to the tree.", zap.Error(err), zap.String("path", s.Path), zap.String("method", s.Method), zap.String("eventType", string(s.EventType)))
		}
	} else {
		c.ensureAsyncMethodPath(s.Method, s.Path)
		ids, exists := c.async[s.Method][s.Path][s.EventType]
		if exists {
			ids = append(ids, key)
		} else {
			ids = []libkv.FunctionKey{key}
		}
		c.async[s.Method][s.Path][s.EventType] = ids
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

func (c *subscriptionCache) ensureAsyncMethodPath(method, path string) {
	_, exists := c.async[method]
	if !exists {
		c.async[method] = map[string]map[eventpkg.TypeName][]libkv.FunctionKey{}
	}

	_, exists = c.async[method][path]
	if !exists {
		c.async[method][path] = map[eventpkg.TypeName][]libkv.FunctionKey{}
	}
}

func (c *subscriptionCache) ensureSyncMethod(method string) {
	_, exists := c.sync[method]
	if !exists {
		c.sync[method] = map[eventpkg.TypeName]*pathtree.Node{}
	}
}

func (c *subscriptionCache) deleteEndpoint(sub subscription.Subscription) {
	root := c.sync[sub.Method][sub.EventType]
	if root == nil {
		return
	}
	err := root.DeleteRoute(sub.Path)
	if err != nil {
		c.log.Error("Could not delete path from the tree.", zap.Error(err), zap.String("path", sub.Path), zap.String("method", sub.Method))
	}
}

func (c *subscriptionCache) deleteSubscription(sub subscription.Subscription) {
	ids, exists := c.async[sub.Method][sub.Path][sub.EventType]
	if exists {
		for i, id := range ids {
			key := libkv.FunctionKey{Space: sub.Space, ID: sub.FunctionID}
			if id == key {
				ids = append(ids[:i], ids[i+1:]...)
				break
			}
		}
		c.async[sub.Method][sub.Path][sub.EventType] = ids

		if len(ids) == 0 {
			delete(c.async[sub.Method][sub.Path], sub.EventType)
		}
	}
}
