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

func (f *Functions) validateFunction(fn *Function) error {
	validate := validator.New()
	err := validate.Struct(fn)
	if err != nil {
		return &ErrorValidation{err}
	}

	if fn.Provider.Type == Weighted {
		if len(fn.Provider.Weighted) == 0 {
			return &ErrorNoFunctionsProvided{}
		}

		weightTotal := uint(0)
		for _, wf := range fn.Provider.Weighted {
			weightTotal += wf.Weight
		}

		if weightTotal < 1 {
			return &ErrorTotalFunctionWeightsZero{}
		}
	}

	return nil
}
