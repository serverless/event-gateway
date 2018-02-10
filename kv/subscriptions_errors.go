package kv

import (
	"fmt"

	"github.com/serverless/event-gateway/subscription"
)

// ErrSubscriptionAlreadyExists occurs when subscription with the same ID already exists.
type ErrSubscriptionAlreadyExists struct {
	ID subscription.ID
}

func (e ErrSubscriptionAlreadyExists) Error() string {
	return fmt.Sprintf("Subscription %q already exists.", e.ID)
}

// ErrSubscriptionValidation occurs when subscription payload doesn't validate.
type ErrSubscriptionValidation struct {
	original string
}

func (e ErrSubscriptionValidation) Error() string {
	return fmt.Sprintf("Subscription doesn't validate. Validation error: %q", e.original)
}

// ErrSubscriptionNotFound occurs when subscription cannot be found.
type ErrSubscriptionNotFound struct {
	ID subscription.ID
}

func (e ErrSubscriptionNotFound) Error() string {
	return fmt.Sprintf("Subscription %q not found.", e.ID)
}

// ErrFunctionNotFound occurs when subscription cannot be created because backing function doesn't exist.
type ErrFunctionNotFound struct {
	functionID string
}

func (e ErrFunctionNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", e.functionID)
}

// ErrPathConfict occurs when HTTP subscription path conflicts with existing path.
type ErrPathConfict struct {
	original string
}

func (e ErrPathConfict) Error() string {
	return fmt.Sprintf("Subscription path conflict: %s.", e.original)
}
