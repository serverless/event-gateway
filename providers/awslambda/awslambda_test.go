package awslambda_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/awslambda"
	"github.com/serverless/event-gateway/providers/awslambda/mock"
	"github.com/stretchr/testify/assert"
)

func TestLoad_MalformedJSON(t *testing.T) {
	config := []byte(`{"arn": "", "region": `)
	loader := awslambda.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Unable to load function provider config: unexpected end of JSON input"})
}

func TestLoad_MissingARN(t *testing.T) {
	config := []byte(`{"arn": "", "region": "us-east-1"}`)
	loader := awslambda.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."})
}

func TestLoad_MissingRegion(t *testing.T) {
	config := []byte(`{"arn": "test", "region": ""}`)
	loader := awslambda.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."})
}

func TestCall(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	serviceMock := mock.NewMockLambdaAPI(mockCtrl)
	opts := &lambda.InvokeInput{
		FunctionName: aws.String("testarn"),
		Payload:      []byte("testpayload"),
	}
	serviceMock.EXPECT().Invoke(opts).Return(&lambda.InvokeOutput{Payload: []byte("testoutput")}, nil)
	provider := awslambda.AWSLambda{
		Service: serviceMock,

		ARN:    "testarn",
		Region: "us-east-1",
	}

	output, err := provider.Call([]byte("testpayload"))

	assert.Nil(t, err)
	assert.Equal(t, []byte("testoutput"), output)
}
