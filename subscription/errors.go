package subscription

import (
	"fmt"
)

// ErrSubscriptionNotFound occurs when subscription cannot be found.
type ErrSubscriptionNotFound struct {
	ID ID
}

func (e ErrSubscriptionNotFound) Error() string {
	return fmt.Sprintf("Subscription %q not found.", e.ID)
}

// ErrSubscriptionAlreadyExists occurs when subscription with the same ID already exists.
type ErrSubscriptionAlreadyExists struct {
	ID ID
}

func (e ErrSubscriptionAlreadyExists) Error() string {
	return fmt.Sprintf("Subscription %q already exists.", e.ID)
}

// ErrInvalidSubscriptionUpdate occurs when a client tries to update a subscription in a way that changes the subscription ID.
type ErrInvalidSubscriptionUpdate struct {
	Field string
}

func (e ErrInvalidSubscriptionUpdate) Error() string {
	return fmt.Sprintf("Invalid update. '%v' of existing subscription cannot be updated.", e.Field)
}

// ErrSubscriptionValidation occurs when subscription payload doesn't validate.
type ErrSubscriptionValidation struct {
	Message string
}

func (e ErrSubscriptionValidation) Error() string {
	return fmt.Sprintf("Subscription doesn't validate. Validation error: %s", e.Message)
}

// ErrPathConfict occurs when HTTP subscription path conflicts with existing path.
type ErrPathConfict struct {
	Message string
}

func (e ErrPathConfict) Error() string {
	return fmt.Sprintf("Subscription path conflict: %s", e.Message)
}
