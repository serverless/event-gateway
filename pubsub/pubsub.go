package pubsub

import (
	"bytes"
	"encoding/json"

	"github.com/docker/libkv/store"
	"github.com/serverless/event-gateway/functions"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// TopicID uniquely identifies a pubsub topic
type TopicID string

// Topic allows stores events that function can subsribe to.
type Topic struct {
	ID TopicID `json:"topicId" validate:"required"`
}

// PublisherID uniquely identifies a publisher function
type PublisherID string

// Subscriber maps from TopicID to FunctionID
type Subscriber struct {
	TopicID    TopicID
	FunctionID functions.FunctionID
}

// FunctionEnd is used to specify whether the input or output
// from a function is to be used.
type FunctionEnd uint

const (
	// Input means the input to a function should feed a topic
	Input FunctionEnd = iota
	// Output means the output of a function should feed a topic
	Output
)

// Publisher maps from {input,output} + FunctionID to TopicID
type Publisher struct {
	FunctionEnd FunctionEnd
	FunctionID  functions.FunctionID
	TopicID     TopicID
}

// PubSub allows functions to subscribe to custom events.
type PubSub struct {
	DB     store.Store
	Logger *zap.Logger
}

// Create topic.
func (p PubSub) Create(t *Topic) (*Topic, error) {
	validate := validator.New()
	err := validate.Struct(t)
	if err != nil {
		return nil, &ErrorValidation{err}
	}

	_, err = p.DB.Get(string(t.ID))
	if err == nil {
		return nil, &ErrorAlreadyExists{
			ID: t.ID,
		}
	}

	buf, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	err = p.DB.Put(string(t.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Delete topic.
func (p PubSub) Delete(id TopicID) error {
	err := p.DB.Delete(string(id))
	if err != nil {
		return &ErrorNotFound{string(id)}
	}
	return nil
}

// GetAll returns array of all Topics.
func (p PubSub) GetAll() ([]*Topic, error) {
	topics := []*Topic{}

	kvs, err := p.DB.List("")
	if err != nil {
		return topics, nil
	}

	for _, kv := range kvs {
		t := &Topic{}
		dec := json.NewDecoder(bytes.NewReader(kv.Value))
		err = dec.Decode(t)
		if err != nil {
			return nil, err
		}

		topics = append(topics, t)
	}

	return topics, nil
}
