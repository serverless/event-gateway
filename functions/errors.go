package functions

import (
	"fmt"
)

// ErrorNotFound occurs when function couldn't been found in the discovery.
type ErrorNotFound struct {
	name string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", e.name)
}

// ErrorPropertiesNotSpecified occurs when function payload doesn't include function properties.
type ErrorPropertiesNotSpecified struct{}

func (e ErrorPropertiesNotSpecified) Error() string {
	return "Function properties not specified."
}

// ErrorValidation occurs when function payload doesn't validate.
type ErrorValidation struct {
	original error
}

func (e ErrorValidation) Error() string {
	return fmt.Sprintf("Function doesn't validate. Validation error: %q", e.original)
}
