package endpoints

import (
	"bytes"
	"encoding/json"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/functions"
)

// EndpointID uniquely identifies an endpoint
type EndpointID string

// Endpoint represents single endpoint
type Endpoint struct {
	ID         EndpointID           `json:"endpointId"`
	FunctionID functions.FunctionID `json:"functionId" validate:"required"`
	Method     string               `json:"method" validate:"required,eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
	Path       string               `json:"path" validate:"required"`
}

// FunctionExister is an interface used to check if function exists in the discovery.
type FunctionExister interface {
	Exist(name string) bool
}

// Endpoints enable exposing public HTTP/REST endpoints that allow communicating with backend functions.
type Endpoints struct {
	DB     store.Store
	Logger *zap.Logger
	FunctionExister
}

// Create creates endpoint.
func (e *Endpoints) Create(en *Endpoint) (*Endpoint, error) {
	validate := validator.New()
	err := validate.Struct(en)
	if err != nil {
		return nil, &ErrorValidation{err}
	}

	exists := e.FunctionExister.Exist(string(en.FunctionID))
	if !exists {
		return nil, &ErrorFunctionNotFound{string(en.FunctionID)}
	}

	en.Path = strings.TrimPrefix(en.Path, "/")
	en.ID = EndpointID(en.Method + "-" + en.Path)

	_, err = e.DB.Get(string(en.ID))
	if err == nil {
		return nil, &ErrorAlreadyExists{
			Method: en.Method,
			Path:   en.Path,
		}
	}

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

// Delete endpoint.
func (e *Endpoints) Delete(id EndpointID) error {
	err := e.DB.Delete(string(id))
	if err != nil {
		return &ErrorNotFound{string(id)}
	}
	return nil
}

// GetAll returns array of all Endpoints.
func (e *Endpoints) GetAll() ([]*Endpoint, error) {
	ens := []*Endpoint{}

	kvs, err := e.DB.List("")
	if err != nil {
		return ens, nil
	}

	for _, kv := range kvs {
		en := &Endpoint{}
		dec := json.NewDecoder(bytes.NewReader(kv.Value))
		err = dec.Decode(en)
		if err != nil {
			return nil, err
		}

		ens = append(ens, en)
	}

	return ens, nil
}
