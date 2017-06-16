package types

// FunctionID uniquely identifies a function
type FunctionID string

// Function registered in the function discovery. Function repesents FaaS function deployed on one of the supported providers.
type Function struct {
	ID        FunctionID `json:"id"`
	Instances []Instance `json:"instances"`
}

// WeightedFunction is a function along with its load-balacing proportional weight.
type WeightedFunction struct {
	Function FunctionID
	Weight   uint
}

// Instance of function. A function can have multiple instances. Each instance of a function is deployed in different regions.
type Instance struct {
	Provider    string      `json:"provider"`
	OriginID    string      `json:"originId"`
	Region      string      `json:"region"`
	Credentials Credentials `json:"credentials"`
}

// Credentials that allows calling user's function.
type Credentials struct {
	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
}

// Endpoint represents single endpoint
type Endpoint struct {
	ID        string           `json:"id"`
	Functions []FunctionTarget `json:"functions"`
}

// FunctionTarget is a function exposed by Endpoints
type FunctionTarget struct {
	FunctionID string `json:"functionId"`
	Method     string `json:"method"`
	Path       string `json:"path"`
}

// TopicID uniquely identifies a pubsub topic
type TopicID string

// Subscriber maps from TopicID to FunctionID
type Subscriber struct {
	TopicID    TopicID
	FunctionID FunctionID
}

// FunctionEnd is used to specify whether the input or output
// from a function is to be used.
type FunctionEnd uint

const (
	Input FunctionEnd = iota
	Output
)

// Publisher maps from {input,output} + FunctionID to TopicID
type Publisher struct {
	FunctionEnd FunctionEnd
	FunctionID  FunctionID
	TopicID     TopicID
}
