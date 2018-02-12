package event

import (
	"net/http"

	"github.com/serverless/event-gateway/function"
)

// SystemEventReceivedType is a system event emitted when the Event Gateway receives an event.
const SystemEventReceivedType = Type("gateway.event.received")

// SystemEventReceivedData struct.
type SystemEventReceivedData struct {
	Path    string      `json:"path"`
	Event   Event       `json:"event"`
	Headers http.Header `json:"header"`
}

// SystemFunctionInvokingType is a system event emitted before invoking a function.
const SystemFunctionInvokingType = Type("gateway.function.invoking")

// SystemFunctionInvokingData struct.
type SystemFunctionInvokingData struct {
	FunctionID function.ID `json:"functionId"`
	Event      Event       `json:"event"`
}

// SystemFunctionInvokedType is a system event emitted after successful function invocation.
const SystemFunctionInvokedType = Type("gateway.function.invoked")

// SystemFunctionInvokedData struct.
type SystemFunctionInvokedData struct {
	FunctionID function.ID `json:"functionId"`
	Event      Event       `json:"event"`
	Result     []byte      `json:"result"`
}

// SystemFunctionInvocationFailedType is a system event emitted after successful function invocation.
const SystemFunctionInvocationFailedType = Type("gateway.function.invocationFailed")

// SystemFunctionInvocationFailedData struct.
type SystemFunctionInvocationFailedData struct {
	FunctionID function.ID `json:"functionId"`
	Event      Event       `json:"event"`
	Error      []byte      `json:"result"`
}
