package endpoints

import (
	"bytes"
	"encoding/gob"
	"strings"

	"go.uber.org/zap"

	"github.com/serverless/gateway/db"
	shortid "github.com/ventu-io/go-shortid"
)

// Endpoints enable exposing public HTTP/REST endpoints that allow communicating with backend functions.
type Endpoints struct {
	DB      *db.DB
	Invoker Invoker
	Logger  *zap.Logger
}

// Endpoint represents single endpoint
type Endpoint struct {
	ID        string           `json:"id"`
	Functions []FunctionTarget `json:"functions"`
}

// FunctionTarget is a function exposed by Endpoints
type FunctionTarget struct {
	FunctionID string `json:"functionId"`
	Method     string `json:"method"`
	Path       string `json:"path"`
}

// Invoker invokes function from function discovery
type Invoker interface {
	Invoke(name string, payload []byte) ([]byte, error)
}

// GetEndpoint returns registered endpoint.
func (e *Endpoints) GetEndpoint(name string) (*Endpoint, error) {
	value, err := e.DB.Get(bucket, name)
	if err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return nil, &ErrorNotFound{name}
	}

	fn := new(Endpoint)
	buf := bytes.NewBuffer(value)
	err = gob.NewDecoder(buf).Decode(fn)
	if err != nil {
		e.Logger.Info("fetching endpoint failed", zap.Error(err))
		return nil, err
	}
	return fn, nil
}

// CreateEndpoint creates endpoint.
func (e *Endpoints) CreateEndpoint(en *Endpoint) (*Endpoint, error) {
	id, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	en.ID = id

	buf := new(bytes.Buffer)
	err = gob.NewEncoder(buf).Encode(en)
	if err != nil {
		return nil, err
	}

	err = e.DB.Set(bucket, en.ID, buf.Bytes())
	if err != nil {
		return nil, err
	}

	return en, nil
}

// CallEndpoint calls registered endpoints.
func (e *Endpoints) CallEndpoint(name, method, path string, payload []byte) ([]byte, error) {
	en, err := e.GetEndpoint(name)
	if err != nil {
		return nil, err
	}

	for _, fn := range en.Functions {
		if fn.Method == strings.ToLower(method) && fn.Path == path {
			return e.Invoker.Invoke(fn.FunctionID, payload)
		}
	}

	return nil, &ErrorTargetNotFound{name}
}

const bucket = "endpoints"
