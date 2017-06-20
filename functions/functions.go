package functions

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/docker/libkv/store"
	validator "gopkg.in/go-playground/validator.v9"
)

// Functions is a discovery tool for FaaS functions.
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
		f.Logger.Info("Marshalling function payload failed.", zap.Error(err))
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

func (f *Functions) validateFunction(fn *Function) error {
	validate := validator.New()
	err := validate.Struct(fn)
	if err != nil {
		return &ErrorValidation{err}
	}

	if fn.AWSLambda != nil {
		fn.Type = AWSLambda
	} else if fn.AzureFunction != nil {
		fn.Type = AzureFunction
	} else if fn.GCloudFunction != nil {
		fn.Type = GCloudFunction
	} else if fn.OpenWhiskAction != nil {
		fn.Type = OpenWhiskAction
	} else if fn.Group != nil {
		fn.Type = Group
	} else if fn.HTTP != nil {
		fn.Type = HTTP
	}

	if fn.Type == 0 {
		return &ErrorPropertiesNotSpecified{}
	}

	return nil
}
