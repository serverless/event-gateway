package functions

// FunctionID uniquely identifies a function
type FunctionID string

// FunctionSpec is the JSON representation of a Function.
type FunctionSpec struct {
	ID         FunctionID        `json:"functionId"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
}

// FunctionType represents what kind of function this is.
type FunctionType uint

const (
	// AwsLambda means this function is an AWS Lambda function.
	AwsLambda FunctionType = iota
	// GcloudFunction means this function is a Google Cloud Function.
	GcloudFunction
	// AzureFunction means this function is a Microsoft Azure Function.
	AzureFunction
	// OpenWhiskAction means this function is an OpenWhisk Action.
	OpenWhiskAction
	// Group means this is a load balancing group of functions with associated weights.
	Group
	// HTTP means this is an http endpoint.
	HTTP
)

// Function repesents a deployed on one of the supported providers.
type Function struct {
	ID              FunctionID
	Type            FunctionType
	AwsLambda       *AwsLambdaProperties
	GcloudFunction  *GcloudFunctionProperties
	AzureFunction   *AzureFunctionProperties
	OpenWhiskAction *OpenWhiskActionProperties
	Group           *GroupProperties
	HTTP            *HTTPProperties
}

// AwsLambdaProperties contains the configuration required to call an AWS Lambda function
type AwsLambdaProperties struct{}

// GcloudFunctionProperties contains the configuration required to call a Google Cloud Function
type GcloudFunctionProperties struct{}

// AzureFunctionProperties contains the configuration required to call an Azure Function
type AzureFunctionProperties struct{}

// OpenWhiskActionProperties contains the configuration required to call an OpenWhisk action
type OpenWhiskActionProperties struct{}

// GroupProperties contains a set of other functions and their load balancing weights
type GroupProperties struct {
	Functions []WeightedFunction
}

// HTTPProperties contains the configuration required to call an http endpoint
type HTTPProperties struct{}

// WeightedFunction is a function along with its load-balacing proportional weight.
type WeightedFunction struct {
	GroupFunctionID FunctionID
	FunctionID      FunctionID
	Weight          uint
}

// Credentials that allows calling user's function.
type Credentials struct {
	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
}
