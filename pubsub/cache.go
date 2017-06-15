package pubsub

import (
	"bytes"
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

type PublisherCache struct {
	sync.RWMutex
	cache map[string]Publisher
	log   *zap.Logger
}

func NewPublisherCache(log *zap.Logger) PublisherCache {
	return PublisherCache{
		cache: map[string]Publisher{},
		log:   log,
	}
}

// Created is called when a new publisher is detected in the config.
func (p *PublisherCache) Created(key string, value []byte) {
	p.log.Debug("Received Created publisher.",
		zap.String("key", key),
		zap.String("value", string(value)))

	publisher := Publisher{}
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
func (p *PublisherCache) Modified(key string, newValue []byte) {
	p.log.Debug("Received Modified publisher.",
		zap.String("key", key),
		zap.String("newValue", string(newValue)))
	publisher := Publisher{}
	dec := json.NewDecoder(bytes.NewReader(value))
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
func (p *PublisherCache) Deleted(key string, lastKnownValue []byte) {
	p.log.Debug("Received Deleted publisher.",
		zap.String("key", key),
		zap.String("lastKnownValue", string(lastKnownValue)))
	p.Lock()
	defer p.Unlock()
	delete(p.cache, key)
}

type SubscriberCache struct {
	sync.RWMutex
	cache map[string]Subscriber
	log   *zap.Logger
}

func NewSubscriberCache(log *zap.Logger) SubscriberCache {
	return SubscriberCache{
		cache: map[string]Subscriber{},
		log:   log,
	}
}

// Created is called when a new subscriber is detected in the config.
func (s *SubscriberCache) Created(key string, value []byte) {
	s.log.Debug("Received Created subscriber.",
		zap.String("key", key),
		zap.String("value", string(value)))

	subscriber := &Subscriber{}
	dec := json.NewDecoder(bytes.NewReader(value))
	err = dec.Decode(subscriber)
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
	subscriber := &Subscriber{}
	dec := json.NewDecoder(bytes.NewReader(value))
	err = dec.Decode(subscriber)
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
