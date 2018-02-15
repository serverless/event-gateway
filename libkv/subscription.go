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
func (service Service) CreateSubscription(s *subscription.Subscription) (*subscription.Subscription, error) {
	err := service.validateSubscription(s)
	if err != nil {
		return nil, err
	}

	s.ID = newSubscriptionID(s)
	_, err = service.SubscriptionStore.Get(subscriptionPath(s.Space, s.ID), &store.ReadOptions{Consistent: true})
	if err == nil {
		return nil, &subscription.ErrSubscriptionAlreadyExists{
			ID: s.ID,
		}
	}

	if s.Event == event.TypeHTTP {
		err = service.createEndpoint(s.Space, s.Method, s.Path)
		if err != nil {
			return nil, err
		}
	}

	f, err := service.GetFunction(s.Space, s.FunctionID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, &function.ErrFunctionNotFound{ID: s.FunctionID}
	}

	buf, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	err = service.SubscriptionStore.Put(subscriptionPath(s.Space, s.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Subscription created.",
		zap.String("event", string(s.Event)),
		zap.String("space", s.Space),
		zap.String("functionId", string(s.FunctionID)))
	return s, nil
}

// DeleteSubscription deletes subscription.
func (service Service) DeleteSubscription(space string, id subscription.ID) error {
	sub, err := service.GetSubscription(space, id)
	if err != nil {
		return err
	}

	err = service.SubscriptionStore.Delete(subscriptionPath(space, id))
	if err != nil {
		return &subscription.ErrSubscriptionNotFound{ID: sub.ID}
	}

	if sub.Event == event.TypeHTTP {
		err = service.deleteEndpoint(space, sub.Method, sub.Path)
		if err != nil {
			return err
		}
	}

	service.Log.Debug("Subscription deleted.",
		zap.String("event", string(sub.Event)),
		zap.String("space", sub.Space),
		zap.String("functionId", string(sub.FunctionID)))

	return nil
}

// GetSubscriptions returns array of all Subscription.
func (service Service) GetSubscriptions(space string) (subscription.Subscriptions, error) {
	subs := []*subscription.Subscription{}

	kvs, err := service.SubscriptionStore.List(spacePath(space), &store.ReadOptions{Consistent: true})
	if err != nil && err.Error() != errKeyNotFound {
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

	return subscription.Subscriptions(subs), nil
}

// GetSubscription return single subscription.
func (service Service) GetSubscription(space string, id subscription.ID) (*subscription.Subscription, error) {
	rawsub, err := service.SubscriptionStore.Get(subscriptionPath(space, id), &store.ReadOptions{Consistent: true})
	if err != nil {
		if err.Error() == errKeyNotFound {
			return nil, &subscription.ErrSubscriptionNotFound{ID: id}
		}
		return nil, err
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
func (service Service) createEndpoint(space, method, path string) error {
	e := NewEndpoint(method, path)

	kvs, err := service.EndpointStore.List(spacePath(space), &store.ReadOptions{Consistent: true})
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
		tree.AddRoute(sub.Path, sub.Space, function.ID(""), nil)
	}

	err = tree.AddRoute(path, space, function.ID(""), nil)
	if err != nil {
		return &subscription.ErrPathConfict{Message: err.Error()}
	}

	buf, err := json.Marshal(e)
	if err != nil {
		return err
	}
	err = service.EndpointStore.Put(endpointPath(space, e.ID), buf, nil)
	if err != nil {
		return err
	}

	return nil
}

// deleteEndpoint deletes endpoint.
func (service Service) deleteEndpoint(space, method, path string) error {
	err := service.EndpointStore.Delete(endpointPath(space, NewEndpointID(method, path)))
	if err != nil {
		return err
	}
	return nil
}

func (service Service) validateSubscription(s *subscription.Subscription) error {
	if s.Space == "" {
		s.Space = defaultSpace
	}

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
	validate.RegisterValidation("space", spaceValidator)
	err := validate.Struct(s)
	if err != nil {
		return &subscription.ErrSubscriptionValidation{Message: err.Error()}
	}

	if s.Event == event.TypeHTTP && s.Method == "" {
		return &subscription.ErrSubscriptionValidation{Message: "Missing required fields (method, path) for HTTP event."}
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

func subscriptionPath(space string, id subscription.ID) string {
	return spacePath(space) + string(id)
}

func endpointPath(space string, id EndpointID) string {
	return spacePath(space) + string(id)
}

// urlPathValidator validates if field contains URL path
func urlPathValidator(fl validator.FieldLevel) bool {
	return path.IsAbs(fl.Field().String())
}

// eventTypeValidator validates if field contains event name
func eventTypeValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`).MatchString(fl.Field().String())
}
