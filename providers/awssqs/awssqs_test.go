package awssqs_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/awssqs"
	"github.com/serverless/event-gateway/providers/awssqs/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLoad(t *testing.T) {
	for _, testCase := range loadTests {
		config := []byte(testCase.config)
		loader := awssqs.ProviderLoader{}

		_, err := loader.Load(config)

		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestCall(t *testing.T) {
	for _, testCase := range callTests {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		serviceMock := mock.NewMockSQSAPI(mockCtrl)
		serviceMock.EXPECT().SendMessage(gomock.Any()).Return(testCase.sendMessageResult, testCase.sendMessageError)

		provider := awssqs.AWSSQS{
			Service:    serviceMock,
			QueueURL:   "https://sqs.us-east-2.amazonaws.com/123456789012/MyQueue",
			Region:     "us-east-1",
		}

		output, err := provider.Call([]byte("testpayload"))

		assert.Equal(t, testCase.expectedResult, output)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestMarshalLogObject(t *testing.T) {
	for _, testCase := range logTests {
		enc := zapcore.NewMapObjectEncoder()

		testCase.provider.MarshalLogObject(enc)

		assert.Equal(t, testCase.expectedFields, enc.Fields)
	}
}

var loadTests = []struct {
	config        string
	expectedError error
}{
	{
		`{"queueUrl": "", "region": `,
		errors.New("unable to load function provider config: unexpected end of JSON input"),
	},
	{
		`{"queueUrl": "", "region": "us-east-1"}`,
		errors.New("missing required fields for AWS SQS function"),
	},
	{
		`{"queueUrl": "https://sqs.us-east-2.amazonaws.com/123456789012/MyQueue", "region": ""}`,
		errors.New("missing required fields for AWS SQS function"),
	},
}

var callTests = []struct {
	sendMessageResult *sqs.SendMessageOutput
	sendMessageError  error
	expectedResult  []byte
	expectedError   error
}{
	{
		&sqs.SendMessageOutput{MessageId: aws.String("testid")},
		nil,
		[]byte("testid"),
		nil,
	},
	{
		nil,
		awserr.New("", "", nil),
		[]byte(nil),
		&function.ErrFunctionCallFailed{Original: awserr.New("", "", nil)},
	},
}

var logTests = []struct {
	provider       function.Provider
	expectedFields map[string]interface{}
}{
	{
		awssqs.AWSSQS{
			QueueURL  : "https://sqs.us-east-2.amazonaws.com/123456789012/MyQueue",
			Region:     "us-east-1",
		},
		map[string]interface{}{
			"queueUrl": "https://sqs.us-east-2.amazonaws.com/123456789012/MyQueue",
			"region":     "us-east-1",
		},
	},
	{
		awssqs.AWSSQS{
			AWSAccessKeyID:     "id",
			AWSSecretAccessKey: "key",
			AWSSessionToken:    "token",
		},
		map[string]interface{}{
			"queueUrl":           "",
			"region":             "",
			"awsAccessKeyId":     "*****",
			"awsSecretAccessKey": "*****",
			"awsSessionToken":    "*****",
		},
	},
}
