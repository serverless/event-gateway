package kv

import (
	"bytes"
	"encoding/json"
	"regexp"

	validator "gopkg.in/go-playground/validator.v9"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/libkv/store"
)

// Functions implements Registry.
type Functions struct {
	DB  store.Store
	Log *zap.Logger
}

// RegisterFunction registers function in configuration.
func (f *Functions) RegisterFunction(fn *function.Function) (*function.Function, error) {
	if err := f.validateFunction(fn); err != nil {
		return nil, err
	}

	_, err := f.DB.Get(string(fn.ID), &store.ReadOptions{Consistent: true})
	if err == nil {
		return nil, &ErrAlreadyRegistered{fn.ID}
	}

	byt, err := json.Marshal(fn)
	if err != nil {
		return nil, err
	}

	err = f.DB.Put(string(fn.ID), byt, nil)
	if err != nil {
		return nil, err
	}

	f.Log.Debug("Function registered.", zap.String("functionId", string(fn.ID)), zap.String("type", string(fn.Provider.Type)))

	return fn, nil
}

// UpdateFunction updates function configuration.
func (f *Functions) UpdateFunction(fn *function.Function) (*function.Function, error) {
	_, err := f.DB.Get(string(fn.ID), &store.ReadOptions{Consistent: true})
	if err != nil {
		return nil, &ErrNotFound{fn.ID}
	}

	if err = f.validateFunction(fn); err != nil {
		return nil, err
	}

	byt, err := json.Marshal(fn)
	if err != nil {
		return nil, err
	}

	err = f.DB.Put(string(fn.ID), byt, nil)
	if err != nil {
		return nil, err
	}

	f.Log.Debug("Function updated.", zap.String("functionId", string(fn.ID)))

	return fn, nil
}

// GetFunction returns function from configuration.
func (f *Functions) GetFunction(id function.ID) (*function.Function, error) {
	kv, err := f.DB.Get(string(id), &store.ReadOptions{Consistent: true})
	if err != nil {
		return nil, &ErrNotFound{id}
	}

	fn := function.Function{}
	dec := json.NewDecoder(bytes.NewReader(kv.Value))
	err = dec.Decode(&fn)
	if err != nil {
		return nil, err
	}
	return &fn, nil
}

// GetAllFunctions returns an array of all Function
func (f *Functions) GetAllFunctions() ([]*function.Function, error) {
	fns := []*function.Function{}

	kvs, err := f.DB.List("", &store.ReadOptions{Consistent: true})
	if err != nil {
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

	return fns, nil
}

// DeleteFunction deletes function from configuration.
func (f *Functions) DeleteFunction(id function.ID) error {
	err := f.DB.Delete(string(id))
	if err != nil {
		return &ErrNotFound{id}
	}

	f.Log.Debug("Function deleted.", zap.String("functionId", string(id)))

	return nil
}

func (f *Functions) validateFunction(fn *function.Function) error {
	validate := validator.New()
	validate.RegisterValidation("functionid", functionIDValidator)
	err := validate.Struct(fn)
	if err != nil {
		return &ErrValidation{err.Error()}
	}

	if fn.Provider.Type == function.AWSLambda {
		if fn.Provider.ARN == "" || fn.Provider.Region == "" {
			return &ErrValidation{"Missing required fields for AWS Lambda function."}
		}
	}

	if fn.Provider.Type == function.Emulator {
		return f.validateEmulator(fn)
	}

	if fn.Provider.Type == function.HTTPEndpoint && fn.Provider.URL == "" {
		return &ErrValidation{"Missing required fields for HTTP endpoint."}
	}

	if fn.Provider.Type == function.Weighted {
		return f.validateWeighted(fn)
	}

	return nil
}

func (f *Functions) validateEmulator(fn *function.Function) error {
	if fn.Provider.EmulatorURL == "" {
		return &ErrValidation{"Missing required field emulatorURL for Emulator function."}
	} else if fn.Provider.APIVersion == "" {
		return &ErrValidation{"Missing required field apiVersion for Emulator function."}
	}
	return nil
}

func (f *Functions) validateWeighted(fn *function.Function) error {
	if len(fn.Provider.Weighted) == 0 {
		return &ErrValidation{"Missing required fields for weighted function."}
	}

	weightTotal := uint(0)
	for _, wf := range fn.Provider.Weighted {
		weightTotal += wf.Weight
	}

	if weightTotal < 1 {
		return &ErrValidation{"Function weights sum to zero."}
	}

	return nil
}

// functionIDValidator validates if field contains allowed characters for function ID
func functionIDValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`).MatchString(fl.Field().String())
}
