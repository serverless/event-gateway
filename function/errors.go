package function

import (
	"fmt"
)

// ErrFunctionNotFound occurs when function couldn't been found in the discovery.
type ErrFunctionNotFound struct {
	ID ID
}

func (e ErrFunctionNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", string(e.ID))
}

// ErrFunctionAlreadyRegistered occurs when function with specified name is already registered.
type ErrFunctionAlreadyRegistered struct {
	ID ID
}

func (e ErrFunctionAlreadyRegistered) Error() string {
	return fmt.Sprintf("Function %q already registered.", string(e.ID))
}

// ErrFunctionValidation occurs when function payload doesn't validate.
type ErrFunctionValidation struct {
	Message string
}

func (e ErrFunctionValidation) Error() string {
	return fmt.Sprintf("Function doesn't validate. Validation error: %q", e.Message)
}

// ErrFunctionCallFailed occurs when function call failed because of provider error.
type ErrFunctionCallFailed struct {
	Original error
}

func (e ErrFunctionCallFailed) Error() string {
	return fmt.Sprintf("Function call failed. Error: %q", e.Original)
}

// ErrFunctionAccessDenied occurs when Event Gateway don't have access to call a function.
type ErrFunctionAccessDenied struct {
	Original error
}

func (e ErrFunctionAccessDenied) Error() string {
	return fmt.Sprintf("Function access denied. Error: %q", e.Original)
}

// ErrFunctionProviderError occurs when function call failed because of provider error.
type ErrFunctionProviderError struct {
	Original error
}

func (e ErrFunctionProviderError) Error() string {
	return fmt.Sprintf("Function call failed because of provider error. Error: %q", e.Original)
}

// ErrFunctionError occurs when function call failed because of function error.
type ErrFunctionError struct {
	Original error
}

func (e ErrFunctionError) Error() string {
	return fmt.Sprintf("Function call failed because of runtime error. Error: %q", e.Original)
}

// ErrFunctionHasSubscriptionsError occurs when function with subscription is being deleted.
type ErrFunctionHasSubscriptionsError struct{}

func (e ErrFunctionHasSubscriptionsError) Error() string {
	return fmt.Sprintf("Function cannot be deleted because it's subscribed to a least one event.")
}
