package functions

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/serverless/gateway/db"
)

// Functions is a discovery tool for FaaS functions.
type Functions struct {
	DB     *db.ReactiveCfgStore
	Logger *zap.Logger
}

// Function registered in the function discovery. Function repesents FaaS function deployed on one of the supported providers.
type Function struct {
	ID        string     `json:"id"`
	Instances []Instance `json:"instances"`
}

// Instance of function. A function can have multiple instances. Each instance of a function is deployed in different regions.
type Instance struct {
	Provider    string      `json:"provider"`
	OriginID    string      `json:"originId"`
	Region      string      `json:"region"`
	Credentials Credentials `json:"credentials"`
}

// Credentials that allows calling user's function.
type Credentials struct {
	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
}

// Created is called when a new function is detected in the config.
func (f *Functions) Created(key string, value []byte) {
	f.Logger.Debug("Received Created event.",
		zap.String("key", key),
		zap.String("value", string(value)))
}

// Modified is called when an existing function is modified in the config.
func (f *Functions) Modified(key string, newValue []byte) {
	f.Logger.Debug("Received Modified event.",
		zap.String("key", key),
		zap.String("newValue", string(newValue)))
}

// Deleted is called when a function is deleted in the config.
func (f *Functions) Deleted(key string, lastKnownValue []byte) {
	f.Logger.Debug("Received Deleted event.",
		zap.String("key", key),
		zap.String("lastKnownValue", string(lastKnownValue)))
}

// RegisterFunction registers function in the discovery.
func (f *Functions) RegisterFunction(fn *Function) (*Function, error) {
	byt, err := json.Marshal(fn)
	if err != nil {
		return nil, err
	}

	err = f.DB.Put(fn.ID, byt, nil)
	if err != nil {
		return nil, err
	}

	return fn, nil
}

// GetFunction returns function from the discovery.
func (f *Functions) GetFunction(name string) (*Function, error) {
	value, err := f.DB.CachedGet(name)
	if err != nil {
		return nil, &ErrorNotFound{name}
	}

	fn := &Function{}
	dec := json.NewDecoder(bytes.NewReader(value))
	err = dec.Decode(fn)
	if err != nil {
		f.Logger.Info("Fetching function failed.", zap.Error(err))
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
		creds := credentials.NewStaticCredentials(instance.Credentials.AWSAccessKeyID, instance.Credentials.AWSSecretAccessKey, "")
		awslambda := lambda.New(session.New(aws.NewConfig().WithRegion(instance.Region).WithCredentials(creds)))

		output, err := awslambda.Invoke(&lambda.InvokeInput{
			FunctionName: &instance.OriginID,
			Payload:      payload,
		})
		if err != nil {
			f.Logger.Info("Calling function failed.", zap.Error(err))
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
