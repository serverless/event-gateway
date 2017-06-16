package types

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
