package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"go.uber.org/zap/zapcore"
)

// Function represents a function deployed on one of the supported providers.
type Function struct {
	Space    string    `json:"space" validate:"required,min=3,space"`
	ID       ID        `json:"functionId" validate:"required,functionid"`
	Provider *Provider `json:"provider" validate:"required"`
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (f Function) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("space", string(f.Space))
	enc.AddString("functionId", string(f.ID))
	if f.Provider != nil {
		enc.AddObject("provider", f.Provider)
	}

	return nil
}

// Functions is an array of functions.
type Functions []*Function

// ID uniquely identifies a function.
type ID string

// Provider provides provider specific info about a function.
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

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (p Provider) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", string(p.Type))
	if p.ARN != "" {
		enc.AddString("arn", p.ARN)
	}
	if p.AWSAccessKeyID != "" {
		enc.AddString("awsAccessKeyId", "*****")
	}
	if p.AWSSecretAccessKey != "" {
		enc.AddString("awsSecretAccessKey", "*****")
	}
	if p.AWSSessionToken != "" {
		enc.AddString("awsSessionToken", "*****")
	}
	if p.URL != "" {
		enc.AddString("url", p.URL)
	}
	if p.EmulatorURL != "" {
		enc.AddString("emulatorUrl", p.EmulatorURL)
	}
	if p.APIVersion != "" {
		enc.AddString("apiVersion", p.APIVersion)
	}

	return nil
}

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

// Call tries to send a payload to a target function
func (f *Function) Call(payload []byte) ([]byte, error) {
	switch f.Provider.Type {
	case AWSLambda:
		return f.callAWSLambda(payload)
	case Emulator:
		return f.callEmulator(payload)
	case HTTPEndpoint:
		return f.callHTTP(payload)
	}

	return []byte{}, errors.New("calling this kind of function is not implemented")
}

// WeightedFunction is a function along with its load-balacing proportional weight.
type WeightedFunction struct {
	FunctionID ID   `json:"functionId" validate:"required"`
	Weight     uint `json:"weight" validate:"required"`
}

// WeightedFunctions is a slice of WeightedFunction's that you can choose from based on weight
type WeightedFunctions []WeightedFunction

// Choose uses the function weights to pick a single one.
func (w WeightedFunctions) Choose() (ID, error) {
	var chosenFunction ID

	if len(w) == 1 {
		chosenFunction = w[0].FunctionID
	} else {
		weightTotal := uint(0)
		for _, wf := range w {
			weightTotal += wf.Weight
		}

		if weightTotal < 1 {
			err := errors.New("target function weights sum to 0, there is not one function to target")
			return ID(""), err
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

// nolint: gocyclo
func (f *Function) callAWSLambda(payload []byte) ([]byte, error) {
	config := aws.NewConfig().WithRegion(f.Provider.Region)
	if f.Provider.AWSAccessKeyID != "" && f.Provider.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(f.Provider.AWSAccessKeyID, f.Provider.AWSSecretAccessKey, f.Provider.AWSSessionToken))
	}

	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, &ErrFunctionProviderError{err}
	}
	awslambda := lambda.New(awsSession)

	invokeOutput, err := awslambda.Invoke(&lambda.InvokeInput{
		FunctionName: &f.Provider.ARN,
		Payload:      payload,
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			switch awserr.Code() {
			case "AccessDeniedException",
				"ExpiredTokenException",
				"UnrecognizedClientException":
				return nil, &ErrFunctionAccessDenied{awserr}
			case lambda.ErrCodeServiceException:
				return nil, &ErrFunctionProviderError{awserr}
			default:
				return nil, &ErrFunctionCallFailed{awserr}
			}
		}
	}

	if invokeOutput.FunctionError != nil {
		return nil, &ErrFunctionError{errors.New(*invokeOutput.FunctionError)}
	}

	return invokeOutput.Payload, err
}

func (f *Function) callHTTP(payload []byte) ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Post(f.Provider.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, &ErrFunctionCallFailed{err}
	}
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, &ErrFunctionError{fmt.Errorf("HTTP status code: %d", http.StatusInternalServerError)}
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (f *Function) callEmulator(payload []byte) ([]byte, error) {
	type emulatorInvokeSchema struct {
		FunctionID string      `json:"functionId"`
		Payload    interface{} `json:"payload"`
	}

	client := http.Client{
		Timeout: time.Second * 5,
	}

	var invokePayload interface{}
	err := json.Unmarshal(payload, &invokePayload)
	if err != nil {
		return nil, err
	}

	emulatorURL, err := url.Parse(f.Provider.EmulatorURL)
	if err != nil {
		return nil, fmt.Errorf("Invalid Emulator URL %q for Function %q", f.Provider.EmulatorURL, f.ID)
	}

	switch f.Provider.APIVersion {
	case "v0":
		emulatorURL.Path = path.Join(f.Provider.APIVersion, "emulator/api/functions/invoke")
	default:
		return nil, fmt.Errorf("Invalid Emulator API version %q for Function %q", f.Provider.APIVersion, f.ID)
	}

	emulatorPayload, err := json.Marshal(emulatorInvokeSchema{FunctionID: string(f.ID), Payload: invokePayload})
	if err != nil {
		return nil, err
	}

	resp, err := client.Post(emulatorURL.String(), "application/json", bytes.NewReader(emulatorPayload))
	if err != nil {
		return nil, &ErrFunctionCallFailed{err}
	}
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, &ErrFunctionError{fmt.Errorf("HTTP status code: %d", http.StatusInternalServerError)}
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
