package pubsub

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// PubSub allows functions to subscribe to custom events.
type PubSub struct {
	TopicsDB        store.Store
	SubscriptionsDB store.Store
	PublishersDB    store.Store
	FunctionsDB     store.Store
	Logger          *zap.Logger
}

// CreateTopic creates topic.
func (p PubSub) CreateTopic(t *Topic) (*Topic, error) {
	validate := validator.New()
	err := validate.Struct(t)
	if err != nil {
		return nil, &ErrorValidation{err}
	}

	_, err = p.TopicsDB.Get(string(t.ID))
	if err == nil {
		return nil, &ErrorAlreadyExists{
			ID: t.ID,
		}
	}

	buf, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	err = p.TopicsDB.Put(string(t.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// DeleteTopic deletes topic.
func (p PubSub) DeleteTopic(id TopicID) error {
	err := p.TopicsDB.Delete(string(id))
	if err != nil {
		return &ErrorNotFound{id}
	}
	return nil
}

// GetAllTopics returns array of all Topics.
func (p PubSub) GetAllTopics() ([]*Topic, error) {
	topics := []*Topic{}

	kvs, err := p.TopicsDB.List("")
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

// CreateSubscription creates subscription.
func (p PubSub) CreateSubscription(topicID TopicID, s *Subscription) (*Subscription, error) {
	s.ID = subscriptionID(topicID, s.FunctionID)
	s.TopicID = topicID

	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		return nil, &ErrorSubscriptionValidation{err}
	}

	_, err = p.SubscriptionsDB.Get(string(s.ID))
	if err == nil {
		return nil, &ErrorSubscriptionAlreadyExists{
			ID: s.ID,
		}
	}

	_, err = p.TopicsDB.Get(string(s.TopicID))
	if err != nil {
		return nil, &ErrorNotFound{s.TopicID}
	}

	exists, err := p.FunctionsDB.Exists(string(s.FunctionID))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrorFunctionNotFound{string(s.FunctionID)}
	}

	buf, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	err = p.SubscriptionsDB.Put(string(s.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// DeleteSubscription deletes subscription.
func (p PubSub) DeleteSubscription(id SubscriptionID) error {
	err := p.SubscriptionsDB.Delete(string(id))
	if err != nil {
		return &ErrorSubscriptionNotFound{id}
	}
	return nil
}

// GetAllSubscriptions returns array of all Subscription.
func (p PubSub) GetAllSubscriptions(tid TopicID) ([]*Subscription, error) {
	subs := []*Subscription{}

	kvs, err := p.SubscriptionsDB.List("")
	if err != nil {
		return subs, nil
	}

	for _, kv := range kvs {
		if strings.HasPrefix(kv.Key, string(tid)) {
			s := &Subscription{}
			dec := json.NewDecoder(bytes.NewReader(kv.Value))
			err = dec.Decode(s)
			if err != nil {
				return nil, err
			}

			subs = append(subs, s)
		}
	}

	return subs, nil
}
