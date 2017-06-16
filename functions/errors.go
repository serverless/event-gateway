package functions

import (
	"fmt"

	"github.com/serverless/gateway/types"
)

// ErrorNotFound occurs when function couldn't been found in the discovery.
type ErrorNotFound struct {
	name string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("Function %q not found.", e.name)
}

// ErrorInvocationFailed occurs when function invocation failed.
type ErrorInvocationFailed struct {
	err      error
	function types.Function
	instance types.Instance
}

func (e ErrorInvocationFailed) Error() string {
	return fmt.Sprintf("Calling function %q (%s, %s): %s.", e.function.ID, e.instance.Provider, e.instance.OriginID, e.err.Error())
}
