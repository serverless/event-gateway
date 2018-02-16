package libkv

import (
	"bytes"
	"encoding/json"
	"regexp"

	validator "gopkg.in/go-playground/validator.v9"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/function"
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

// RegisterFunction registers function in configuration.
func (service Service) RegisterFunction(fn *function.Function) (*function.Function, error) {
	if err := service.validateFunction(fn); err != nil {
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

	err = service.FunctionStore.Put(FunctionKey{fn.Space, fn.ID}.String(), byt, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Function registered.",
		zap.String("space", fn.Space),
		zap.String("functionId", string(fn.ID)),
		zap.String("type", string(fn.Provider.Type)))

	return fn, nil
}

// UpdateFunction updates function configuration.
func (service Service) UpdateFunction(space string, fn *function.Function) (*function.Function, error) {
	_, err := service.FunctionStore.Get(FunctionKey{space, fn.ID}.String(), &store.ReadOptions{Consistent: true})
	if err != nil {
		return nil, &function.ErrFunctionNotFound{ID: fn.ID}
	}

	if err = service.validateFunction(fn); err != nil {
		return nil, err
	}

	byt, err := json.Marshal(fn)
	if err != nil {
		return nil, err
	}

	err = service.FunctionStore.Put(FunctionKey{fn.Space, fn.ID}.String(), byt, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("Function updated.",
		zap.String("space", fn.Space),
		zap.String("functionId", string(fn.ID)))

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

// GetFunctions returns an array of all functions in the space.
func (service Service) GetFunctions(space string) (function.Functions, error) {
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

		fns = append(fns, fn)
	}

	return function.Functions(fns), nil
}

// DeleteFunction deletes function from the registry.
func (service Service) DeleteFunction(space string, id function.ID) error {
	subs, err := service.GetSubscriptions(space)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if id == sub.FunctionID {
			return &function.ErrFunctionHasSubscriptionsError{}
		}
	}

	err = service.FunctionStore.Delete(FunctionKey{space, id}.String())
	if err != nil {
		return &function.ErrFunctionNotFound{ID: id}
	}

	service.Log.Debug("Function deleted.", zap.String("functionId", string(id)))

	return nil
}

func (service Service) validateFunction(fn *function.Function) error {
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

	if fn.Provider.Type == function.AWSLambda {
		if fn.Provider.ARN == "" || fn.Provider.Region == "" {
			return &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."}
		}
	}

	if fn.Provider.Type == function.Emulator {
		return service.validateEmulator(fn)
	}

	if fn.Provider.Type == function.HTTPEndpoint && fn.Provider.URL == "" {
		return &function.ErrFunctionValidation{Message: "Missing required fields for HTTP endpoint."}
	}

	if fn.Provider.Type == function.Weighted {
		return service.validateWeighted(fn)
	}

	return nil
}

func (service Service) validateEmulator(fn *function.Function) error {
	if fn.Provider.EmulatorURL == "" {
		return &function.ErrFunctionValidation{Message: "Missing required field emulatorURL for Emulator function."}
	} else if fn.Provider.APIVersion == "" {
		return &function.ErrFunctionValidation{Message: "Missing required field apiVersion for Emulator function."}
	}
	return nil
}

func (service Service) validateWeighted(fn *function.Function) error {
	if len(fn.Provider.Weighted) == 0 {
		return &function.ErrFunctionValidation{Message: "Missing required fields for weighted function."}
	}

	weightTotal := uint(0)
	for _, wf := range fn.Provider.Weighted {
		weightTotal += wf.Weight
	}

	if weightTotal < 1 {
		return &function.ErrFunctionValidation{Message: "Function weights sum to zero."}
	}

	return nil
}

// functionIDValidator validates if field contains allowed characters for function ID
func functionIDValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`).MatchString(fl.Field().String())
}
