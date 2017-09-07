package subscriptions

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/libkv/store"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// Subscriptions allows functions to subscribe to custom events.
type Subscriptions struct {
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

// createEndpoint creates endpoint.
func (ps Subscriptions) createEndpoint(functionID functions.FunctionID, method, path string) error {
	e := NewEndpoint(functionID, method, path)

	kvs, err := ps.EndpointsDB.List("")
	if err != nil {
		return err
	}

	for _, kv := range kvs {
		sub := &Subscription{}
		err = json.NewDecoder(bytes.NewReader(kv.Value)).Decode(sub)
		if err != nil {
			return err
		}

		if sub.Method == method && isPathInConflict(sub.Path, path) {
			return &ErrPathConfict{sub.Path, path}
		}
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
	if s.Event == SubscriptionHTTP {
		s.Path = ensurePrefix(s.Path, "/")
		s.Method = strings.ToUpper(s.Method)
	}

	validate := validator.New()
	validate.RegisterValidation("urlpath", urlPathValidator)
	validate.RegisterValidation("eventtype", eventTypeValidator)
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

// ensurePrefix ensures s starts with prefix.
func ensurePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s
	}
	return prefix + s
}

func toSegments(route string) []string {
	segments := strings.Split(route, "/")
	// remove first "" element
	_, segments = segments[0], segments[1:]

	return segments
}

// nolint: gocyclo
func isPathInConflict(existing, new string) bool {
	existingSegments := toSegments(existing)
	newSegments := toSegments(new)

	for i, newSegment := range newSegments {
		// no segment at this stage, no issue
		if len(existingSegments) < i+1 {
			return false
		}

		existing := existingSegments[i]
		existingIsParam := strings.HasPrefix(existing, ":")
		existingIsWildcard := strings.HasPrefix(existing, "*")
		newIsParam := strings.HasPrefix(newSegment, ":")
		newIsWildcard := strings.HasPrefix(newSegment, "*")

		// both segments static
		if !existingIsParam && !existingIsWildcard && !newIsParam && !newIsWildcard {
			// static are the same and it's the end of the path
			if existing == newSegment && len(existingSegments) == i+1 {
				return false
			}

			continue
		}

		if existingIsWildcard {
			return true
		}

		// different parameters
		if existingIsParam && newIsParam && existing != newSegment {
			return true
		}
	}

	return true
}
