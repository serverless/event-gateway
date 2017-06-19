package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// HTTPAPI for endpints sub-service
type HTTPAPI struct {
	Endpoints *Endpoints
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.GET("/v0/gateway/api/endpoint/:name", h.getEndpoint)
	router.POST("/v0/gateway/api/endpoint", h.createEndpoint)
}

func (h HTTPAPI) getEndpoint(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	en, err := h.Endpoints.GetEndpoint(params.ByName("name"))
	if err != nil {
		if _, ok := err.(*ErrorNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(en)
	}
}

func (h HTTPAPI) createEndpoint(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	en := &Endpoint{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(en)

	output, err := h.Endpoints.CreateEndpoint(en)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(output)
	}
}
