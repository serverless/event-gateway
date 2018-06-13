package cache

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/subscription/cors"
	"go.uber.org/zap"
)

type corsCache struct {
	sync.RWMutex
	endpoints map[string]*pathtree.Node
	log       *zap.Logger
}

func newCORSCache(log *zap.Logger) *corsCache {
	return &corsCache{
		endpoints: map[string]*pathtree.Node{},
		log:       log,
	}
}

func (c *corsCache) Modified(k string, v []byte) {
	config := cors.CORS{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&config)
	if err != nil {
		c.log.Error("Could not deserialize CORS state.", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
		return
	}

	c.log.Debug("CORS local cache received value update.", zap.String("key", k), zap.Object("value", config))

	c.Lock()
	defer c.Unlock()

	root := c.endpoints[config.Method]
	if root == nil {
		root = pathtree.NewNode()
		c.endpoints[config.Method] = root
	}
	err = root.AddRoute(config.Path, config)
	if err != nil {
		c.log.Error("Could not add path to the tree.", zap.Error(err), zap.String("path", config.Path), zap.String("method", config.Method))
	}
}

func (c *corsCache) Deleted(k string, v []byte) {
	c.Lock()
	defer c.Unlock()

	config := cors.CORS{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(&config)
	if err != nil {
		c.log.Error("Could not deserialize CORS state during deletion.", zap.Error(err), zap.String("key", k))
		return
	}

	c.deleteEndpoint(config)
}

func (c *corsCache) deleteEndpoint(config cors.CORS) {
	root := c.endpoints[config.Method]
	if root == nil {
		return
	}
	err := root.DeleteRoute(config.Path)
	if err != nil {
		c.log.Error("Could not delete path from the tree.", zap.Error(err), zap.String("path", config.Path), zap.String("method", config.Method))
	}
}
