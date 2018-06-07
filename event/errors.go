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

// ErrEventTypeHasSubscriptions occurs when there are subscription for the event type.
type ErrEventTypeHasSubscriptions struct{}

func (e ErrEventTypeHasSubscriptions) Error() string {
	return fmt.Sprintf("Event type cannot be deleted because there are subscriptions using it.")
}

// ErrAuthorizerDoesNotExists occurs when there authorizer function doesn't exists.
type ErrAuthorizerDoesNotExists struct{}

func (e ErrAuthorizerDoesNotExists) Error() string {
	return fmt.Sprintf("Authorizer function doesn't exists.")
}

// ErrParsingCloudEvent occurs when payload is not valid CloudEvent.
type ErrParsingCloudEvent struct {
	Message string
}

func (e ErrParsingCloudEvent) Error() string {
	return fmt.Sprintf("CloudEvent doesn't validate: %s", e.Message)
}
