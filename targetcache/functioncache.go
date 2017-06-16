package targetcache

import (
	"bytes"
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	"github.com/serverless/gateway/functions/types"
)

type functionCache struct {
	sync.RWMutex
	cache map[string]types.Function
	log   *zap.Logger
}

func NewFunctionCache(log *zap.Logger) *functionCache {
	return &functionCache{
		cache: map[string]types.Function{},
		log:   log,
	}
}

// Created is called when a new function is detected in the config.
func (f *functionCache) Created(key string, value []byte) {
	f.log.Debug("Received Created function.",
		zap.String("key", key),
		zap.String("value", string(value)))

	fn := types.Function{}
	dec := json.NewDecoder(bytes.NewReader(value))
	err := dec.Decode(&fn)
	if err != nil {
		f.log.Error("Could not deserialize function!.",
			zap.Error(err),
			zap.String("value", string(value)))
	} else {
		f.Lock()
		defer f.Unlock()
		f.cache[key] = fn
	}
}

// Modified is called when an existing function is modified in the config.
func (f *functionCache) Modified(key string, newValue []byte) {
	f.log.Debug("Received Modified function.",
		zap.String("key", key),
		zap.String("newValue", string(newValue)))
	fn := types.Function{}
	dec := json.NewDecoder(bytes.NewReader(newValue))
	err := dec.Decode(&fn)
	if err != nil {
		f.log.Error("Could not deserialize function!.",
			zap.Error(err),
			zap.String("value", string(newValue)))
	} else {
		f.Lock()
		defer f.Unlock()
		f.cache[key] = fn
	}
}

// Deleted is called when a function is deleted in the config.
func (f *functionCache) Deleted(key string, lastKnownValue []byte) {
	f.log.Debug("Received Deleted function.",
		zap.String("key", key),
		zap.String("lastKnownValue", string(lastKnownValue)))
	f.Lock()
	defer f.Unlock()
	delete(f.cache, key)
}
