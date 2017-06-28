package pubsub

import (
	"fmt"

	"github.com/serverless/event-gateway/functions"
)

// PublisherID uniquely identifies a publisher function
type PublisherID string

// FunctionEnd is used to specify whether the input or output
// from a function is to be used.
type FunctionEnd uint

const (
	// Input means the input to a function should feed a topic
	Input FunctionEnd = iota
	// Output means the output of a function should feed a topic
	Output
)

// Publisher maps from {input,output} + FunctionID to TopicID
type Publisher struct {
	ID         PublisherID          `json:"id" validate:"required"`
	Type       string               `json:"type" validate:"required,eq=input|eq=output"`
	FunctionID functions.FunctionID `json:"functionId" validate:"required"`
	TopicID    TopicID              `json:"topicId" validate:"required"`
}

func publisherID(tid TopicID, end string, fid functions.FunctionID) PublisherID {
	return PublisherID(string(tid) + "-" + end + "-" + string(fid))
}

// ErrorPublisherValidation occurs when publisher payload doesn't validate.
type ErrorPublisherValidation struct {
	original error
}

func (e ErrorPublisherValidation) Error() string {
	return fmt.Sprintf("Publisher doesn't validate. Validation error: %q", e.original)
}

// ErrorPublisherAlreadyExists occurs when the publisher has already been created.
type ErrorPublisherAlreadyExists struct {
	ID PublisherID
}

func (e ErrorPublisherAlreadyExists) Error() string {
	return fmt.Sprintf("Publisher %q already exits.", e.ID)
}

// ErrorPublisherNotFound occurs when publisher cannot be found.
type ErrorPublisherNotFound struct {
	ID PublisherID
}

func (e ErrorPublisherNotFound) Error() string {
	return fmt.Sprintf("Publisher %q not found.", e.ID)
}
