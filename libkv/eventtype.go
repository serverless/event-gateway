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
	if err := service.validateEventType(eventType); err != nil {
		return nil, err
	}

	_, err := service.EventTypeStore.Get(EventTypeKey{eventType.Space, eventType.Name}.String(), &store.ReadOptions{Consistent: true})
	if err == nil {
		return nil, &event.ErrEventTypeAlreadyExists{Name: eventType.Name}
	}

	byt, err := json.Marshal(eventType)
	if err != nil {
		return nil, err
	}

	err = service.EventTypeStore.Put(EventTypeKey{eventType.Space, eventType.Name}.String(), byt, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Event Type created.",
		zap.String("space", eventType.Space),
		zap.String("name", string(eventType.Name)))

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

// DeleteEventType deletes event type from the configuration.
func (service Service) DeleteEventType(space string, name event.TypeName) error {
	subs, err := service.GetSubscriptions(space)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if name == sub.EventType {
			return &event.ErrEventTypeHasSubscriptionsError{}
		}
	}

	err = service.EventTypeStore.Delete(EventTypeKey{space, name}.String())
	if err != nil {
		return &event.ErrEventTypeNotFound{Name: name}
	}

	service.Log.Debug("Event Type deleted.", zap.String("name", string(name)))

	return nil
}

func (service Service) validateEventType(eventType *event.Type) error {
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
