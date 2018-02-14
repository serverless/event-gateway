package function

// Service represents service for managing functions.
type Service interface {
	RegisterFunction(fn *Function) (*Function, error)
	UpdateFunction(space string, fn *Function) (*Function, error)
	GetFunction(space string, id ID) (*Function, error)
	GetFunctions(space string) (Functions, error)
	DeleteFunction(space string, id ID) error
}
