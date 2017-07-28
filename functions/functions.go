package functions

import (
	"bytes"
	"encoding/json"

	validator "gopkg.in/go-playground/validator.v9"

	"go.uber.org/zap"

	"github.com/docker/libkv/store"
)

// Functions implements Registry.
type Functions struct {
	DB     store.Store
	Logger *zap.Logger
}

// RegisterFunction registers function in configuration.
func (f *Functions) RegisterFunction(fn *Function) (*Function, error) {
	_, err := f.DB.Get(string(fn.ID))
	if err == nil {
		return nil, &ErrorAlreadyRegistered{fn.ID}
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

	return fn, nil
}

// UpdateFunction updates function configuration.
func (f *Functions) UpdateFunction(fn *Function) (*Function, error) {
	_, err := f.DB.Get(string(fn.ID))
	if err != nil {
		return nil, &ErrorNotFound{fn.ID}
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

	return fn, nil
}

// GetFunction returns function from configuration.
func (f *Functions) GetFunction(id FunctionID) (*Function, error) {
	kv, err := f.DB.Get(string(id))
	if err != nil {
		return nil, &ErrorNotFound{id}
	}

	fn := Function{}
	dec := json.NewDecoder(bytes.NewReader(kv.Value))
	err = dec.Decode(&fn)
	if err != nil {
		f.Logger.Info("Fetching function failed.", zap.Error(err))
		return nil, err
	}
	return &fn, nil
}

// GetAllFunctions returns an array of all Function
func (f *Functions) GetAllFunctions() ([]*Function, error) {
	fns := []*Function{}

	kvs, err := f.DB.List("")
	if err != nil {
		return nil, err
	}

	for _, kv := range kvs {
		fn := &Function{}
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
func (f *Functions) DeleteFunction(id FunctionID) error {
	err := f.DB.Delete(string(id))
	if err != nil {
		return &ErrorNotFound{id}
	}
	return nil
}

func (f *Functions) validateFunction(fn *Function) error {
	validate := validator.New()
	err := validate.Struct(fn)
	if err != nil {
		return &ErrorValidation{err.Error()}
	}

	if fn.Provider.Type == AWSLambda {
		if fn.Provider.ARN == "" || fn.Provider.Region == "" {
			return &ErrorValidation{"Missing required fields for AWS Lambda function."}
		}
	}

	if fn.Provider.Type == HTTPEndpoint && fn.Provider.URL == "" {
		return &ErrorValidation{"Missing required fields for HTTP endpoint."}
	}

	if fn.Provider.Type == Weighted {
		return f.validateWeighted(fn)
	}

	return nil
}

func (f *Functions) validateWeighted(fn *Function) error {
	if len(fn.Provider.Weighted) == 0 {
		return &ErrorValidation{"Missing required fields for weighted function."}
	}

	weightTotal := uint(0)
	for _, wf := range fn.Provider.Weighted {
		weightTotal += wf.Weight
	}

	if weightTotal < 1 {
		return &ErrorValidation{"Function weights sum to zero."}
	}

	return nil
}
