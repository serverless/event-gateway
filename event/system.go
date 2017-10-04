package event

import (
	"github.com/serverless/event-gateway/functions"
)

// SystemEventReceivedType is a system event emmited when the Event Gateway receives an event.
const SystemEventReceivedType = Type("gateway.event.received")

// SystemEventReceived struct.
type SystemEventReceived struct {
	Path  string `json:"path"`
	Event Event  `json:"event"`
}

// SystemFunctionInvokingType is a system event emmited before invoking a function.
const SystemFunctionInvokingType = Type("gateway.function.invoking")

// SystemFunctionInvoking struct.
type SystemFunctionInvoking struct {
	FunctionID functions.FunctionID `json:"functionId"`
	Event      Event                `json:"event"`
}
