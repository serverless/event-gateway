package cache

import (
	"bytes"
	"encoding/json"

	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/subscriptions"
	"go.uber.org/zap"
)

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
		c.log.Error("Could not deserialize Endpoint state.", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
	} else {
		c.Lock()
		defer c.Unlock()

		root := c.paths[e.Method]
		if root == nil {
			root = pathtree.NewNode()
			c.paths[e.Method] = root
		}
		err := root.AddRoute(e.Path, e.FunctionID)
		if err != nil {
			c.log.Error("Could not add path to the tree.", zap.Error(err), zap.String("path", e.Path), zap.String("method", e.Method))
		}
	}
}

func (c *endpointCache) Deleted(k string, v []byte) {
	e := &subscriptions.Endpoint{}
	err := json.NewDecoder(bytes.NewReader(v)).Decode(e)

	if err != nil {
		c.log.Error("Could not deserialize Endpoint state.", zap.Error(err), zap.String("key", k), zap.String("value", string(v)))
	} else {
		c.Lock()
		defer c.Unlock()

		root := c.paths[e.Method]
		if root == nil {
			return
		}
		err := root.DeleteRoute(e.Path)
		if err != nil {
			c.log.Error("Could not delete path from the tree.", zap.Error(err), zap.String("path", e.Path), zap.String("method", e.Method))
		}
	}
}
