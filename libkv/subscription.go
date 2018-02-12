package libkv

import (
	"bytes"
	"encoding/json"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/internal/pathtree"
	istrings "github.com/serverless/event-gateway/internal/strings"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/libkv/store"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// CreateSubscription creates subscription.
func (ps Service) CreateSubscription(s *subscription.Subscription) (*subscription.Subscription, error) {
	err := ps.validateSubscription(s)
	if err != nil {
		return nil, err
	}

	s.ID = newSubscriptionID(s)
	_, err = ps.SubscriptionStore.Get(string(s.ID), &store.ReadOptions{Consistent: true})
	if err == nil {
		return nil, &ErrSubscriptionAlreadyExists{
			ID: s.ID,
		}
	}

	if s.Event == event.TypeHTTP {
		err = ps.createEndpoint(s.Method, s.Path)
		if err != nil {
			return nil, err
		}
	}

	exists, err := ps.FunctionStore.Exists(string(s.FunctionID), &store.ReadOptions{Consistent: true})
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

	err = ps.SubscriptionStore.Put(string(s.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	ps.Log.Debug("Subscription created.", zap.String("event", string(s.Event)), zap.String("functionId", string(s.FunctionID)))
	return s, nil
}

// DeleteSubscription deletes subscription.
func (ps Service) DeleteSubscription(id subscription.ID) error {
	sub, err := ps.getSubscription(id)
	if err != nil {
		return err
	}

	err = ps.SubscriptionStore.Delete(string(sub.ID))
	if err != nil {
		return &ErrSubscriptionNotFound{sub.ID}
	}

	if sub.Event == event.TypeHTTP {
		err = ps.deleteEndpoint(sub.Method, sub.Path)
		if err != nil {
			return err
		}
	}

	ps.Log.Debug("Subscription deleted.", zap.String("event", string(sub.Event)), zap.String("functionId", string(sub.FunctionID)))

	return nil
}

// GetAllSubscriptions returns array of all Subscription.
func (ps Service) GetAllSubscriptions() ([]*subscription.Subscription, error) {
	subs := []*subscription.Subscription{}

	kvs, err := ps.SubscriptionStore.List("", &store.ReadOptions{Consistent: true})
	if err != nil {
		return nil, err
	}

	for _, kv := range kvs {
		s := &subscription.Subscription{}
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
func (ps Service) getSubscription(id subscription.ID) (*subscription.Subscription, error) {
	rawsub, err := ps.SubscriptionStore.Get(string(id), &store.ReadOptions{Consistent: true})
	if err != nil {
		return nil, &ErrSubscriptionNotFound{id}
	}

	sub := &subscription.Subscription{}
	dec := json.NewDecoder(bytes.NewReader(rawsub.Value))
	err = dec.Decode(sub)
	if err != nil {
		return nil, err
	}

	return sub, err
}

// createEndpoint creates endpoint.
func (ps Service) createEndpoint(method, path string) error {
	e := NewEndpoint(method, path)

	kvs, err := ps.EndpointStore.List("", &store.ReadOptions{Consistent: true})
	// We need to check for not found key as there is no Endpoint cached that creates the directory.
	if err != nil && err.Error() != "Key not found in store" {
		return err
	}

	tree := pathtree.NewNode()

	for _, kv := range kvs {
		sub := &subscription.Subscription{}
		err = json.NewDecoder(bytes.NewReader(kv.Value)).Decode(sub)
		if err != nil {
			return err
		}

		// add existing paths to check
		tree.AddRoute(sub.Path, function.ID(""), nil)
	}

	err = tree.AddRoute(path, function.ID(""), nil)
	if err != nil {
		return &ErrPathConfict{err.Error()}
	}

	buf, err := json.Marshal(e)
	if err != nil {
		return err
	}
	err = ps.EndpointStore.Put(string(e.ID), buf, nil)
	if err != nil {
		return err
	}

	return nil
}

// deleteEndpoint deletes endpoint.
func (ps Service) deleteEndpoint(method, path string) error {
	err := ps.EndpointStore.Delete(string(NewEndpointID(method, path)))
	if err != nil {
		return err
	}
	return nil
}

func (ps Service) validateSubscription(s *subscription.Subscription) error {
	s.Path = istrings.EnsurePrefix(s.Path, "/")
	if s.Event == event.TypeHTTP {
		s.Method = strings.ToUpper(s.Method)

		if s.CORS != nil {
			if s.CORS.Headers == nil {
				s.CORS.Headers = []string{"Origin", "Accept", "Content-Type"}
			}

			if s.CORS.Methods == nil {
				s.CORS.Methods = []string{"HEAD", "GET", "POST"}
			}

			if s.CORS.Origins == nil {
				s.CORS.Origins = []string{"*"}
			}
		}
	}

	validate := validator.New()
	validate.RegisterValidation("urlpath", urlPathValidator)
	validate.RegisterValidation("eventtype", eventTypeValidator)
	err := validate.Struct(s)
	if err != nil {
		return &ErrSubscriptionValidation{err.Error()}
	}

	if s.Event == event.TypeHTTP && s.Method == "" {
		return &ErrSubscriptionValidation{"Missing required fields (method, path) for HTTP event."}
	}

	return nil
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

		if existingIsParam && !newIsParam {
			return true
		}
	}

	return true
}

func newSubscriptionID(s *subscription.Subscription) subscription.ID {
	if s.Event == event.TypeHTTP {
		return subscription.ID(string(s.Event) + "," + s.Method + "," + url.PathEscape(s.Path))
	}
	return subscription.ID(string(s.Event) + "," + string(s.FunctionID) + "," + url.PathEscape(s.Path))
}

// urlPathValidator validates if field contains URL path
func urlPathValidator(fl validator.FieldLevel) bool {
	return path.IsAbs(fl.Field().String())
}

// eventTypeValidator validates if field contains event name
func eventTypeValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`).MatchString(fl.Field().String())
}
