package awsfirehose_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/awsfirehose"
	"github.com/serverless/event-gateway/providers/awsfirehose/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLoad(t *testing.T) {
	for _, testCase := range loadTests {
		config := []byte(testCase.config)
		loader := awsfirehose.ProviderLoader{}

		_, err := loader.Load(config)

		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestCall(t *testing.T) {
	for _, testCase := range callTests {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		serviceMock := mock.NewMockFirehoseAPI(mockCtrl)
		serviceMock.EXPECT().PutRecord(gomock.Any()).Return(testCase.putRecordResult, testCase.putRecordError)

		provider := awsfirehose.AWSFirehose{
			Service:            serviceMock,
			DeliveryStreamName: "teststream",
			Region:             "us-east-1",
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
		`{"streamName": "", "region": `,
		errors.New("unable to load function provider config: unexpected end of JSON input"),
	},
	{
		`{"streamName": "", "region": "us-east-1"}`,
		errors.New("missing required fields for AWS Firehose function"),
	},
	{
		`{"streamName": "test", "region": ""}`,
		errors.New("missing required fields for AWS Firehose function"),
	},
}

var callTests = []struct {
	putRecordResult *firehose.PutRecordOutput
	putRecordError  error
	expectedResult  []byte
	expectedError   error
}{
	{
		&firehose.PutRecordOutput{RecordId: aws.String("testid")},
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
		awsfirehose.AWSFirehose{
			DeliveryStreamName: "test",
			Region:             "us-east-1",
		},
		map[string]interface{}{
			"deliveryStreamName": "test",
			"region":             "us-east-1",
		},
	},
	{
		awsfirehose.AWSFirehose{
			AWSAccessKeyID:     "id",
			AWSSecretAccessKey: "key",
			AWSSessionToken:    "token",
		},
		map[string]interface{}{
			"deliveryStreamName": "",
			"region":             "",
			"awsAccessKeyId":     "*****",
			"awsSecretAccessKey": "*****",
			"awsSessionToken":    "*****",
		},
	},
}
