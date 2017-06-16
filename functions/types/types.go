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
