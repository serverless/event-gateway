package cache

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/subscriptions"
	"go.uber.org/zap"
)

type subscriptionCache struct {
	sync.RWMutex
	eventToFunctions map[string]map[event.Type][]functions.FunctionID
	log              *zap.Logger
}

func newSubscriptionCache(log *zap.Logger) *subscriptionCache {
	return &subscriptionCache{
		eventToFunctions: map[string]map[event.Type][]functions.FunctionID{},
		log:              log,
	}
}

func (c *subscriptionCache) Modified(k string, v []byte) {
	c.log.Debug("Subscription local cache received value update.", zap.String("key", k), zap.String("value", string(v)))

	s := subscriptions.Subscription{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&s)
	if err != nil {
		c.log.Error("Could not deserialize Subscription state.", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
		return
	}
	c.Lock()
	defer c.Unlock()

	c.makeSpace(s.Space)
	ids, exists := c.eventToFunctions[s.Space][s.Event]
	if exists {
		ids = append(ids, s.FunctionID)
	} else {
		ids = []functions.FunctionID{s.FunctionID}
	}
	c.eventToFunctions[s.Space][s.Event] = ids
}

func (c *subscriptionCache) Deleted(k string, v []byte) {
	c.Lock()
	defer c.Unlock()

	oldSub := subscriptions.Subscription{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&oldSub)
	if err != nil {
		c.log.Error("Could not deserialize Subscription state during deletion.", zap.Error(err), zap.String("key", k))
		return
	}

	ids, exists := c.eventToFunctions[oldSub.Space][oldSub.Event]
	if exists {
		for i, id := range ids {
			if id == oldSub.FunctionID {
				ids = append(ids[:i], ids[i+1:]...)
				break
			}
		}
		c.eventToFunctions[oldSub.Space][oldSub.Event] = ids

		if len(ids) == 0 {
			delete(c.eventToFunctions[oldSub.Space], oldSub.Event)
		}
	}
}

func (c *subscriptionCache) makeSpace(space string) {
	_, exists := c.eventToFunctions[space]
	if !exists {
		c.eventToFunctions[space] = map[event.Type][]functions.FunctionID{}
	}
}
