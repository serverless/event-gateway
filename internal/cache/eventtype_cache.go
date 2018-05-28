package cache

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/libkv"
	"go.uber.org/zap"
)

type eventTypeCache struct {
	sync.RWMutex
	cache map[libkv.EventTypeKey]*eventpkg.Type
	log   *zap.Logger
}

func newEventTypeCache(log *zap.Logger) *eventTypeCache {
	return &eventTypeCache{
		cache: map[libkv.EventTypeKey]*eventpkg.Type{},
		log:   log,
	}
}

func (c *eventTypeCache) Modified(k string, v []byte) {
	eventType := &eventpkg.Type{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(eventType)
	if err != nil {
		c.log.Error("Could not deserialize Event Type state.", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
		return

	}

	c.log.Debug("Event Type local cache received value update.", zap.String("key", k), zap.Object("value", eventType))

	c.Lock()
	defer c.Unlock()
	segments := strings.Split(k, "/")
	c.cache[libkv.EventTypeKey{Space: segments[0], Name: eventpkg.TypeName(segments[1])}] = eventType
}

func (c *eventTypeCache) Deleted(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	segments := strings.Split(k, "/")
	delete(c.cache, libkv.EventTypeKey{Space: segments[0], Name: eventpkg.TypeName(segments[1])})
}
