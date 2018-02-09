package httpapi

import "fmt"

type Response struct {
	Errors []Error `json:"errors"`
}

// Error represents generic HTTP error returned by Configuration API.
type Error struct {
	Message string `json:"message"`
}

// ErrMalformedJSON occurring when it's impossible to decode JSON payload.
type ErrMalformedJSON Error

// NewErrMalformedJSON creates ErrMalformedJSON.
func NewErrMalformedJSON(err error) *ErrMalformedJSON {
	return &ErrMalformedJSON{fmt.Sprintf("Malformed JSON payload: %s.", err.Error())}
}
