package functions

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// AWSLambdaProperties contains the configuration required to call an AWS Lambda function.
type AWSLambdaProperties struct {
	ARN             string `json:"arn" validate:"required"`
	Region          string `json:"region" validate:"required"`
	Version         string `json:"version" validate:"required"`
	AccessKeyID     string `json:"accessKeyID" validate:"required"`
	SecretAccessKey string `json:"secretAccessKey" validate:"required"`
}

// Call tries to send a payload to a target function
func (p *AWSLambdaProperties) Call(payload []byte) ([]byte, error) {
	creds := credentials.NewStaticCredentials(p.AccessKeyID, p.SecretAccessKey, "")

	awslambda := lambda.New(session.New(aws.NewConfig().WithRegion(p.Region).WithCredentials(creds)))

	invokeOutput, err := awslambda.Invoke(&lambda.InvokeInput{
		FunctionName: &p.ARN,
		Payload:      payload,
	})
	return invokeOutput.Payload, err
}
