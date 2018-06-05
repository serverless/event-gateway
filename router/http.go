package router

// HTTPResponse is a response schema returned by subscribed function in case of HTTP event.
type HTTPResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}
