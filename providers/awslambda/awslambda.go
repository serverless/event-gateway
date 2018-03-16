package awslambda

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
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
	ARN                string `json:"arn" validate:"required"`
	Region             string `json:"region" validate:"required"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSSessionToken    string `json:"awsSessionToken,omitempty"`
}

// Call AWS Lambda function.
// nolint: gocyclo
func (a AWSLambda) Call(payload []byte) ([]byte, error) {
	config := aws.NewConfig().WithRegion(a.Region)
	if a.AWSAccessKeyID != "" && a.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(a.AWSAccessKeyID, a.AWSSecretAccessKey, a.AWSSessionToken))
	}

	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, &function.ErrFunctionProviderError{Original: err}
	}
	awslambda := lambda.New(awsSession)

	invokeOutput, err := awslambda.Invoke(&lambda.InvokeInput{
		FunctionName: &a.ARN,
		Payload:      payload,
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			switch awserr.Code() {
			case "AccessDeniedException",
				"ExpiredTokenException",
				"UnrecognizedClientException":
				return nil, &function.ErrFunctionAccessDenied{Original: awserr}
			case lambda.ErrCodeServiceException:
				return nil, &function.ErrFunctionProviderError{Original: awserr}
			default:
				return nil, &function.ErrFunctionCallFailed{Original: awserr}
			}
		}
	}

	if invokeOutput.FunctionError != nil {
		return nil, &function.ErrFunctionError{Original: errors.New(*invokeOutput.FunctionError)}
	}

	return invokeOutput.Payload, err
}

// Validate provider config.
func (a AWSLambda) Validate() error {
	validate := validator.New()
	err := validate.Struct(a)
	if err != nil {
		return &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."}
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

	return provider, nil
}
