package targetcache

import (
	"bytes"
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	"github.com/serverless/gateway/types"
)

type SubscriberCache struct {
	sync.RWMutex
	cache map[string]types.Subscriber
	log   *zap.Logger
}

func NewSubscriberCache(log *zap.Logger) *SubscriberCache {
	return &SubscriberCache{
		cache: map[string]types.Subscriber{},
		log:   log,
	}
}

// Created is called when a new subscriber is detected in the config.
func (s *SubscriberCache) Created(key string, value []byte) {
	s.log.Debug("Received Created subscriber.",
		zap.String("key", key),
		zap.String("value", string(value)))

	subscriber := types.Subscriber{}
	dec := json.NewDecoder(bytes.NewReader(value))
	err := dec.Decode(&subscriber)
	if err != nil {
		s.log.Error("Could not deserialize subscriber!.",
			zap.Error(err),
			zap.String("value", string(value)))
	} else {
		s.Lock()
		defer s.Unlock()
		s.cache[key] = subscriber
	}
}

// Modified is called when an existing subscriber is modified in the config.
func (s *SubscriberCache) Modified(key string, newValue []byte) {
	s.log.Debug("Received Modified subscriber.",
		zap.String("key", key),
		zap.String("newValue", string(newValue)))
	subscriber := types.Subscriber{}
	dec := json.NewDecoder(bytes.NewReader(newValue))
	err := dec.Decode(&subscriber)
	if err != nil {
		s.log.Error("Could not deserialize subscriber!.",
			zap.Error(err),
			zap.String("value", string(newValue)))
	} else {
		s.Lock()
		defer s.Unlock()
		s.cache[key] = subscriber
	}
}

// Deleted is called when a subscriber is deleted in the config.
func (s *SubscriberCache) Deleted(key string, lastKnownValue []byte) {
	s.log.Debug("Received Deleted subscriber.",
		zap.String("key", key),
		zap.String("lastKnownValue", string(lastKnownValue)))
	s.Lock()
	defer s.Unlock()
	delete(s.cache, key)
}
