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

// ErrorMoreThanOneFunctionTypeSpecified occurs when function payload include properties for more than one type of function.
type ErrorMoreThanOneFunctionTypeSpecified struct{}

func (e ErrorMoreThanOneFunctionTypeSpecified) Error() string {
	return "More that one function type specified."
}

// ErrorValidation occurs when function payload doesn't validate.
type ErrorValidation struct {
	original error
}

func (e ErrorValidation) Error() string {
	return fmt.Sprintf("Function doesn't validate. Validation error: %q", e.original)
}
