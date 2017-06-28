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
func (ps PubSub) CreateTopic(t *Topic) (*Topic, error) {
	validate := validator.New()
	err := validate.Struct(t)
	if err != nil {
		return nil, &ErrorValidation{err}
	}

	_, err = ps.TopicsDB.Get(string(t.ID))
	if err == nil {
		return nil, &ErrorAlreadyExists{
			ID: t.ID,
		}
	}

	buf, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	err = ps.TopicsDB.Put(string(t.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// DeleteTopic deletes topic.
func (ps PubSub) DeleteTopic(id TopicID) error {
	err := ps.SubscriptionsDB.DeleteTree("")
	if err != nil {
		return err
	}

	err = ps.TopicsDB.Delete(string(id))
	if err != nil {
		return &ErrorNotFound{id}
	}
	return nil
}

// GetAllTopics returns array of all Topics.
func (ps PubSub) GetAllTopics() ([]*Topic, error) {
	topics := []*Topic{}

	kvs, err := ps.TopicsDB.List("")
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
func (ps PubSub) CreateSubscription(topicID TopicID, s *Subscription) (*Subscription, error) {
	s.ID = subscriptionID(topicID, s.FunctionID)
	s.TopicID = topicID

	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		return nil, &ErrorSubscriptionValidation{err}
	}

	_, err = ps.SubscriptionsDB.Get(string(s.ID))
	if err == nil {
		return nil, &ErrorSubscriptionAlreadyExists{
			ID: s.ID,
		}
	}

	_, err = ps.TopicsDB.Get(string(s.TopicID))
	if err != nil {
		return nil, &ErrorNotFound{s.TopicID}
	}

	exists, err := ps.FunctionsDB.Exists(string(s.FunctionID))
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

	err = ps.SubscriptionsDB.Put(string(s.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// DeleteSubscription deletes subscription.
func (ps PubSub) DeleteSubscription(id SubscriptionID) error {
	err := ps.SubscriptionsDB.Delete(string(id))
	if err != nil {
		return &ErrorSubscriptionNotFound{id}
	}
	return nil
}

// GetAllSubscriptions returns array of all Subscription.
func (ps PubSub) GetAllSubscriptions(tid TopicID) ([]*Subscription, error) {
	subs := []*Subscription{}

	kvs, err := ps.SubscriptionsDB.List("")
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

// CreatePublisher creates a publisher.
func (ps PubSub) CreatePublisher(topicID TopicID, publisher *Publisher) (*Publisher, error) {
	publisher.ID = publisherID(topicID, publisher.Type, publisher.FunctionID)
	publisher.TopicID = topicID

	validate := validator.New()
	err := validate.Struct(publisher)
	if err != nil {
		return nil, &ErrorPublisherValidation{err}
	}

	_, err = ps.PublishersDB.Get(string(publisher.ID))
	if err == nil {
		return nil, &ErrorPublisherAlreadyExists{
			ID: publisher.ID,
		}
	}

	_, err = ps.TopicsDB.Get(string(publisher.TopicID))
	if err != nil {
		return nil, &ErrorNotFound{publisher.TopicID}
	}

	exists, err := ps.FunctionsDB.Exists(string(publisher.FunctionID))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrorFunctionNotFound{string(publisher.FunctionID)}
	}

	buf, err := json.Marshal(publisher)
	if err != nil {
		return nil, err
	}

	err = ps.PublishersDB.Put(string(publisher.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

// DeletePublisher deletes publisher.
func (ps PubSub) DeletePublisher(id PublisherID) error {
	err := ps.PublishersDB.Delete(string(id))
	if err != nil {
		return &ErrorPublisherNotFound{id}
	}
	return nil
}

// GetAllPublishers returns array of all Publishers.
func (ps PubSub) GetAllPublishers(tid TopicID) ([]*Publisher, error) {
	pubs := []*Publisher{}

	kvs, err := ps.PublishersDB.List("")
	if err != nil {
		return pubs, nil
	}

	for _, kv := range kvs {
		if strings.HasPrefix(kv.Key, string(tid)) {
			p := &Publisher{}
			dec := json.NewDecoder(bytes.NewReader(kv.Value))
			err = dec.Decode(p)
			if err != nil {
				return nil, err
			}

			pubs = append(pubs, p)
		}
	}

	return pubs, nil
}
