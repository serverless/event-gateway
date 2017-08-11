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
	original string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("Function doesn't validate. Validation error: %q", e.original)
}
