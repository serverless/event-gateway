package endpoints

import (
	"bytes"
	"encoding/gob"
	"sync"

	shortid "github.com/ventu-io/go-shortid"

	"go.uber.org/zap"

	"github.com/serverless/gateway/db"
	"github.com/serverless/gateway/endpoints/types"
	"github.com/serverless/gateway/targetcache"
)

// Endpoints enable exposing public HTTP/REST endpoints that allow communicating with backend functions.
type Endpoints struct {
	sync.RWMutex
	DB          *db.PrefixedStore
	TargetCache targetcache.TargetCache
	Logger      *zap.Logger
}

// GetEndpoint returns registered endpoint.
func (e *Endpoints) GetEndpoint(name string) (*types.Endpoint, error) {
	kv, err := e.DB.Get(name)
	if err != nil {
		return nil, err
	}

	if len(kv.Value) == 0 {
		return nil, &ErrorNotFound{name}
	}

	endpoint := types.Endpoint{}
	buf := bytes.NewBuffer(kv.Value)
	err = gob.NewDecoder(buf).Decode(&endpoint)
	if err != nil {
		e.Logger.Info("Fetching endpoint failed.", zap.Error(err))
		return nil, err
	}
	return &endpoint, nil
}

// CreateEndpoint creates endpoint.
func (e *Endpoints) CreateEndpoint(en *types.Endpoint) (*types.Endpoint, error) {
	id, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	en.ID = id

	buf := &bytes.Buffer{}
	err = gob.NewEncoder(buf).Encode(en)
	if err != nil {
		return nil, err
	}

	err = e.DB.Put(en.ID, buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	return en, nil
}

// CallEndpoint calls registered endpoints.
func (e *Endpoints) CallEndpoint(name, method, path string, payload []byte) ([]byte, error) {
	_, err := e.GetEndpoint(name)
	if err != nil {
		return nil, err
	}

	return nil, &ErrorTargetNotFound{name}
}
