package functions

import (
	"fmt"
)

// ErrorNotFound occurs when function couldn't been found in the discovery.
type ErrorNotFound struct {
	ID FunctionID
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", string(e.ID))
}

// ErrorAlreadyRegistered occurs when function with specified name is already registered..
type ErrorAlreadyRegistered struct {
	ID FunctionID
}

func (e ErrorAlreadyRegistered) Error() string {
	return fmt.Sprintf("Function %q already registered.", string(e.ID))
}

// ErrorValidation occurs when function payload doesn't validate.
type ErrorValidation struct {
	original string
}

func (e ErrorValidation) Error() string {
	return fmt.Sprintf("Function doesn't validate. Validation error: %q", e.original)
}
