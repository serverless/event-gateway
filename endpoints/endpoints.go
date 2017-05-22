package endpoints

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/serverless/gateway/db"
	shortid "github.com/ventu-io/go-shortid"
)

// Endpoints enable exposing public HTTP/REST endpoints that allow communicating with backend functions.
type Endpoints struct {
	DB *db.DB
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
		log.Printf("fetching endpoint failed: %q", err)
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

const bucket = "endpoints"
