package libkv

import (
	"bytes"
	"encoding/json"

	validator "gopkg.in/go-playground/validator.v9"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/libkv/store"
)

// EventTypeKey is a key under which event type data is stored KV store.
type EventTypeKey struct {
	Space string
	Name  event.TypeName
}

func (key EventTypeKey) String() string {
	return key.Space + "/" + string(key.Name)
}

// CreateEventType creates event type in configuration.
func (service Service) CreateEventType(eventType *event.Type) (*event.Type, error) {
	if err := validateEventType(eventType); err != nil {
		return nil, err
	}

	_, err := service.EventTypeStore.Get(EventTypeKey{eventType.Space, eventType.Name}.String(), &store.ReadOptions{Consistent: true})
	if err == nil {
		return nil, &event.ErrEventTypeAlreadyExists{Name: eventType.Name}
	}

	if eventType.AuthorizerID != nil {
		function, _ := service.GetFunction(eventType.Space, *eventType.AuthorizerID)
		if function == nil {
			return nil, &event.ErrAuthorizerDoesNotExists{}
		}
	}

	byt, err := json.Marshal(eventType)
	if err != nil {
		return nil, &event.ErrEventTypeValidation{Message: err.Error()}
	}

	_, _, err = service.EventTypeStore.AtomicPut(EventTypeKey{eventType.Space, eventType.Name}.String(), byt, nil, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Event Type created.", zap.Object("eventType", eventType))

	return eventType, nil
}

// GetEventType returns function from configuration.
func (service Service) GetEventType(space string, name event.TypeName) (*event.Type, error) {
	kv, err := service.EventTypeStore.Get(EventTypeKey{space, name}.String(), &store.ReadOptions{Consistent: true})
	if err != nil {
		if err.Error() == errKeyNotFound {
			return nil, &event.ErrEventTypeNotFound{Name: name}
		}
		return nil, err
	}

	eventType := event.Type{}
	dec := json.NewDecoder(bytes.NewReader(kv.Value))
	err = dec.Decode(&eventType)
	if err != nil {
		return nil, err
	}
	return &eventType, nil
}

// GetEventTypes returns an array of all event types in the space.
func (service Service) GetEventTypes(space string) (event.Types, error) {
	types := []*event.Type{}

	kvs, err := service.EventTypeStore.List(spacePath(space), &store.ReadOptions{Consistent: true})
	if err != nil && err.Error() != errKeyNotFound {
		return nil, err
	}

	for _, kv := range kvs {
		eventType := &event.Type{}
		dec := json.NewDecoder(bytes.NewReader(kv.Value))
		err = dec.Decode(eventType)
		if err != nil {
			return nil, err
		}

		types = append(types, eventType)
	}

	return event.Types(types), nil
}

// UpdateEventType updates subscription.
func (service Service) UpdateEventType(newEventType *event.Type) (*event.Type, error) {
	if err := validateEventType(newEventType); err != nil {
		return nil, err
	}

	_, err := service.GetEventType(newEventType.Space, newEventType.Name)
	if err != nil {
		return nil, err
	}

	if newEventType.AuthorizerID != nil {
		function, _ := service.GetFunction(newEventType.Space, *newEventType.AuthorizerID)
		if function == nil {
			return nil, &event.ErrAuthorizerDoesNotExists{}
		}
	}

	buf, err := json.Marshal(newEventType)
	if err != nil {
		return nil, &event.ErrEventTypeValidation{Message: err.Error()}
	}

	err = service.EventTypeStore.Put(EventTypeKey{newEventType.Space, newEventType.Name}.String(), buf, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Event Type updated.", zap.Object("eventType", newEventType))

	return newEventType, nil
}

// DeleteEventType deletes event type from the configuration.
func (service Service) DeleteEventType(space string, name event.TypeName) error {
	subs, err := service.GetSubscriptions(space)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if name == sub.EventType {
			return &event.ErrEventTypeHasSubscriptions{}
		}
	}

	err = service.EventTypeStore.Delete(EventTypeKey{space, name}.String())
	if err != nil {
		return &event.ErrEventTypeNotFound{Name: name}
	}

	service.Log.Debug("Event Type deleted.", zap.String("space", space), zap.String("name", string(name)))

	return nil
}

func validateEventType(eventType *event.Type) error {
	if eventType.Space == "" {
		eventType.Space = defaultSpace
	}

	validate := validator.New()
	validate.RegisterValidation("space", spaceValidator)
	err := validate.Struct(eventType)
	if err != nil {
		return &event.ErrEventTypeValidation{Message: err.Error()}
	}

	return nil
}
