package awskinesis_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/awskinesis"
	"github.com/serverless/event-gateway/providers/awskinesis/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLoad(t *testing.T) {
	for _, testCase := range loadTests {
		config := []byte(testCase.config)
		loader := awskinesis.ProviderLoader{}

		_, err := loader.Load(config)

		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestCall(t *testing.T) {
	for _, testCase := range callTests {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		serviceMock := mock.NewMockKinesisAPI(mockCtrl)
		serviceMock.EXPECT().PutRecord(gomock.Any()).Return(testCase.putRecordResult, testCase.putRecordError)

		provider := awskinesis.AWSKinesis{
			Service:    serviceMock,
			StreamName: "teststream",
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
		`{"streamName": "", "region": `,
		errors.New("unable to load function provider config: unexpected end of JSON input"),
	},
	{
		`{"streamName": "", "region": "us-east-1"}`,
		errors.New("missing required fields for AWS Kinesis function"),
	},
	{
		`{"streamName": "test", "region": ""}`,
		errors.New("missing required fields for AWS Kinesis function"),
	},
}

var callTests = []struct {
	putRecordResult *kinesis.PutRecordOutput
	putRecordError  error
	expectedResult  []byte
	expectedError   error
}{
	{
		&kinesis.PutRecordOutput{SequenceNumber: aws.String("testseq")},
		nil,
		[]byte("testseq"),
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
		awskinesis.AWSKinesis{
			StreamName: "test",
			Region:     "us-east-1",
		},
		map[string]interface{}{
			"streamName": "test",
			"region":     "us-east-1",
		},
	},
	{
		awskinesis.AWSKinesis{
			AWSAccessKeyID:     "id",
			AWSSecretAccessKey: "key",
			AWSSessionToken:    "token",
		},
		map[string]interface{}{
			"streamName":         "",
			"region":             "",
			"awsAccessKeyId":     "*****",
			"awsSecretAccessKey": "*****",
			"awsSessionToken":    "*****",
		},
	},
}
