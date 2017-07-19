package pubsub

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/docker/libkv/store"
	"github.com/serverless/event-gateway/functions"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// PubSub allows functions to subscribe to custom events.
type PubSub struct {
	TopicsDB        store.Store
	SubscriptionsDB store.Store
	FunctionsDB     store.Store
	EndpointsDB     store.Store
	Logger          *zap.Logger
}

// CreateSubscription creates subscription.
func (ps PubSub) CreateSubscription(s *Subscription) (*Subscription, error) {
	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		return nil, &ErrorSubscriptionValidation{err}
	}

	s.ID = newSubscriptionID(s)

	_, err = ps.SubscriptionsDB.Get(string(s.ID))
	if err == nil {
		return nil, &ErrorSubscriptionAlreadyExists{
			ID: s.ID,
		}
	}

	if s.TopicID == "http" {
		err = ps.createEndpoint(s.FunctionID, s.Method, strings.TrimPrefix(s.Path, "/"))
		if err != nil {
			return nil, err
		}
	} else {
		err = ps.ensureTopic(s.TopicID)
		if err != nil {
			return nil, err
		}
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

	if sub.TopicID == "http" {
		err = ps.deleteEndpoint(sub.Method, sub.Path)
		if err != nil {
			return err
		}
	} else {
		err = ps.deleteEmptyTopic(sub.TopicID)
		if err != nil {
			return err
		}
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
func (ps PubSub) ensureTopic(id TopicID) error {
	_, err := ps.TopicsDB.Get(string(id))
	if err == nil {
		return nil
	}

	buf, err := json.Marshal(&Topic{ID: id})
	if err != nil {
		return err
	}
	err = ps.TopicsDB.Put(string(id), buf, nil)
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

// createEndpoint creates endpoint.
func (ps PubSub) createEndpoint(functionID functions.FunctionID, method, path string) error {
	e := &Endpoint{
		ID:         newEndpointID(method, path),
		FunctionID: functionID,
		Method:     method,
		Path:       path,
	}

	buf, err := json.Marshal(e)
	if err != nil {
		return err
	}

	err = ps.EndpointsDB.Put(string(e.ID), buf, nil)
	if err != nil {
		return err
	}

	return nil
}

// deleteEndpoint deletes endpoint.
func (ps PubSub) deleteEndpoint(method, path string) error {
	err := ps.EndpointsDB.Delete(string(newEndpointID(method, path)))
	if err != nil {
		return err
	}
	return nil
}
