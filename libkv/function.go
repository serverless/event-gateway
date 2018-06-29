package libkv

import (
	"bytes"
	"encoding/json"
	"regexp"

	validator "gopkg.in/go-playground/validator.v9"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/metadata"
	"github.com/serverless/libkv/store"
)

// FunctionKey is a key under which function data is stored KV store.
type FunctionKey struct {
	Space string
	ID    function.ID
}

func (key FunctionKey) String() string {
	return key.Space + "/" + string(key.ID)
}

// CreateFunction registers function in configuration.
func (service Service) CreateFunction(fn *function.Function) (*function.Function, error) {
	if err := validateFunction(fn); err != nil {
		return nil, err
	}

	_, err := service.FunctionStore.Get(FunctionKey{fn.Space, fn.ID}.String(), &store.ReadOptions{Consistent: true})
	if err == nil {
		return nil, &function.ErrFunctionAlreadyRegistered{ID: fn.ID}
	}

	byt, err := json.Marshal(fn)
	if err != nil {
		return nil, err
	}

	_, _, err = service.FunctionStore.AtomicPut(FunctionKey{fn.Space, fn.ID}.String(), byt, nil, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Function registered.", zap.Object("function", fn))

	return fn, nil
}

// UpdateFunction updates function configuration.
func (service Service) UpdateFunction(fn *function.Function) (*function.Function, error) {
	if err := validateFunction(fn); err != nil {
		return nil, err
	}

	_, err := service.FunctionStore.Get(FunctionKey{fn.Space, fn.ID}.String(), &store.ReadOptions{Consistent: true})
	if err != nil {
		return nil, &function.ErrFunctionNotFound{ID: fn.ID}
	}

	byt, err := json.Marshal(fn)
	if err != nil {
		return nil, &function.ErrFunctionValidation{Message: err.Error()}
	}

	err = service.FunctionStore.Put(FunctionKey{fn.Space, fn.ID}.String(), byt, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Function updated.", zap.Object("function", fn))

	return fn, nil
}

// GetFunction returns function from configuration.
func (service Service) GetFunction(space string, id function.ID) (*function.Function, error) {
	kv, err := service.FunctionStore.Get(FunctionKey{space, id}.String(), &store.ReadOptions{Consistent: true})
	if err != nil {
		if err.Error() == errKeyNotFound {
			return nil, &function.ErrFunctionNotFound{ID: id}
		}
		return nil, err
	}

	fn := function.Function{}
	dec := json.NewDecoder(bytes.NewReader(kv.Value))
	err = dec.Decode(&fn)
	if err != nil {
		return nil, err
	}
	return &fn, nil
}

// ListFunctions returns an array of all functions in the space.
func (service Service) ListFunctions(space string, filters ...metadata.Filter) (function.Functions, error) {
	fns := []*function.Function{}

	kvs, err := service.FunctionStore.List(spacePath(space), &store.ReadOptions{Consistent: true})
	if err != nil && err.Error() != errKeyNotFound {
		return nil, err
	}

	for _, kv := range kvs {
		fn := &function.Function{}
		dec := json.NewDecoder(bytes.NewReader(kv.Value))
		err = dec.Decode(fn)
		if err != nil {
			return nil, err
		}

		if !fn.Metadata.Check(filters...) {
			continue
		}
		fns = append(fns, fn)
	}

	return function.Functions(fns), nil
}

// DeleteFunction deletes function from the registry.
func (service Service) DeleteFunction(space string, id function.ID) error {
	subs, err := service.ListSubscriptions(space)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if id == sub.FunctionID {
			return &function.ErrFunctionHasSubscriptions{}
		}
	}

	eventTypes, err := service.ListEventTypes(space)
	if err != nil {
		return err
	}
	for _, eventType := range eventTypes {
		if id == *eventType.AuthorizerID {
			return &function.ErrFunctionIsAuthorizer{ID: id, EventType: string(eventType.Name)}
		}
	}

	err = service.FunctionStore.Delete(FunctionKey{space, id}.String())
	if err != nil {
		return &function.ErrFunctionNotFound{ID: id}
	}

	service.Log.Debug("Function deleted.", zap.String("space", space), zap.String("functionId", string(id)))

	return nil
}

func validateFunction(fn *function.Function) error {
	if fn.Space == "" {
		fn.Space = defaultSpace
	}

	validate := validator.New()
	validate.RegisterValidation("functionid", functionIDValidator)
	validate.RegisterValidation("space", spaceValidator)
	err := validate.Struct(fn)
	if err != nil {
		return &function.ErrFunctionValidation{Message: err.Error()}
	}

	return nil
}

// functionIDValidator validates if field contains allowed characters for function ID
func functionIDValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`).MatchString(fl.Field().String())
}
