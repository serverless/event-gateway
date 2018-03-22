package awskinesis_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/providers/awskinesis"
	"github.com/serverless/event-gateway/providers/awskinesis/mock"
	"github.com/stretchr/testify/assert"
)

func TestLoad_MalformedJSON(t *testing.T) {
	config := []byte(`{"streamName": "", "region": `)
	loader := awskinesis.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.EqualError(t, err, "unable to load function provider config: unexpected end of JSON input")
}

func TestLoad_MissingStreamName(t *testing.T) {
	config := []byte(`{"streamName": "", "region": "us-east-1"}`)
	loader := awskinesis.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.EqualError(t, err, "missing required fields for AWS Kinesis function")
}

func TestLoad_MissingRegion(t *testing.T) {
	config := []byte(`{"streamName": "test", "region": ""}`)
	loader := awskinesis.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.EqualError(t, err, "missing required fields for AWS Kinesis function")
}

func TestCall(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	serviceMock := mock.NewMockKinesisAPI(mockCtrl)
	serviceMock.EXPECT().PutRecord(gomock.Any()).Return(&kinesis.PutRecordOutput{SequenceNumber: aws.String("testseq")}, nil)
	provider := awskinesis.AWSKinesis{
		Service:    serviceMock,
		StreamName: "teststream",
		Region:     "us-east-1",
	}

	output, err := provider.Call([]byte("testpayload"))

	assert.Nil(t, err)
	assert.Equal(t, []byte("testseq"), output)
}
