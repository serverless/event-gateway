package function

import "fmt"

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
