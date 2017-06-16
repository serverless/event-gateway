package targetcache

import (
	"bytes"
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	"github.com/serverless/gateway/pubsub/types"
)

type publisherCache struct {
	sync.RWMutex
	cache map[string]types.Publisher
	log   *zap.Logger
}

func NewPublisherCache(log *zap.Logger) *publisherCache {
	return &publisherCache{
		cache: map[string]types.Publisher{},
		log:   log,
	}
}

// Created is called when a new publisher is detected in the config.
func (p *publisherCache) Created(key string, value []byte) {
	p.log.Debug("Received Created publisher.",
		zap.String("key", key),
		zap.String("value", string(value)))

	publisher := types.Publisher{}
	dec := json.NewDecoder(bytes.NewReader(value))
	err := dec.Decode(&publisher)
	if err != nil {
		p.log.Error("Could not deserialize publisher!.",
			zap.Error(err),
			zap.String("value", string(value)))
	} else {
		p.Lock()
		defer p.Unlock()
		p.cache[key] = publisher
	}
}

// Modified is called when an existing publisher is modified in the config.
func (p *publisherCache) Modified(key string, newValue []byte) {
	p.log.Debug("Received Modified publisher.",
		zap.String("key", key),
		zap.String("newValue", string(newValue)))
	publisher := types.Publisher{}
	dec := json.NewDecoder(bytes.NewReader(newValue))
	err := dec.Decode(&publisher)
	if err != nil {
		p.log.Error("Could not deserialize publisher!.",
			zap.Error(err),
			zap.String("value", string(newValue)))
	} else {
		p.Lock()
		defer p.Unlock()
		p.cache[key] = publisher
	}
}

// Deleted is called when a publisher is deleted in the config.
func (p *publisherCache) Deleted(key string, lastKnownValue []byte) {
	p.log.Debug("Received Deleted publisher.",
		zap.String("key", key),
		zap.String("lastKnownValue", string(lastKnownValue)))
	p.Lock()
	defer p.Unlock()
	delete(p.cache, key)
}
