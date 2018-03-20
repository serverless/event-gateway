package awsfirehose

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/firehose/firehoseiface"
	"github.com/serverless/event-gateway/function"
	"go.uber.org/zap/zapcore"
	validator "gopkg.in/go-playground/validator.v9"
)

// Type of provider.
const Type = function.ProviderType("awsfirehose")

func init() {
	function.RegisterProvider(Type, ProviderLoader{})
}

// AWSFirehose function implementation
type AWSFirehose struct {
	Service firehoseiface.FirehoseAPI `json:"-" validate:"-"`

	DeliveryStreamName         string `json:"deliveryStreamName" validate:"required"`
	Region                     string `json:"region" validate:"required"`
	AWSAccessKeyID             string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey         string `json:"awsSecretAccessKey,omitempty"`
	AWSSessionToken            string `json:"awsSessionToken,omitempty"`
}

// Call puts record into AWS Firehose stream.
func (a AWSFirehose) Call(payload []byte) ([]byte, error) {
	putRecordOutput, err := a.Service.PutRecord(&firehose.PutRecordInput{
		DeliveryStreamName:   &a.DeliveryStreamName,
		Record:     &firehose.Record{Data: payload},
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			return nil, &function.ErrFunctionCallFailed{Original: awserr}
		}
	}

	return []byte(*putRecordOutput.RecordId), err
}

// validate provider config.
func (a AWSFirehose) validate() error {
	validate := validator.New()
	err := validate.Struct(a)
	if err != nil {
		return err
	}
	return nil
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface.
func (a AWSFirehose) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("deliveryStreamName", a.DeliveryStreamName)
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
	provider := &AWSFirehose{}
	err := json.Unmarshal(data, provider)
	if err != nil {
		return nil, &function.ErrFunctionValidation{Message: "Unable to load function provider config: " + err.Error()}
	}

	err = provider.validate()
	if err != nil {
		return nil, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Firehose function."}
	}

	config := aws.NewConfig().WithRegion(provider.Region)
	if provider.AWSAccessKeyID != "" && provider.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(provider.AWSAccessKeyID, provider.AWSSecretAccessKey, provider.AWSSessionToken))
	}

	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, &function.ErrFunctionValidation{Message: "Unable to create AWS Session: " + err.Error()}
	}

	provider.Service = firehose.New(awsSession)
	return provider, nil
}
