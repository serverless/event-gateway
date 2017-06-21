package endpoints

import "fmt"

// ErrorNotFound occurs when endpoint cannot be found.
type ErrorNotFound struct {
	ID string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("Endpoint %q not found.", e.ID)
}

// ErrorFunctionNotFound occurs when endpoint cannot be created because backing function doesn't exist.
type ErrorFunctionNotFound struct {
	functionID string
}

func (e ErrorFunctionNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", e.functionID)
}

// ErrorAlreadyExists occurs when endpoint cannot be created because mapping for HTTP method and path is already created.
type ErrorAlreadyExists struct {
	Method string
	Path   string
}

func (e ErrorAlreadyExists) Error() string {
	return fmt.Sprintf("Endpoint with method %q and path %q already exits.", e.Method, e.Path)
}

// ErrorValidation occurs when function payload doesn't validate.
type ErrorValidation struct {
	original error
}

func (e ErrorValidation) Error() string {
	return fmt.Sprintf("Function doesn't validate. Validation error: %q", e.original)
}
