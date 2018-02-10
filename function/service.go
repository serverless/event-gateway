package function

// Service represents service for managing functions.
type Service interface {
	RegisterFunction(fn *Function) (*Function, error)
	UpdateFunction(fn *Function) (*Function, error)
	GetFunction(id ID) (*Function, error)
	GetAllFunctions() ([]*Function, error)
	DeleteFunction(id ID) error
}
