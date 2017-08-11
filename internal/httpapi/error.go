package httpapi

import "fmt"

// Error represents generic HTTP error returned by Configuration API.
type Error struct {
	Error string `json:"error"`
}

// ErrorMalformedJSON occuring when it's impossible to decode JSON payload.
type ErrorMalformedJSON Error

// NewErrorMalformedJSON creates ErrorMalformedJSON.
func NewErrorMalformedJSON(err error) *ErrorMalformedJSON {
	return &ErrorMalformedJSON{fmt.Sprintf("Malformed JSON payload: %s.", err.Error())}
}
