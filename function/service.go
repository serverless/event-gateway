package function

// Service represents service for managing functions.
type Service interface {
	GetFunction(space string, id ID) (*Function, error)
	ListFunctions(space string) (Functions, error)
	CreateFunction(fn *Function) (*Function, error)
	UpdateFunction(fn *Function) (*Function, error)
	DeleteFunction(space string, id ID) error
}
