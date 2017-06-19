package functions

import (
	"encoding/json"
	"io/ioutil"
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
	router.POST("/v0/gateway/api/invoke/:name", h.invoke)
}

func (h HTTPAPI) getFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fn, err := h.Functions.GetFunction(params.ByName("name"))
	if err != nil {
		if _, ok := err.(*ErrorNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(fn)
	}
}

func (h HTTPAPI) registerFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fn := new(Function)
	dec := json.NewDecoder(r.Body)
	dec.Decode(fn)

	output, err := h.Functions.RegisterFunction(fn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(output)
	}
}

func (h HTTPAPI) invoke(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	payload, _ := ioutil.ReadAll(r.Body)
	output, err := h.Functions.Invoke(params.ByName("name"), payload)
	if err != nil {
		if _, ok := err.(*ErrorNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	}
}
