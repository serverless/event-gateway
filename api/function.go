package api

import (
	"errors"
	"math/rand"
)

// Function represents a deployed on one of the supported providers.
type Function struct {
	ID       FunctionID `json:"functionId" validate:"required,functionid"`
	Provider *Provider  `json:"provider" validate:"required"`
}

// FunctionID uniquely identifies a function
type FunctionID string

// Provider provides provider specific info about a function
type Provider struct {
	Type ProviderType `json:"type" validate:"required,eq=awslambda|eq=http|eq=weighted|eq=emulator"`

	// AWS Lambda function
	ARN                string `json:"arn,omitempty"`
	Region             string `json:"region,omitempty"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSSessionToken    string `json:"awsSessionToken,omitempty"`

	// Group weighted function
	Weighted WeightedFunctions `json:"weighted,omitempty"`

	// HTTP function
	URL string `json:"url,omitempty" validate:"omitempty,url"`

	// Emulator function
	EmulatorURL string `json:"emulatorUrl,omitempty"`
	APIVersion  string `json:"apiVersion,omitempty"`
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
	// Emulator represents a function registered with the local emulator.
	Emulator ProviderType = "emulator"
)

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

// FunctionService represents service for managing functions.
type FunctionService interface {
	RegisterFunction(fn *Function) (*Function, error)
	UpdateFunction(fn *Function) (*Function, error)
	GetFunction(id FunctionID) (*Function, error)
	GetAllFunctions() ([]*Function, error)
	DeleteFunction(id FunctionID) error
}
