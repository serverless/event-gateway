package httpapi

import "fmt"

// Response is a generic response object from Configuration and Events API.
type Response struct {
	Errors []Error `json:"errors"`
}

// Error represents generic HTTP error returned by Configuration and Events API.
type Error struct {
	Message string `json:"message"`
}

// ErrMalformedJSON occurring when it's impossible to decode JSON payload.
type ErrMalformedJSON struct {
	Original error
}

func (e ErrMalformedJSON) Error() string {
	return fmt.Sprintf("Malformed JSON payload: %q", e.Original)
}
