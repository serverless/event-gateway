package subscriptions

import (
	"bytes"
	"encoding/json"

	"github.com/docker/libkv/store"
	"github.com/serverless/event-gateway/functions"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// Subscriptions allows functions to subscribe to custom events.
type Subscriptions struct {
	TopicsDB        store.Store
	SubscriptionsDB store.Store
	FunctionsDB     store.Store
	EndpointsDB     store.Store
	Log             *zap.Logger
}

// CreateSubscription creates subscription.
func (ps Subscriptions) CreateSubscription(s *Subscription) (*Subscription, error) {
	err := ps.validateSubscription(s)
	if err != nil {
		return nil, err
	}

	s.ID = newSubscriptionID(s)
	_, err = ps.SubscriptionsDB.Get(string(s.ID))
	if err == nil {
		return nil, &ErrSubscriptionAlreadyExists{
			ID: s.ID,
		}
	}

	if s.Event == SubscriptionHTTP {
		err = ps.createEndpoint(s.FunctionID, s.Method, s.Path)
		if err != nil {
			return nil, err
		}
	} else {
		err = ps.ensureTopic(s.Event)
		if err != nil {
			return nil, err
		}
	}

	exists, err := ps.FunctionsDB.Exists(string(s.FunctionID))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrFunctionNotFound{string(s.FunctionID)}
	}

	buf, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	err = ps.SubscriptionsDB.Put(string(s.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	ps.Log.Debug("Subscription created.", zap.String("event", string(s.Event)), zap.String("functionId", string(s.FunctionID)))
	return s, nil
}

// DeleteSubscription deletes subscription.
func (ps Subscriptions) DeleteSubscription(id SubscriptionID) error {
	sub, err := ps.getSubscription(id)
	if err != nil {
		return err
	}

	err = ps.SubscriptionsDB.Delete(string(sub.ID))
	if err != nil {
		return &ErrSubscriptionNotFound{sub.ID}
	}

	if sub.Event == SubscriptionHTTP {
		err = ps.deleteEndpoint(sub.Method, sub.Path)
		if err != nil {
			return err
		}
	} else {
		err = ps.deleteEmptyTopic(sub.Event)
		if err != nil {
			return err
		}
	}

	ps.Log.Debug("Subscription deleted.", zap.String("event", string(sub.Event)), zap.String("functionId", string(sub.FunctionID)))

	return nil
}

// GetAllSubscriptions returns array of all Subscription.
func (ps Subscriptions) GetAllSubscriptions() ([]*Subscription, error) {
	subs := []*Subscription{}

	kvs, err := ps.SubscriptionsDB.List("")
	if err != nil {
		return nil, err
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
func (ps Subscriptions) getSubscription(id SubscriptionID) (*Subscription, error) {
	rawsub, err := ps.SubscriptionsDB.Get(string(id))
	if err != nil {
		return nil, &ErrSubscriptionNotFound{id}
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
func (ps Subscriptions) ensureTopic(id TopicID) error {
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
func (ps Subscriptions) deleteEmptyTopic(id TopicID) error {
	subs, err := ps.GetAllSubscriptions()
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if sub.Event == id {
			return nil
		}
	}

	return ps.TopicsDB.Delete(string(id))
}

// createEndpoint creates endpoint.
func (ps Subscriptions) createEndpoint(functionID functions.FunctionID, method, path string) error {
	e := &Endpoint{
		ID:         NewEndpointID(method, path),
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
func (ps Subscriptions) deleteEndpoint(method, path string) error {
	err := ps.EndpointsDB.Delete(string(NewEndpointID(method, path)))
	if err != nil {
		return err
	}
	return nil
}

func (ps Subscriptions) validateSubscription(s *Subscription) error {
	validate := validator.New()
	validate.RegisterValidation("urlpath", urlPathValidator)
	validate.RegisterValidation("eventname", eventNameValidator)
	err := validate.Struct(s)
	if err != nil {
		return &ErrSubscriptionValidation{err.Error()}
	}

	if s.Event == SubscriptionHTTP {
		if s.Method == "" || s.Path == "" {
			return &ErrSubscriptionValidation{"Missing required fields (method, path) for HTTP event."}
		}
	}

	return nil
}
