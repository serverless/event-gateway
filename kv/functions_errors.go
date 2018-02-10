package kv

import (
	"fmt"

	"github.com/serverless/event-gateway/function"
)

// ErrNotFound occurs when function couldn't been found in the discovery.
type ErrNotFound struct {
	ID function.ID
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", string(e.ID))
}

// ErrAlreadyRegistered occurs when function with specified name is already registered..
type ErrAlreadyRegistered struct {
	ID function.ID
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
