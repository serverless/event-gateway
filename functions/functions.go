package functions

import (
	"bytes"
	"encoding/gob"

	"github.com/serverless/gateway/db"
)

// Functions is a discovery tool for FaaS functions
type Functions struct {
	DB *db.DB
}

// Function registered in function discovery
type Function struct {
	ID        string     `json:"id"`
	Instances []Instance `json:"instances"`
}

// Instance of function. A function can have multiple instances. Each instance of a function is deployed in different regions.
type Instance struct {
	Provider string `json:"provider"`
	OriginID string `json:"originId"`
	Region   string `json:"region"`
}

// RegisterFunction registers function to Function discovery
func (f *Functions) RegisterFunction(fn *Function) (*Function, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(fn)
	if err != nil {
		return nil, err
	}

	err = f.DB.Set(bucket, fn.ID, buf.Bytes())
	if err != nil {
		return nil, err
	}

	return fn, nil
}

// GetFunction returns function from the discovery
func (f *Functions) GetFunction(name string) (*Function, error) {
	value, err := f.DB.Get(bucket, name)
	if err != nil {
		return nil, err
	}

	fn := new(Function)
	buf := bytes.NewBuffer(value)
	err = gob.NewDecoder(buf).Decode(fn)
	if err != nil {
		return nil, err
	}
	return fn, nil
}

const bucket = "functions"
