package functions

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/docker/libkv/store"
)

// Functions is a discovery tool for FaaS functions.
type Functions struct {
	DB     store.Store
	Logger *zap.Logger
}

// RegisterFunction registers function in the discovery.
func (f *Functions) RegisterFunction(fn *Function) (*Function, error) {
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

const providerAWSLambda = "aws-lambda"
