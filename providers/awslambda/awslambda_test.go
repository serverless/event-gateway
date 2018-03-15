package awslambda_test

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/awslambda"
	"github.com/stretchr/testify/assert"
)

func TestLoad_MissingARN(t *testing.T) {
	loader := &awslambda.ProviderLoader{}

	_, err := loader.Load([]byte(`{"region": "us-east-1"}`))

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."})
}

func TestLoad_MissingRegion(t *testing.T) {
	loader := &awslambda.ProviderLoader{}

	_, err := loader.Load([]byte(`{"arn": "testarn"}`))

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."})
}
