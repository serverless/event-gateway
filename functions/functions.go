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

// RegisterFunction registers function in configuration.
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

// GetFunction returns function from configuration.
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

// GetAllFunctions returns an array of all Function
func (f *Functions) GetAllFunctions() ([]*Function, error) {
    fns := []*Function{}

    kvs, err := f.DB.List("")
    // ABD: Should this return the empty list of functions, or the error?
    // Compare to https://github.com/serverless/event-gateway/blob/29d16c3fa36ad8927c90019e5f601e91a9285a9c/pubsub/pubsub.go#L98-L100
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
func (f *Functions) DeleteFunction(name string) error {
	err := f.DB.Delete(name)
	if err != nil {
		return &ErrorNotFound{name}
	}
	return nil
}

func (fn *Function) targetCount() int {
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
	return count
}

func (f *Functions) validateFunction(fn *Function) error {
	if fn.Group != nil {
		if len(fn.Group.Functions) == 0 {
			return &ErrorNoFunctionsProvided{}
		}

		weightTotal := uint(0)
		for _, wf := range fn.Group.Functions {
			weightTotal += wf.Weight
		}

		if weightTotal < 1 {
			return &ErrorTotalFunctionWeightsZero{}
		}
	}

	count := fn.targetCount()

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
