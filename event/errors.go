package event

import "fmt"

// ErrParsingCloudEvent occurs when payload is not valid CloudEvent.
type ErrParsingCloudEvent struct {
	Message string
}

func (e ErrParsingCloudEvent) Error() string {
	return fmt.Sprintf("CloudEvent doesn't validate: %s", e.Message)
}
