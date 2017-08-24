package router

import (
	"fmt"
	"net/http"
)

// ErrHTTPResponseObjectMalformed occurs when HTTP response object is not valid JSON.
type ErrHTTPResponseObjectMalformed struct {
	StatusCode int
}

func (e ErrHTTPResponseObjectMalformed) Error() string {
	return fmt.Sprintf("HTTP response object returned by function malformed.")
}

// NewErrHTTPResponseObjectMalformed return ErrHTTPResponseObjectMalformed
func NewErrHTTPResponseObjectMalformed() ErrHTTPResponseObjectMalformed {
	return ErrHTTPResponseObjectMalformed{http.StatusInternalServerError}
}
