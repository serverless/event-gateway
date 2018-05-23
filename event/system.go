package event

import (
	"github.com/serverless/event-gateway/function"
)

// SystemEventReceivedType is a system event emitted when the Event Gateway receives an event.
const SystemEventReceivedType = TypeName("gateway.event.received")

// SystemEventReceivedData struct.
type SystemEventReceivedData struct {
	Path    string            `json:"path"`
	Event   Event             `json:"event"`
	Headers map[string]string `json:"headers"`
}

// SystemFunctionInvokingType is a system event emitted before invoking a function.
const SystemFunctionInvokingType = TypeName("gateway.function.invoking")

// SystemFunctionInvokingData struct.
type SystemFunctionInvokingData struct {
	Space      string      `json:"space"`
	FunctionID function.ID `json:"functionId"`
	Event      Event       `json:"event"`
}

// SystemFunctionInvokedType is a system event emitted after successful function invocation.
const SystemFunctionInvokedType = TypeName("gateway.function.invoked")

// SystemFunctionInvokedData struct.
type SystemFunctionInvokedData struct {
	Space      string      `json:"space"`
	FunctionID function.ID `json:"functionId"`
	Event      Event       `json:"event"`
	Result     []byte      `json:"result"`
}

// SystemFunctionInvocationFailedType is a system event emitted after successful function invocation.
const SystemFunctionInvocationFailedType = TypeName("gateway.function.invocationFailed")

// SystemFunctionInvocationFailedData struct.
type SystemFunctionInvocationFailedData struct {
	Space      string      `json:"space"`
	FunctionID function.ID `json:"functionId"`
	Event      Event       `json:"event"`
	Error      error       `json:"result"`
}
