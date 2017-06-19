package types

import (
	functionTypes "github.com/serverless/gateway/functions/types"
)

// TopicID uniquely identifies a pubsub topic
type TopicID string

// PublisherID uniquely identifies a publisher function
type PublisherID string

// Subscriber maps from TopicID to FunctionID
type Subscriber struct {
	TopicID    TopicID
	FunctionID functionTypes.FunctionID
}

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
	FunctionID  functionTypes.FunctionID
	TopicID     TopicID
}
