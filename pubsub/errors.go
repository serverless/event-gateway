package pubsub

import "fmt"

// ErrorAlreadyExists occurs when topic with the same name already exists.
type ErrorAlreadyExists struct {
	ID TopicID
}

func (e ErrorAlreadyExists) Error() string {
	return fmt.Sprintf("Topic %q already exits.", e.ID)
}

// ErrorValidation occurs when function payload doesn't validate.
type ErrorValidation struct {
	original error
}

func (e ErrorValidation) Error() string {
	return fmt.Sprintf("Topic doesn't validate. Validation error: %q", e.original)
}

// ErrorNotFound occurs when topic cannot be found.
type ErrorNotFound struct {
	ID string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("Topic %q not found.", e.ID)
}
