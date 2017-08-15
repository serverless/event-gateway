package functions

import (
	"fmt"
)

// ErrNotFound occurs when function couldn't been found in the discovery.
type ErrNotFound struct {
	ID FunctionID
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", string(e.ID))
}

// ErrAlreadyRegistered occurs when function with specified name is already registered..
type ErrAlreadyRegistered struct {
	ID FunctionID
}

func (e ErrAlreadyRegistered) Error() string {
	return fmt.Sprintf("Function %q already registered.", string(e.ID))
}

// ErrValidation occurs when function payload doesn't validate.
type ErrValidation struct {
	message string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("Function doesn't validate. Validation error: %q", e.message)
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
