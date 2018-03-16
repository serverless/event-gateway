package awslambda_test

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/awslambda"
	"github.com/stretchr/testify/assert"
)

func TestValidation_MissingARN(t *testing.T) {
	provider := &awslambda.AWSLambda{
		Region: "us-east-1",
	}

	err := provider.Validate()

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."})
}

func TestValidation_MissingRegion(t *testing.T) {
	provider := &awslambda.AWSLambda{
		ARN: "arn",
	}

	err := provider.Validate()

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."})
}
