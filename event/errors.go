package event

import "fmt"

// ErrEventTypeNotFound occurs when event type cannot be found.
type ErrEventTypeNotFound struct {
	Name TypeName
}

func (e ErrEventTypeNotFound) Error() string {
	return fmt.Sprintf("Event Type %q not found.", e.Name)
}

// ErrEventTypeAlreadyExists occurs when event type with specified name already exists.
type ErrEventTypeAlreadyExists struct {
	Name TypeName
}

func (e ErrEventTypeAlreadyExists) Error() string {
	return fmt.Sprintf("Event Type %q already exists.", e.Name)
}

// ErrEventTypeValidation occurs when event type payload doesn't validate.
type ErrEventTypeValidation struct {
	Message string
}

func (e ErrEventTypeValidation) Error() string {
	return fmt.Sprintf("Event Type doesn't validate. Validation error: %s", e.Message)
}

// ErrEventTypeHasSubscriptionsError occurs when there are subscription for the event type.
type ErrEventTypeHasSubscriptionsError struct{}

func (e ErrEventTypeHasSubscriptionsError) Error() string {
	return fmt.Sprintf("Event type cannot be deleted because there are subscriptions using it.")
}

// ErrParsingCloudEvent occurs when payload is not valid CloudEvent.
type ErrParsingCloudEvent struct {
	Message string
}

func (e ErrParsingCloudEvent) Error() string {
	return fmt.Sprintf("CloudEvent doesn't validate: %s", e.Message)
}
