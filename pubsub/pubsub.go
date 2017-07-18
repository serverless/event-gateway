package pubsub

import (
	"bytes"
	"encoding/json"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// PubSub allows functions to subscribe to custom events.
type PubSub struct {
	TopicsDB        store.Store
	SubscriptionsDB store.Store
	FunctionsDB     store.Store
	Logger          *zap.Logger
}

// CreateSubscription creates subscription.
func (ps PubSub) CreateSubscription(s *Subscription) (*Subscription, error) {
	s.ID = subscriptionID(s.TopicID, s.FunctionID)

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

	err = ps.ensureTopic(&Topic{ID: s.TopicID})
	if err != nil {
		return nil, err
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
	sub, err := ps.getSubscription(id)
	if err != nil {
		return err
	}

	err = ps.SubscriptionsDB.Delete(string(sub.ID))
	if err != nil {
		return &ErrorSubscriptionNotFound{sub.ID}
	}

	err = ps.deleteEmptyTopic(sub.TopicID)
	if err != nil {
		return err
	}

	return nil
}

// GetAllSubscriptions returns array of all Subscription.
func (ps PubSub) GetAllSubscriptions() ([]*Subscription, error) {
	subs := []*Subscription{}

	kvs, err := ps.SubscriptionsDB.List("")
	if err != nil {
		return subs, nil
	}

	for _, kv := range kvs {
		s := &Subscription{}
		dec := json.NewDecoder(bytes.NewReader(kv.Value))
		err = dec.Decode(s)
		if err != nil {
			return nil, err
		}

		subs = append(subs, s)
	}

	return subs, nil
}

// getSubscription returns subscription.
func (ps PubSub) getSubscription(id SubscriptionID) (*Subscription, error) {
	rawsub, err := ps.SubscriptionsDB.Get(string(id))
	if err != nil {
		return nil, err
	}

	sub := &Subscription{}
	dec := json.NewDecoder(bytes.NewReader(rawsub.Value))
	err = dec.Decode(sub)
	if err != nil {
		return nil, err
	}

	return sub, err
}

// ensureTopic creates topic if it doesn't exists.
func (ps PubSub) ensureTopic(t *Topic) error {
	_, err := ps.TopicsDB.Get(string(t.ID))
	if err == nil {
		return nil
	}

	buf, err := json.Marshal(t)
	if err != nil {
		return err
	}
	err = ps.TopicsDB.Put(string(t.ID), buf, nil)
	if err != nil {
		return err
	}

	return nil
}

// deleteEmptyTopic deletes topic without subscriptions.
func (ps PubSub) deleteEmptyTopic(id TopicID) error {
	subs, err := ps.GetAllSubscriptions()
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if sub.TopicID == id {
			return nil
		}
	}

	err = ps.TopicsDB.Delete(string(id))
	if err != nil {
		return err
	}
	return nil
}
