package function

import "github.com/serverless/event-gateway/metadata"

// Service represents service for managing functions.
type Service interface {
	GetFunction(space string, id ID) (*Function, error)
	ListFunctions(space string, filters ...metadata.Filter) (Functions, error)
	CreateFunction(fn *Function) (*Function, error)
	UpdateFunction(fn *Function) (*Function, error)
	DeleteFunction(space string, id ID) error
}
