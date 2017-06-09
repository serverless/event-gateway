package functions

import (
	"bytes"
	"encoding/gob"
	"sync"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/serverless/gateway/db"
)

// Functions is a discovery tool for FaaS functions.
type Functions struct {
	sync.RWMutex
	DB        *db.ReactiveCfgStore
	AWSLambda lambdaiface.LambdaAPI
	Logger    *zap.Logger
}

// Function registered in the function discovery. Function repesents FaaS function deployed on one of the supported providers.
type Function struct {
	ID        string     `json:"id"`
	Instances []Instance `json:"instances"`
}

// Instance of function. A function can have multiple instances. Each instance of a function is deployed in different regions.
type Instance struct {
	Provider string `json:"provider"`
	OriginID string `json:"originId"`
	Region   string `json:"region"`
}

// Created is called when a new function is detected in the config.
func (f *Functions) Created(key string, value []byte) {
	f.Lock()
	defer f.Unlock()
	// TODO put any necessary function conf initialization code here
	f.Logger.Debug("received Created event",
		zap.String("key", key),
		zap.String("value", string(value)))
}

// Modified is called when an existing function is modified in the config.
func (f *Functions) Modified(key string, newValue []byte) {
	f.Lock()
	defer f.Unlock()
	// TODO put any necessary function conf modification code here
	f.Logger.Debug("received Modified event",
		zap.String("key", key),
		zap.String("newValue", string(newValue)))
}

// Deleted is called when a function is deleted in the config.
func (f *Functions) Deleted(key string, lastKnownValue []byte) {
	f.Lock()
	defer f.Unlock()
	// TODO put any necessary function conf deletion code here
	f.Logger.Debug("received Deleted event",
		zap.String("key", key),
		zap.String("lastKnownValue", string(lastKnownValue)))
}

// RegisterFunction registers function in the discovery.
func (f *Functions) RegisterFunction(fn *Function) (*Function, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(fn)
	if err != nil {
		return nil, err
	}

	err = f.DB.Put(fn.ID, buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	return fn, nil
}

// GetFunction returns function from the discovery.
func (f *Functions) GetFunction(name string) (*Function, error) {
	value, err := f.DB.CachedGet(name)
	if err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return nil, &ErrorNotFound{name}
	}

	fn := new(Function)
	buf := bytes.NewBuffer(value)
	err = gob.NewDecoder(buf).Decode(fn)
	if err != nil {
		f.Logger.Info("fetching function failed", zap.Error(err))
		return nil, err
	}
	return fn, nil
}

// Invoke function registered in the discovery.
func (f *Functions) Invoke(name string, payload []byte) ([]byte, error) {
	fn, err := f.GetFunction(name)
	if err != nil {
		return nil, err
	}

	instance := fn.Instances[0]
	if instance.Provider == providerAWSLambda {
		params := &lambda.InvokeInput{
			FunctionName: &instance.OriginID,
			Payload:      payload,
		}
		output, err := f.AWSLambda.Invoke(params)
		if err != nil {
			f.Logger.Info("calling function failed", zap.Error(err))
			return nil, &ErrorInvocationFailed{
				function: *fn,
				instance: instance,
				err:      err,
			}
		}

		return output.Payload, nil
	}

	return nil, nil
}

const providerAWSLambda = "aws-lambda"
