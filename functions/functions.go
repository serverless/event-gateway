package functions

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/docker/libkv/store"
	validator "gopkg.in/go-playground/validator.v9"
)

// Functions implements Registry.
type Functions struct {
	DB     store.Store
	Logger *zap.Logger
}

// RegisterFunction registers function in the discovery.
func (f *Functions) RegisterFunction(fn *Function) (*Function, error) {
	if err := f.validateFunction(fn); err != nil {
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

// GetFunction returns function from the discovery.
func (f *Functions) GetFunction(name string) (*Function, error) {
	kv, err := f.DB.Get(name)
	if err != nil {
		return nil, &ErrorNotFound{name}
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

// Exist checks if a function is registed in the discovery.
func (f *Functions) Exist(name string) bool {
	_, err := f.DB.Get(name)
	if err != nil {
		return false
	}

	return true
}

func (f *Functions) validateFunction(fn *Function) error {
	count := 0
	if fn.AWSLambda != nil {
		count++
	}
	if fn.AzureFunction != nil {
		count++
	}
	if fn.GCloudFunction != nil {
		count++
	}
	if fn.OpenWhiskAction != nil {
		count++
	}
	if fn.Group != nil {
		count++
	}
	if fn.HTTP != nil {
		count++
	}

	if count == 0 {
		return &ErrorPropertiesNotSpecified{}
	}

	if count > 1 {
		return &ErrorMoreThanOneFunctionTypeSpecified{}
	}

	validate := validator.New()
	err := validate.Struct(fn)
	if err != nil {
		return &ErrorValidation{err}
	}

	return nil
}
