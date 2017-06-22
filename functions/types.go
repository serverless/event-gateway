package functions

import (
	"errors"
	"math/rand"
)

// Caller tries to send a payload to a target function
type Caller interface {
	Call([]byte) ([]byte, error)
}

// FunctionID uniquely identifies a function
type FunctionID string

// FunctionType represents what kind of function this is.
type FunctionType uint

// Function repesents a deployed on one of the supported providers.
type Function struct {
	ID FunctionID `json:"functionId"`

	// Only one of the following properties can be defined.
	AWSLambda       *AWSLambdaProperties       `json:"awsLambda,omitempty"`
	GCloudFunction  *GCloudFunctionProperties  `json:"gcloudFunction,omitempty"`
	AzureFunction   *AzureFunctionProperties   `json:"azureFunction,omitempty"`
	OpenWhiskAction *OpenWhiskActionProperties `json:"openWhiskAction,omitempty"`
	Group           *GroupProperties           `json:"group,omitempty"`
	HTTP            *HTTPProperties            `json:"http,omitempty"`
}

// Call tries to send a payload to a target function
func (f *Function) Call(payload []byte) ([]byte, error) {
	if f.AWSLambda != nil {
		return f.AWSLambda.Call(payload)
	} else if f.HTTP != nil {
		return f.HTTP.Call(payload)
	}
	return []byte{}, errors.New("calling this kind of function is not implemented")
}

// AWSLambdaProperties contains the configuration required to call an AWS Lambda function.
type AWSLambdaProperties struct {
	ARN             string `json:"arn" validate:"required"`
	Region          string `json:"region" validate:"required"`
	Version         string `json:"version" validate:"required"`
	AccessKeyID     string `json:"accessKeyID" validate:"required"`
	SecretAccessKey string `json:"secretAccessKey" validate:"required"`
}

// GCloudFunctionProperties contains the configuration required to call a Google Cloud Function.
type GCloudFunctionProperties struct {
	Name              string `json:"name" validate:"required"`
	Region            string `json:"region" validate:"required"`
	ServiceAccountKey string `json:"serviceAccountKey" validate:"required"`
}

// AzureFunctionProperties contains the configuration required to call an Azure Function.
type AzureFunctionProperties struct {
	Name              string `json:"name" validate:"required"`
	AppName           string `json:"appName" validate:"required"`
	FunctionsAdminKey string `json:"functionsAdminKey" validate:"required"`
}

// OpenWhiskActionProperties contains the configuration required to call an OpenWhisk action.
type OpenWhiskActionProperties struct {
	Name             string `json:"name" validate:"required"`
	Namespace        string `json:"namespace" validate:"required"`
	APIHost          string `json:"apiHost" validate:"required"`
	Auth             string `json:"auth" validate:"required"`
	APIGWAccessToken string `json:"apiGwAccessToken" validate:"required"`
}

// GroupProperties contains a set of other functions and their load balancing weights.
type GroupProperties struct {
	Functions WeightedFunctions `json:"functions" validate:"required"`
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

// HTTPProperties contains the configuration required to call an http endpoint.
type HTTPProperties struct {
	URL string `json:"url" validate:"required,url"`
}
