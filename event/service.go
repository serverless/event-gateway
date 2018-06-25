package event

// Service represents service for managing event types.
type Service interface {
	GetEventType(space string, name TypeName) (*Type, error)
	ListEventTypes(space string) (Types, error)
	CreateEventType(eventType *Type) (*Type, error)
	UpdateEventType(newEventType *Type) (*Type, error)
	DeleteEventType(space string, name TypeName) error
}
