package functions

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// HTTPAPI for function discovery
type HTTPAPI struct {
	Functions *Functions
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.GET("/v0/gateway/api/function/:name", h.getFunction)
	router.POST("/v0/gateway/api/function", h.registerFunction)
}

func (h HTTPAPI) getFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn, err := h.Functions.GetFunction(params.ByName("name"))
	if err != nil {
		if _, ok := err.(*ErrorNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpError{err.Error()})
	} else {
		encoder.Encode(fn)
	}
}

func (h HTTPAPI) registerFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn := &Function{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(fn)

	output, err := h.Functions.RegisterFunction(fn)
	if err != nil {
		if _, ok := err.(*ErrorPropertiesNotSpecified); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorOneFunctionTypeCanBeSpecified); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpError{err.Error()})
	} else {
		encoder.Encode(output)
	}
}

type httpError struct {
	Error string `json:"error"`
}
