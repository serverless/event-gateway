package functions

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

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
