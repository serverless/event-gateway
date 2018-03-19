package awskinesis

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/serverless/event-gateway/function"
	"go.uber.org/zap/zapcore"
	validator "gopkg.in/go-playground/validator.v9"
)

// Type of provider.
const Type = function.ProviderType("awskinesis")

func init() {
	function.RegisterProvider(Type, ProviderLoader{})
}

// AWSKinesis function implementation
type AWSKinesis struct {
	Service kinesisiface.KinesisAPI `json:"-" validate:"-"`

	StreamName         string `json:"streamName" validate:"required"`
	Region             string `json:"region" validate:"required"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSSessionToken    string `json:"awsSessionToken,omitempty"`
}

// Call puts record into AWS Kinesis stream.
func (a AWSKinesis) Call(payload []byte) ([]byte, error) {
	putRecordOutput, err := a.Service.PutRecord(&kinesis.PutRecordInput{
		StreamName:   &a.StreamName,
		Data:         payload,
		PartitionKey: aws.String("123"),
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			return nil, &function.ErrFunctionCallFailed{Original: awserr}
		}
	}

	return []byte(*putRecordOutput.SequenceNumber), err
}

// validate provider config.
func (a AWSKinesis) validate() error {
	validate := validator.New()
	err := validate.Struct(a)
	if err != nil {
		return err
	}
	return nil
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface.
func (a AWSKinesis) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("streamName", a.StreamName)
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
	provider := &AWSKinesis{}
	err := json.Unmarshal(data, provider)
	if err != nil {
		return nil, &function.ErrFunctionValidation{Message: "Unable to load function provider config: " + err.Error()}
	}

	err = provider.validate()
	if err != nil {
		return nil, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Kinesis function."}
	}

	config := aws.NewConfig().WithRegion(provider.Region)
	if provider.AWSAccessKeyID != "" && provider.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(provider.AWSAccessKeyID, provider.AWSSecretAccessKey, provider.AWSSessionToken))
	}

	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, &function.ErrFunctionValidation{Message: "Unable to create AWS Session: " + err.Error()}
	}

	provider.Service = kinesis.New(awsSession)
	return provider, nil
}
