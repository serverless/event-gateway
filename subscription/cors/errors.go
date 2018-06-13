package cors

import (
	"fmt"
)

// ErrCORSNotFound occurs when CORS config cannot be found.
type ErrCORSNotFound struct {
	ID ID
}

func (e ErrCORSNotFound) Error() string {
	return fmt.Sprintf("CORS configuration %q not found.", e.ID)
}

// ErrCORSAlreadyExists occurs when CORS config with the same ID already exists.
type ErrCORSAlreadyExists struct {
	ID ID
}

func (e ErrCORSAlreadyExists) Error() string {
	return fmt.Sprintf("CORS configuration %q already exists.", e.ID)
}

// ErrCORSValidation occurs when CORS configuration payload doesn't validate.
type ErrCORSValidation struct {
	Message string
}

func (e ErrCORSValidation) Error() string {
	return fmt.Sprintf("CORS configuration doesn't validate. Validation error: %s", e.Message)
}

// ErrInvalidCORSUpdate occurs when a client tries to update a CORS config in a way that changes the CORS ID.
type ErrInvalidCORSUpdate struct {
	Field string
}

func (e ErrInvalidCORSUpdate) Error() string {
	return fmt.Sprintf("Invalid update. '%v' of existing CORS configuration cannot be updated.", e.Field)
}
