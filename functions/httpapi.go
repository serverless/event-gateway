package functions

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/internal/httpapi"
)

// HTTPAPI for function discovery
type HTTPAPI struct {
	Functions *Functions
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.GET("/v1/functions", h.getFunctions)
	router.POST("/v1/functions", h.registerFunction)
	router.GET("/v1/functions/:id", h.getFunction)
	router.PUT("/v1/functions/:id", h.updateFunction)
	router.DELETE("/v1/functions/:id", h.deleteFunction)
}

func (h HTTPAPI) getFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn, err := h.Functions.GetFunction(FunctionID(params.ByName("id")))
	if err != nil {
		if _, ok := err.(*ErrNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{ Message: err.Error() }}})
	} else {
		encoder.Encode(fn)
	}
}

func (h HTTPAPI) getFunctions(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fns, err := h.Functions.GetAllFunctions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{ Message: err.Error() }}})
	} else {
		encoder.Encode(&functions{fns})
	}
}

type functions struct {
	Functions []*Function `json:"functions"`
}

func (h HTTPAPI) registerFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn := &Function{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(fn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(httpapi.NewErrMalformedJSON(err))
		return
	}

	output, err := h.Functions.RegisterFunction(fn)
	if err != nil {
		if _, ok := err.(*ErrValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrAlreadyRegistered); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{ Message: err.Error() }}})
		return
	}

	encoder.Encode(output)
}

func (h HTTPAPI) updateFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn := &Function{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(fn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(httpapi.NewErrMalformedJSON(err))
		return
	}

	fn.ID = FunctionID(params.ByName("id"))
	output, err := h.Functions.UpdateFunction(fn)
	if err != nil {
		if _, ok := err.(*ErrValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{ Message: err.Error() }}})
		return
	}

	encoder.Encode(output)
}

func (h HTTPAPI) deleteFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	err := h.Functions.DeleteFunction(FunctionID(params.ByName("id")))
	if err != nil {
		if _, ok := err.(*ErrNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{ Message: err.Error() }}})
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
