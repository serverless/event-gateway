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
