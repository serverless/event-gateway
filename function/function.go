package function

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"go.uber.org/zap/zapcore"
)

// Function represents a function deployed on one of the supported providers.
type Function struct {
	Space    string    `json:"space" validate:"required,min=3,space"`
	ID       ID        `json:"functionId" validate:"required,functionid"`
	Provider *Provider `json:"provider" validate:"required"`
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (f Function) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("space", string(f.Space))
	enc.AddString("functionId", string(f.ID))
	if f.Provider != nil {
		enc.AddObject("provider", f.Provider)
	}

	return nil
}

// Functions is an array of functions.
type Functions []*Function

// ID uniquely identifies a function.
type ID string

// Provider provides provider specific info about a function.
type Provider struct {
	Type ProviderType `json:"type" validate:"required,eq=awslambda|eq=http"`

	// AWS Lambda function
	ARN                string `json:"arn,omitempty"`
	Region             string `json:"region,omitempty"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSSessionToken    string `json:"awsSessionToken,omitempty"`

	// HTTP function
	URL string `json:"url,omitempty" validate:"omitempty,url"`
}

// ProviderType represents what kind of function provider this is.
type ProviderType string

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (p Provider) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", string(p.Type))
	if p.ARN != "" {
		enc.AddString("arn", p.ARN)
	}
	if p.AWSAccessKeyID != "" {
		enc.AddString("awsAccessKeyId", "*****")
	}
	if p.AWSSecretAccessKey != "" {
		enc.AddString("awsSecretAccessKey", "*****")
	}
	if p.AWSSessionToken != "" {
		enc.AddString("awsSessionToken", "*****")
	}
	if p.URL != "" {
		enc.AddString("url", p.URL)
	}

	return nil
}

const (
	// AWSLambda represents AWS Lambda function.
	AWSLambda ProviderType = "awslambda"
	// HTTPEndpoint represents function defined as HTTP endpoint.
	HTTPEndpoint ProviderType = "http"
)

// Call tries to send a payload to a target function
func (f *Function) Call(payload []byte) ([]byte, error) {
	switch f.Provider.Type {
	case AWSLambda:
		return f.callAWSLambda(payload)
	case HTTPEndpoint:
		return f.callHTTP(payload)
	}

	return []byte{}, errors.New("calling this kind of function is not implemented")
}

// nolint: gocyclo
func (f *Function) callAWSLambda(payload []byte) ([]byte, error) {
	config := aws.NewConfig().WithRegion(f.Provider.Region)
	if f.Provider.AWSAccessKeyID != "" && f.Provider.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(f.Provider.AWSAccessKeyID, f.Provider.AWSSecretAccessKey, f.Provider.AWSSessionToken))
	}

	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, &ErrFunctionProviderError{err}
	}
	awslambda := lambda.New(awsSession)

	invokeOutput, err := awslambda.Invoke(&lambda.InvokeInput{
		FunctionName: &f.Provider.ARN,
		Payload:      payload,
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			switch awserr.Code() {
			case "AccessDeniedException",
				"ExpiredTokenException",
				"UnrecognizedClientException":
				return nil, &ErrFunctionAccessDenied{awserr}
			case lambda.ErrCodeServiceException:
				return nil, &ErrFunctionProviderError{awserr}
			default:
				return nil, &ErrFunctionCallFailed{awserr}
			}
		}
	}

	if invokeOutput.FunctionError != nil {
		return nil, &ErrFunctionError{errors.New(*invokeOutput.FunctionError)}
	}

	return invokeOutput.Payload, err
}

func (f *Function) callHTTP(payload []byte) ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Post(f.Provider.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, &ErrFunctionCallFailed{err}
	}
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, &ErrFunctionError{fmt.Errorf("HTTP status code: %d", http.StatusInternalServerError)}
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
