package cors

// CORS is used to configure CORS on HTTP subscriptions
type CORS struct {
	Origins          []string `json:"origins" validate:"min=1"`
	Methods          []string `json:"methods" validate:"min=1"`
	Headers          []string `json:"headers" validate:"min=1"`
	AllowCredentials bool     `json:"allowCredentials"`
}
