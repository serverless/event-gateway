package cache

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/libkv"
	"go.uber.org/zap"
)

type functionCache struct {
	sync.RWMutex
	cache map[libkv.FunctionKey]*function.Function
	log   *zap.Logger
}

func newFunctionCache(log *zap.Logger) *functionCache {
	return &functionCache{
		cache: map[libkv.FunctionKey]*function.Function{},
		log:   log,
	}
}

func (c *functionCache) Modified(k string, v []byte) {
	f := &function.Function{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(f)
	if err != nil {
		c.log.Error("Could not deserialize Function state.", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
		return

	}

	c.log.Debug("Function local cache received value update.", zap.String("key", k), zap.Object("value", f))

	c.Lock()
	defer c.Unlock()
	segments := strings.Split(k, "/")
	c.cache[libkv.FunctionKey{Space: segments[0], ID: function.ID(segments[1])}] = f
}

func (c *functionCache) Deleted(k string, v []byte) {
	c.Lock()
	defer c.Unlock()
	segments := strings.Split(k, "/")
	delete(c.cache, libkv.FunctionKey{Space: segments[0], ID: function.ID(segments[1])})
}
