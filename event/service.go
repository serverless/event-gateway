package event

// Service represents service for managing event types.
type Service interface {
	CreateEventType(eventType *Type) (*Type, error)
	UpdateEventType(name TypeName, eventType *Type) (*Type, error)
	GetEventType(space string, name TypeName) (*Type, error)
	GetEventTypes(space string) (Types, error)
	DeleteEventType(space string, name TypeName) error
}
