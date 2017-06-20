package functions

// FunctionID uniquely identifies a function
type FunctionID string

// FunctionType represents what kind of function this is.
type FunctionType uint

// Function repesents a deployed on one of the supported providers.
type Function struct {
	ID   FunctionID   `json:"functionId"`
	Type FunctionType `json:"-"`

	// Only one of the following properties can be defined.
	AWSLambda       *AWSLambdaProperties       `json:"awsLambda,omitempty"`
	GCloudFunction  *GCloudFunctionProperties  `json:"gcloudFunction,omitempty"`
	AzureFunction   *AzureFunctionProperties   `json:"azureFunction,omitempty"`
	OpenWhiskAction *OpenWhiskActionProperties `json:"openWhiskAction,omitempty"`
	Group           *GroupProperties           `json:"group,omitempty"`
	HTTP            *HTTPProperties            `json:"http,omitempty"`
}

const (
	_ FunctionType = iota
	// AWSLambda means this function is an AWS Lambda function.
	AWSLambda
	// GCloudFunction means this function is a Google Cloud Function.
	GCloudFunction
	// AzureFunction means this function is a Microsoft Azure Function.
	AzureFunction
	// OpenWhiskAction means this function is an OpenWhisk Action.
	OpenWhiskAction
	// Group means this is a load balancing group of functions with associated weights.
	Group
	// HTTP means this is an http endpoint.
	HTTP
)

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
	Functions []WeightedFunction `json:"functions" validate:"required"`
}

// WeightedFunction is a function along with its load-balacing proportional weight.
type WeightedFunction struct {
	FunctionID FunctionID `json:"functionId" validate:"required"`
	Weight     uint       `json:"weight" validate:"required"`
}

// HTTPProperties contains the configuration required to call an http endpoint.
type HTTPProperties struct {
	URL string `json:"url" validate:"required,url"`
}
