package endpoints

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/docker/libkv/store"
	shortid "github.com/ventu-io/go-shortid"
	"go.uber.org/zap"

	"github.com/serverless/gateway/functions"
)

// EndpointID uniquely identifies an endpoint
type EndpointID string

// Endpoint represents single endpoint
type Endpoint struct {
	ID         EndpointID           `json:"id"`
	FunctionID functions.FunctionID `json:"functionId"`
	Method     string               `json:"method"`
	Path       string               `json:"path"`
}

// Endpoints enable exposing public HTTP/REST endpoints that allow communicating with backend functions.
type Endpoints struct {
	sync.RWMutex
	DB     store.Store
	Logger *zap.Logger
}

// GetEndpoint returns registered endpoint.
func (e *Endpoints) GetEndpoint(name string) (*Endpoint, error) {
	kv, err := e.DB.Get(name)
	if err != nil {
		return nil, err
	}

	if len(kv.Value) == 0 {
		return nil, &ErrorNotFound{name}
	}

	endpoint := Endpoint{}
	dec := json.NewDecoder(bytes.NewReader(kv.Value))
	err = dec.Decode(&endpoint)
	if err != nil {
		e.Logger.Info("Fetching endpoint failed.", zap.Error(err))
		return nil, err
	}

	return &endpoint, nil
}

// CreateEndpoint creates endpoint.
func (e *Endpoints) CreateEndpoint(en *Endpoint) (*Endpoint, error) {
	id, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	en.ID = EndpointID(id)

	buf, err := json.Marshal(en)
	if err != nil {
		return nil, err
	}

	err = e.DB.Put(string(en.ID), buf, nil)
	if err != nil {
		return nil, err
	}

	return en, nil
}
