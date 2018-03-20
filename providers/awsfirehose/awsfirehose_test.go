package awsfirehose_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/awsfirehose"
	"github.com/serverless/event-gateway/providers/awsfirehose/mock"
	"github.com/stretchr/testify/assert"
)

func TestLoad_MalformedJSON(t *testing.T) {
	config := []byte(`{"deliveryStreamName": "", "region": `)
	loader := awsfirehose.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Unable to load function provider config: unexpected end of JSON input"})
}

func TestLoad_MissingStreamName(t *testing.T) {
	config := []byte(`{"deliveryStreamName": "", "region": "us-east-1"}`)
	loader := awsfirehose.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Firehose function."})
}

func TestLoad_MissingRegion(t *testing.T) {
	config := []byte(`{"deliveryStreamName": "test", "region": ""}`)
	loader := awsfirehose.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Firehose function."})
}

func TestCall(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	serviceMock := mock.NewMockFirehoseAPI(mockCtrl)
	serviceMock.EXPECT().PutRecord(gomock.Any()).Return(&firehose.PutRecordOutput{RecordId: aws.String("testrecord")}, nil)
	provider := awsfirehose.AWSFirehose{
		Service:    serviceMock,
		DeliveryStreamName: "teststream",
		Region:     "us-east-1",
	}

	output, err := provider.Call([]byte("testpayload"))

	assert.Nil(t, err)
	assert.Equal(t, []byte("testrecord"), output)
}
