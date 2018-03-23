package awslambda_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/awslambda"
	"github.com/serverless/event-gateway/providers/awslambda/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLoad(t *testing.T) {
	for _, testCase := range loadTests {
		config := []byte(testCase.config)
		loader := awslambda.ProviderLoader{}

		_, err := loader.Load(config)

		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestCall(t *testing.T) {
	for _, testCase := range callTests {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		serviceMock := mock.NewMockLambdaAPI(mockCtrl)
		opts := &lambda.InvokeInput{
			FunctionName: aws.String("testarn"),
			Payload:      []byte("testpayload"),
		}
		serviceMock.EXPECT().Invoke(opts).Return(testCase.invokeResult, testCase.invokeError)

		provider := awslambda.AWSLambda{
			Service: serviceMock,

			ARN:    "testarn",
			Region: "us-east-1",
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
		`{"arn": "", "region": `,
		errors.New("unable to load function provider config: unexpected end of JSON input"),
	},
	{
		`{"arn": "", "region": "us-east-1"}`,
		errors.New("missing required fields for AWS Lambda function"),
	},
	{
		`{"arn": "test", "region": ""}`,
		errors.New("missing required fields for AWS Lambda function"),
	},
}

var callTests = []struct {
	invokeResult   *lambda.InvokeOutput
	invokeError    error
	expectedResult []byte
	expectedError  error
}{
	{
		&lambda.InvokeOutput{Payload: []byte("testres")},
		nil,
		[]byte("testres"),
		nil,
	},
	{
		&lambda.InvokeOutput{FunctionError: aws.String("FuncErr")},
		nil,
		[]byte(nil),
		&function.ErrFunctionError{Original: errors.New("FuncErr")},
	},
	{
		nil,
		awserr.New("TestCode", "", nil),
		[]byte(nil),
		&function.ErrFunctionCallFailed{Original: awserr.New("TestCode", "", nil)},
	},
	{
		nil,
		awserr.New("AccessDeniedException", "", nil),
		[]byte(nil),
		&function.ErrFunctionAccessDenied{Original: awserr.New("AccessDeniedException", "", nil)},
	},
	{
		nil,
		awserr.New(lambda.ErrCodeServiceException, "", nil),
		[]byte(nil),
		&function.ErrFunctionProviderError{Original: awserr.New(lambda.ErrCodeServiceException, "", nil)},
	},
}

var logTests = []struct {
	provider       function.Provider
	expectedFields map[string]interface{}
}{
	{
		awslambda.AWSLambda{
			ARN:    "test",
			Region: "us-east-1",
		},
		map[string]interface{}{
			"arn":    "test",
			"region": "us-east-1",
		},
	},
	{
		awslambda.AWSLambda{
			AWSAccessKeyID:     "id",
			AWSSecretAccessKey: "key",
			AWSSessionToken:    "token",
		},
		map[string]interface{}{
			"arn":                "",
			"region":             "",
			"awsAccessKeyId":     "*****",
			"awsSecretAccessKey": "*****",
			"awsSessionToken":    "*****",
		},
	},
}
