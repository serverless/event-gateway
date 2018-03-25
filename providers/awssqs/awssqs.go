package awssqs

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/serverless/event-gateway/function"
	"go.uber.org/zap/zapcore"
	validator "gopkg.in/go-playground/validator.v9"
)

// Type of provider.
const Type = function.ProviderType("awssqs")

func init() {
	function.RegisterProvider(Type, ProviderLoader{})
}

// AWSSQS function implementation
type AWSSQS struct {
	Service sqsiface.SQSAPI`json:"-" validate:"-"`

	QueueURL           string `json:"queueUrl" validate:"required"`
	Region             string `json:"region" validate:"required"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSSessionToken    string `json:"awsSessionToken,omitempty"`
}

// Call sends message to AWS SQS Queue
func (a AWSSQS) Call(payload []byte) ([]byte, error) {
    body := string(payload)
	sendMessageOutput, err := a.Service.SendMessage(&sqs.SendMessageInput{
		QueueUrl:       &a.QueueURL,
		MessageBody:    &body,
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			return nil, &function.ErrFunctionCallFailed{Original: awserr}
		}
	}

	return []byte(*sendMessageOutput.MessageId), err
}

// validate provider config.
func (a AWSSQS) validate() error {
	validate := validator.New()
	err := validate.Struct(a)
	if err != nil {
		return err
	}
	return nil
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface.
func (a AWSSQS) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("queueUrl", a.QueueURL)
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
	provider := &AWSSQS{}
	err := json.Unmarshal(data, provider)
	if err != nil {
		return nil, errors.New("unable to load function provider config: " + err.Error())
	}

	err = provider.validate()
	if err != nil {
		return nil, errors.New("missing required fields for AWS SQS function")
	}

	config := aws.NewConfig().WithRegion(provider.Region)
	if provider.AWSAccessKeyID != "" && provider.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(provider.AWSAccessKeyID, provider.AWSSecretAccessKey, provider.AWSSessionToken))
	}

	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, errors.New("unable to create AWS Session: " + err.Error())
	}

	provider.Service = sqs.New(awsSession)
	return provider, nil
}
