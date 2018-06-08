package libkv

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/internal/pathtree"
	istrings "github.com/serverless/event-gateway/internal/strings"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/libkv/store"
	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

// CreateSubscription creates subscription.
func (service Service) CreateSubscription(sub *subscription.Subscription) (*subscription.Subscription, error) {
	err := validateSubscription(sub)
	if err != nil {
		return nil, err
	}
	sub.ID = newSubscriptionID(sub)
	_, err = service.SubscriptionStore.Get(subscriptionPath(sub.Space, sub.ID), &store.ReadOptions{Consistent: true})
	if err == nil {
		return nil, &subscription.ErrSubscriptionAlreadyExists{
			ID: sub.ID,
		}
	}

	if sub.Type == subscription.TypeSync {
		err = service.checkForPathConflict(sub.Space, sub.Method, sub.Path)
		if err != nil {
			return nil, err
		}
	}

	_, err = service.GetEventType(sub.Space, sub.EventType)
	if err != nil {
		return nil, err
	}

	_, err = service.GetFunction(sub.Space, sub.FunctionID)
	if err != nil {
		return nil, err
	}

	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, err
	}

	_, _, err = service.SubscriptionStore.AtomicPut(subscriptionPath(sub.Space, sub.ID), buf, nil, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Subscription created.", zap.Object("subscription", sub))
	return sub, nil
}

// UpdateSubscription updates subscription.
func (service Service) UpdateSubscription(id subscription.ID, newSub *subscription.Subscription) (*subscription.Subscription, error) {
	if err := validateSubscription(newSub); err != nil {
		return nil, err
	}

	oldSub, err := service.GetSubscription(newSub.Space, id)
	if err != nil {
		return nil, err
	}

	err = validateSubscriptionUpdate(newSub, oldSub)
	if err != nil {
		return nil, err
	}

	_, err = service.GetFunction(newSub.Space, newSub.FunctionID)
	if err != nil {
		return nil, err
	}

	buf, err := json.Marshal(newSub)
	if err != nil {
		return nil, &subscription.ErrSubscriptionValidation{Message: err.Error()}
	}

	err = service.SubscriptionStore.Put(subscriptionPath(newSub.Space, newSub.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Subscription updated.", zap.Object("subscription", newSub))
	return newSub, nil
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

	service.Log.Debug("Subscription deleted.", zap.Object("subscription", sub))
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

func (service Service) checkForPathConflict(space, method, path string) error {
	kvs, err := service.SubscriptionStore.List(spacePath(space), &store.ReadOptions{Consistent: true})

	tree := pathtree.NewNode()

	for _, kv := range kvs {
		sub := &subscription.Subscription{}
		err = json.NewDecoder(bytes.NewReader(kv.Value)).Decode(sub)
		if err != nil {
			return err
		}

		if sub.Type == subscription.TypeSync {
			// add existing paths to check
			tree.AddRoute(sub.Path, sub.Space, function.ID(""), nil)
		}
	}

	err = tree.AddRoute(path, space, function.ID(""), nil)
	if err != nil {
		return &subscription.ErrPathConfict{Message: err.Error()}
	}

	return nil
}

func validateSubscription(sub *subscription.Subscription) error {
	sub.Path = istrings.EnsurePrefix(sub.Path, "/")

	if sub.Space == "" {
		sub.Space = defaultSpace
	}

	if sub.Method == "" {
		sub.Method = http.MethodPost
	} else {
		sub.Method = strings.ToUpper(sub.Method)
	}

	if sub.Type == subscription.TypeAsync && sub.CORS != nil {
		return &subscription.ErrSubscriptionValidation{Message: "CORS can be configured only for sync subscriptions."}
	}

	if sub.CORS != nil {
		if sub.CORS.Headers == nil {
			sub.CORS.Headers = []string{"Origin", "Accept", "Content-Type"}
		}

		if sub.CORS.Methods == nil {
			sub.CORS.Methods = []string{"HEAD", "GET", "POST"}
		}

		if sub.CORS.Origins == nil {
			sub.CORS.Origins = []string{"*"}
		}
	}

	validate := validator.New()
	validate.RegisterValidation("urlPath", urlPathValidator)
	validate.RegisterValidation("eventType", eventTypeValidator)
	validate.RegisterValidation("space", spaceValidator)
	err := validate.Struct(sub)
	if err != nil {
		return &subscription.ErrSubscriptionValidation{Message: err.Error()}
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

func newSubscriptionID(sub *subscription.Subscription) subscription.ID {
	raw := string(sub.Type) + "," + string(sub.EventType) + "," + string(sub.FunctionID) + "," + url.PathEscape(sub.Path) + "," + sub.Method
	return subscription.ID(base64.RawURLEncoding.EncodeToString([]byte(raw)))
}

func subscriptionPath(space string, id subscription.ID) string {
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

func validateSubscriptionUpdate(newSub *subscription.Subscription, oldSub *subscription.Subscription) error {
	if newSub.Type != oldSub.Type {
		return &subscription.ErrInvalidSubscriptionUpdate{Field: "Type"}
	}
	if newSub.EventType != oldSub.EventType {
		return &subscription.ErrInvalidSubscriptionUpdate{Field: "EventType"}
	}
	if newSub.FunctionID != oldSub.FunctionID {
		return &subscription.ErrInvalidSubscriptionUpdate{Field: "FunctionID"}
	}
	if newSub.Path != oldSub.Path {
		return &subscription.ErrInvalidSubscriptionUpdate{Field: "Path"}
	}
	if newSub.Method != oldSub.Method {
		return &subscription.ErrInvalidSubscriptionUpdate{Field: "Method"}
	}

	return nil
}
