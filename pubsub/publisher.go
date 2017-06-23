package pubsub

import "github.com/serverless/event-gateway/functions"

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
	FunctionEnd FunctionEnd
	FunctionID  functions.FunctionID
	TopicID     TopicID
}
