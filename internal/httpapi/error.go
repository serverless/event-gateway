package httpapi

import "fmt"

// Error represents generic HTTP error returned by Configuration API.
type Error struct {
	Error string `json:"error"`
}

// ErrMalformedJSON occuring when it's impossible to decode JSON payload.
type ErrMalformedJSON Error

// NewErrMalformedJSON creates ErrMalformedJSON.
func NewErrMalformedJSON(err error) *ErrMalformedJSON {
	return &ErrMalformedJSON{fmt.Sprintf("Malformed JSON payload: %s.", err.Error())}
}
