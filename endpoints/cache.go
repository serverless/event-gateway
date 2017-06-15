package endpoints

import (
	"bytes"
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

type EndpointCache struct {
	sync.RWMutex
	cache map[string]Endpoint
	log   *zap.Logger
}

func NewEndpointCache(log *zap.Logger) EndpointCache {
	return EndpointCache{
		cache: map[string]Endpoint{},
		log:   log,
	}
}

// Created is called when a new endpoint is detected in the config.
func (e *EndpointCache) Created(key string, value []byte) {
	e.log.Debug("Received Created endpoint.",
		zap.String("key", key),
		zap.String("value", string(value)))

	endpoint := Endpoint{}
	dec := json.NewDecoder(bytes.NewReader(value))
	err := dec.Decode(&endpoint)
	if err != nil {
		e.log.Error("Could not deserialize endpoint!.",
			zap.Error(err),
			zap.String("value", string(value)))
	} else {
		e.Lock()
		defer e.Unlock()
		e.cache[key] = endpoint
	}
}

// Modified is called when an existing endpoint is modified in the config.
func (e *EndpointCache) Modified(key string, newValue []byte) {
	e.log.Debug("Received Modified endpoint.",
		zap.String("key", key),
		zap.String("newValue", string(newValue)))
	endpoint := Endpoint{}
	dec := json.NewDecoder(bytes.NewReader(newValue))
	err := dec.Decode(&endpoint)
	if err != nil {
		e.log.Error("Could not deserialize endpoint!.",
			zap.Error(err),
			zap.String("value", string(newValue)))
	} else {
		e.Lock()
		defer e.Unlock()
		e.cache[key] = endpoint
	}
}

// Deleted is called when a endpoint is deleted in the config.
func (e *EndpointCache) Deleted(key string, lastKnownValue []byte) {
	e.log.Debug("Received Deleted endpoint.",
		zap.String("key", key),
		zap.String("lastKnownValue", string(lastKnownValue)))
	e.Lock()
	defer e.Unlock()
	delete(e.cache, key)
}
