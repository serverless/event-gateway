package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/serverless/event-gateway/api"
)

// Caller calls function depending on provider.
type Caller struct{}

// Call calls passed function depending on provider type.
func (c *Caller) Call(function *api.Function, payload []byte) ([]byte, error) {
	switch function.Provider.Type {
	case api.AWSLambda:
		return callAWSLambda(function, payload)
	case api.Emulator:
		return callEmulator(function, payload)
	case api.HTTPEndpoint:
		return callHTTP(function, payload)
	}

	return []byte{}, errors.New("calling this kind of function is not implemented")
}

func callAWSLambda(f *api.Function, payload []byte) ([]byte, error) {
	config := aws.NewConfig().WithRegion(f.Provider.Region)
	if f.Provider.AWSAccessKeyID != "" && f.Provider.AWSSecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(f.Provider.AWSAccessKeyID, f.Provider.AWSSecretAccessKey, f.Provider.AWSSessionToken))
	}

	awslambda := lambda.New(session.New(config))

	invokeOutput, err := awslambda.Invoke(&lambda.InvokeInput{
		FunctionName: &f.Provider.ARN,
		Payload:      payload,
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			switch awserr.Code() {
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

func callHTTP(f *api.Function, payload []byte) ([]byte, error) {
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

func callEmulator(f *api.Function, payload []byte) ([]byte, error) {
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

// ErrFunctionCallFailed occurs when function call failed because of provider error.
type ErrFunctionCallFailed struct {
	original error
}

func (e ErrFunctionCallFailed) Error() string {
	return fmt.Sprintf("Function call failed. Error: %q", e.original)
}

// ErrFunctionProviderError occurs when function call failed because of provider error.
type ErrFunctionProviderError struct {
	original error
}

func (e ErrFunctionProviderError) Error() string {
	return fmt.Sprintf("Function call failed because of provider error. Error: %q", e.original)
}

// ErrFunctionError occurs when function call failed because of function error.
type ErrFunctionError struct {
	original error
}

func (e ErrFunctionError) Error() string {
	return fmt.Sprintf("Function call failed because of runtime error. Error: %q", e.original)
}
