package functions

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/serverless/gateway/db"
	"github.com/serverless/gateway/functions/types"
)

// Functions is a discovery tool for FaaS functions.
type Functions struct {
	DB     *db.PrefixedStore
	Logger *zap.Logger
}

// RegisterFunction registers function in the discovery.
func (f *Functions) RegisterFunction(fn *types.Function) (*types.Function, error) {
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
func (f *Functions) GetFunction(name string) (*types.Function, error) {
	kv, err := f.DB.Get(name)
	if err != nil {
		return nil, &ErrorNotFound{name}
	}

	fn := types.Function{}
	dec := json.NewDecoder(bytes.NewReader(kv.Value))
	err = dec.Decode(&fn)
	if err != nil {
		f.Logger.Info("Fetching function failed.", zap.Error(err))
		return nil, err
	}
	return &fn, nil
}

const providerAWSLambda = "aws-lambda"
