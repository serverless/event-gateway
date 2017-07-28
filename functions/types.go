package functions

import (
	"bytes"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// Caller sends a payload to a target function
type Caller interface {
	Call([]byte) ([]byte, error)
}

// FunctionID uniquely identifies a function
type FunctionID string

// Function repesents a deployed on one of the supported providers.
type Function struct {
	ID       FunctionID `json:"functionId" validate:"required"`
	Provider *Provider  `json:"provider" validate:"required"`
}

// ProviderType represents what kind of function provider this is.
type ProviderType string

const (
	// AWSLambda represents AWS Lambda function.
	AWSLambda ProviderType = "awslambda"
	// Weighted contains a set of other functions and their load balancing weights.
	Weighted ProviderType = "weighted"
	// HTTPEndpoint represents function defined as HTTP endpoint.
	HTTPEndpoint ProviderType = "http"
)

// Provider provides provider specific info about a function
type Provider struct {
	Type ProviderType `json:"type" validate:"required,eq=awslambda|eq=http|eq=weighted"`

	// AWS Lambda function
	ARN                string `json:"arn,omitempty"`
	Region             string `json:"region,omitempty"`
	AWSAccessKeyID     string `json:"awsAccessKeyID,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`

	// Group weighted function
	Weighted WeightedFunctions `json:"weighted,omitempty"`

	// HTTP function
	URL string `json:"url,omitempty" validate:"omitempty,url"`
}

// Call tries to send a payload to a target function
func (f *Function) Call(payload []byte) ([]byte, error) {
	switch f.Provider.Type {
	case AWSLambda:
		return f.callAWSLambda(payload)
	case HTTPEndpoint:
		return f.callHTTP(payload)
	}

	return []byte{}, errors.New("calling this kind of function is not implemented")
}

// WeightedFunction is a function along with its load-balacing proportional weight.
type WeightedFunction struct {
	FunctionID FunctionID `json:"functionId" validate:"required"`
	Weight     uint       `json:"weight" validate:"required"`
}

// WeightedFunctions is a slice of WeightedFunction's that you can choose from based on weight
type WeightedFunctions []WeightedFunction

// Choose uses the function weights to pick a single one.
func (w WeightedFunctions) Choose() (FunctionID, error) {
	var chosenFunction FunctionID

	if len(w) == 1 {
		chosenFunction = w[0].FunctionID
	} else {
		weightTotal := uint(0)
		for _, wf := range w {
			weightTotal += wf.Weight
		}

		if weightTotal < 1 {
			err := errors.New("target function weights sum to 0, there is not one function to target")
			return FunctionID(""), err
		}

		chosenWeight := uint(1 + rand.Intn(int(weightTotal)))
		weightsSoFar := uint(0)
		for _, wf := range w {
			chosenFunction = wf.FunctionID
			weightsSoFar += wf.Weight
			if weightsSoFar >= chosenWeight {
				break
			}
		}
	}

	return chosenFunction, nil
}

func (f *Function) callAWSLambda(payload []byte) ([]byte, error) {
	config := aws.NewConfig().WithRegion(f.Provider.Region)
	if f.Provider.AWSAccessKeyID != "" && f.Provider.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(f.Provider.AWSAccessKeyID, f.Provider.AWSSecretAccessKey, ""))
	}

	awslambda := lambda.New(session.New(config))

	invokeOutput, err := awslambda.Invoke(&lambda.InvokeInput{
		FunctionName: &f.Provider.ARN,
		Payload:      payload,
	})

	return invokeOutput.Payload, err
}

func (f *Function) callHTTP(payload []byte) ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Post(f.Provider.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
