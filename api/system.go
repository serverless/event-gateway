package api

import "net/http"

// SystemEventReceivedType is a system event emitted when the Event Gateway receives an event.
const SystemEventReceivedType = EventType("gateway.event.received")

// SystemEventReceivedData struct.
type SystemEventReceivedData struct {
	Path    string      `json:"path"`
	Event   Event       `json:"event"`
	Headers http.Header `json:"header"`
}

// SystemFunctionInvokingType is a system event emitted before invoking a function.
const SystemFunctionInvokingType = EventType("gateway.function.invoking")

// SystemFunctionInvokingData struct.
type SystemFunctionInvokingData struct {
	FunctionID FunctionID `json:"functionId"`
	Event      Event      `json:"event"`
}

// SystemFunctionInvokedType is a system event emitted after successful function invocation.
const SystemFunctionInvokedType = EventType("gateway.function.invoked")

// SystemFunctionInvokedData struct.
type SystemFunctionInvokedData struct {
	FunctionID FunctionID `json:"functionId"`
	Event      Event      `json:"event"`
	Result     []byte     `json:"result"`
}

// SystemFunctionInvocationFailedType is a system event emitted after successful function invocation.
const SystemFunctionInvocationFailedType = EventType("gateway.function.invocationFailed")

// SystemFunctionInvocationFailedData struct.
type SystemFunctionInvocationFailedData struct {
	FunctionID FunctionID `json:"functionId"`
	Event      Event      `json:"event"`
	Error      []byte     `json:"result"`
}
