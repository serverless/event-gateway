package endpoints

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/serverless/gateway/db"
	shortid "github.com/ventu-io/go-shortid"
)

// Endpoints enable exposing public HTTP/REST endpoints that allow communicating with backend functions.
type Endpoints struct {
	sync.RWMutex
	DB      *db.ReactiveCfgStore
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

// Created is called when a new endpoint is detected in the config.
func (e *Endpoints) Created(key string, value []byte) {
	e.Lock()
	defer e.Unlock()
	// TODO put any necessary endpoint conf initialization code here
	e.Logger.Debug("Received Created event.",
		zap.String("key", key),
		zap.String("value", string(value)))
}

// Modified is called when an existing endpoint is modified in the config.
func (e *Endpoints) Modified(key string, newValue []byte) {
	e.Lock()
	defer e.Unlock()
	// TODO put any necessary endpoint conf modification code here
	e.Logger.Debug("Received Modified event.",
		zap.String("key", key),
		zap.String("newValue", string(newValue)))
}

// Deleted is called when an endpoint is deleted in the config.
func (e *Endpoints) Deleted(key string, lastKnownValue []byte) {
	e.Lock()
	defer e.Unlock()
	// TODO put any necessary endpoint conf deletion code here
	e.Logger.Debug("Received Deleted event.",
		zap.String("key", key),
		zap.String("lastKnownValue", string(lastKnownValue)))
}

// GetEndpoint returns registered endpoint.
func (e *Endpoints) GetEndpoint(name string) (*Endpoint, error) {
	value, err := e.DB.CachedGet(name)
	if err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return nil, &ErrorNotFound{name}
	}

	en := &Endpoint{}
	dec := json.NewDecoder(bytes.NewReader(value))
	err = dec.Decode(en)
	if err != nil {
		e.Logger.Info("Fetching endpoint failed.", zap.Error(err))
		return nil, err
	}
	return en, nil
}

// CreateEndpoint creates endpoint.
func (e *Endpoints) CreateEndpoint(en *Endpoint) (*Endpoint, error) {
	id, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	en.ID = id

	byt, err := json.Marshal(en)
	if err != nil {
		return nil, err
	}
	err = e.DB.Put(en.ID, byt, nil)
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
