package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/httpapi"
)

// HTTPAPI for endpints sub-service
type HTTPAPI struct {
	Endpoints *Endpoints
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.POST("/v0/gateway/api/endpoint", h.createEndpoint)
	router.DELETE("/v0/gateway/api/endpoint/:id", h.deleteEndpoint)
	router.GET("/v0/gateway/api/endpoint", h.getEndpoints)
}

func (h HTTPAPI) createEndpoint(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	en := &Endpoint{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(en)

	output, err := h.Endpoints.Create(en)
	if err != nil {
		if _, ok := err.(*ErrorFunctionNotFound); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorAlreadyExists); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(output)
	}
}

func (h HTTPAPI) deleteEndpoint(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	err := h.Endpoints.Delete(EndpointID(params.ByName("id")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	}
}

func (h HTTPAPI) getEndpoints(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	ens, err := h.Endpoints.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(&endpoints{ens})
	}
}

type endpoints struct {
	Endpoints []*Endpoint `json:"endpoints"`
}
