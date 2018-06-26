package cors

import "github.com/serverless/event-gateway/metadata"

// Service represents service for managing CORS configuration.
type Service interface {
	GetCORS(space string, id ID) (*CORS, error)
	ListCORS(space string, filters ...metadata.Filter) (CORSes, error)
	CreateCORS(c *CORS) (*CORS, error)
	UpdateCORS(c *CORS) (*CORS, error)
	DeleteCORS(space string, id ID) error
}
