package cors

// Service represents service for managing CORS configuration.
type Service interface {
	CreateCORS(c *CORS) (*CORS, error)
	UpdateCORS(c *CORS) (*CORS, error)
	GetCORS(space string, id ID) (*CORS, error)
	GetCORSes(space string) (CORSes, error)
	DeleteCORS(space string, id ID) error
}
