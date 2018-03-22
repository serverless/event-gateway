package httpapi

// Response is a generic response object from Configuration and Events API.
type Response struct {
	Errors []Error `json:"errors"`
}

// Error represents generic HTTP error returned by Configuration and Events API.
type Error struct {
	Message string `json:"message"`
}
