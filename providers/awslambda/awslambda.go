package awslambda

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/serverless/event-gateway/function"
	"go.uber.org/zap/zapcore"
	validator "gopkg.in/go-playground/validator.v9"
)

// Type of provider.
const Type = function.ProviderType("awslambda")

func init() {
	function.RegisterProvider(Type, ProviderLoader{})
}

// AWSLambda function implementation
type AWSLambda struct {
	Service lambdaiface.LambdaAPI `json:"-" validate:"-"`

	ARN                string `json:"arn" validate:"required"`
	Region             string `json:"region" validate:"required"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSSessionToken    string `json:"awsSessionToken,omitempty"`
}

// Call AWS Lambda function.
func (a AWSLambda) Call(payload []byte) ([]byte, error) {
	invokeOutput, err := a.Service.Invoke(&lambda.InvokeInput{
		FunctionName: &a.ARN,
		Payload:      payload,
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			switch awserr.Code() {
			case "AccessDeniedException":
				return nil, function.ErrFunctionCallFailed{Original: err, Message: "Function call failed with AccessDeniedException. The provided credentials do not" +
					" have the required IAM permissions to invoke this function. Please attach the" +
					" lambda:invokeFunction permission to these credentials."}
			case "ExpiredTokenException":
				return nil, function.ErrFunctionCallFailed{Original: err, Message: "Function call failed with ExpiredTokenException. The provided security token for" +
					" the function has expired. Please provide an updated security token or provide" +
					" permanent credentials."}
			case "UnrecognizedClientException":
				return nil, function.ErrFunctionCallFailed{Original: err, Message: "Function call failed with UnrecognizedClientException. The provided credentials" +
					" are invalid. Please provide valid credentials."}
			case lambda.ErrCodeServiceException:
				return nil, function.ErrFunctionCallFailed{Original: err, Message: "Function call failed with ServiceException. AWS Lambda service wasn't" +
					" able to handle the request."}
			default:
				return nil, function.ErrFunctionCallFailed{Original: err, Message: "Function call failed. Please check logs."}
			}
		}
	}

	if invokeOutput.FunctionError != nil {
		return nil, function.ErrFunctionError{Original: errors.New(*invokeOutput.FunctionError)}
	}

	return invokeOutput.Payload, err
}

// validate provider config.
func (a AWSLambda) validate() error {
	validate := validator.New()
	err := validate.Struct(a)
	if err != nil {
		return err
	}
	return nil
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface.
func (a AWSLambda) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("arn", a.ARN)
	enc.AddString("region", a.Region)
	if a.AWSAccessKeyID != "" {
		enc.AddString("awsAccessKeyId", "*****")
	}
	if a.AWSSecretAccessKey != "" {
		enc.AddString("awsSecretAccessKey", "*****")
	}
	if a.AWSSessionToken != "" {
		enc.AddString("awsSessionToken", "*****")
	}
	return nil
}

// ProviderLoader implementation
type ProviderLoader struct{}

// Load decode JSON data as Config and return initialized Provider instance.
func (p ProviderLoader) Load(data []byte) (function.Provider, error) {
	provider := &AWSLambda{}
	err := json.Unmarshal(data, provider)
	if err != nil {
		return nil, errors.New("unable to load function provider config: " + err.Error())
	}

	err = provider.validate()
	if err != nil {
		return nil, errors.New("missing required fields for AWS Lambda function")
	}

	config := aws.NewConfig().WithRegion(provider.Region)
	if provider.AWSAccessKeyID != "" && provider.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(provider.AWSAccessKeyID, provider.AWSSecretAccessKey, provider.AWSSessionToken))
	}

	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, errors.New("unable to create AWS Session: " + err.Error())
	}

	provider.Service = lambda.New(awsSession)
	return provider, nil
}
