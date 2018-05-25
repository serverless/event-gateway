package event

const (
	// TypeHTTPRequest is a special type of event http requests that are not CloudEvents.
	TypeHTTPRequest = TypeName("http.request")
)

// TypeName uniquely identifies an event type.
type TypeName string

// Type is a registered event type.
type Type struct {
	Space string   `json:"space" validate:"required,min=3,space"`
	Name  TypeName `json:"name" validate:"required"`
}

// Types is an array of subscriptions.
type Types []*Type
