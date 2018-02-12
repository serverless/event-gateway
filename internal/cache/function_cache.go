package cache

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/serverless/event-gateway/function"
	"go.uber.org/zap"
)

type functionCache struct {
	sync.RWMutex
	cache map[function.ID]*function.Function
	log   *zap.Logger
}

func newFunctionCache(log *zap.Logger) *functionCache {
	return &functionCache{
		cache: map[function.ID]*function.Function{},
		log:   log,
	}
}

func (c *functionCache) Modified(k string, v []byte) {
	c.log.Debug("Function local cache received value update.", zap.String("key", k), zap.String("value", string(v)))

	f := &function.Function{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(f)
	if err != nil {
		c.log.Error("Could not deserialize Function state.", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
	} else {
		c.Lock()
		defer c.Unlock()
		c.cache[function.ID(k)] = f
	}
}

func (c *functionCache) Deleted(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, function.ID(k))
}
