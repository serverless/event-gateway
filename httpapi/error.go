package httpapi

// Error represents generic HTTP error returned by Configuration API.
type Error struct {
	Error string `json:"error"`
}
