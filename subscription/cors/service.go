package cors

// Service represents service for managing CORS configuration.
type Service interface {
	GetCORS(space string, id ID) (*CORS, error)
	ListCORS(space string) (CORSes, error)
	CreateCORS(c *CORS) (*CORS, error)
	UpdateCORS(c *CORS) (*CORS, error)
	DeleteCORS(space string, id ID) error
}
