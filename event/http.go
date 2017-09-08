package event

// HTTPEvent is a event schema used for sending events to HTTP subscriptions.
type HTTPEvent struct {
	Headers map[string][]string `json:"headers"`
	Query   map[string][]string `json:"query"`
	Body    interface{}         `json:"body"`
	Path    string              `json:"path"`
	Method  string              `json:"method"`
	Params  map[string]string   `json:"params"`
}
