package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/subscription"
)

// HTTPAPI exposes REST API for configuring EG
type HTTPAPI struct {
	Functions     function.Service
	Subscriptions subscription.Service
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.GET("/v1/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.Handler("GET", "/metrics", promhttp.Handler())

	router.GET("/v1/functions", h.getFunctions)
	router.POST("/v1/functions", h.registerFunction)
	router.GET("/v1/functions/:id", h.getFunction)
	router.PUT("/v1/functions/:id", h.updateFunction)
	router.DELETE("/v1/functions/:id", h.deleteFunction)

	router.POST("/v1/subscriptions", h.createSubscription)
	router.DELETE("/v1/subscriptions/*subscriptionID", h.deleteSubscription)
	router.GET("/v1/subscriptions", h.getSubscriptions)
}

func (h HTTPAPI) getFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn, err := h.Functions.GetFunction(function.ID(params.ByName("id")))
	if err != nil {
		if _, ok := err.(*function.ErrFunctionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
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
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(&function.Functions{Functions: fns})
	}
}

func (h HTTPAPI) registerFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn := &function.Function{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(fn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(NewErrMalformedJSON(err))
		return
	}

	output, err := h.Functions.RegisterFunction(fn)
	if err != nil {
		if _, ok := err.(*function.ErrFunctionValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*function.ErrFunctionAlreadyRegistered); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
		return
	}

	encoder.Encode(output)
}

func (h HTTPAPI) updateFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn := &function.Function{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(fn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(NewErrMalformedJSON(err))
		return
	}

	fn.ID = function.ID(params.ByName("id"))
	output, err := h.Functions.UpdateFunction(fn)
	if err != nil {
		if _, ok := err.(*function.ErrFunctionValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*function.ErrFunctionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
		return
	}

	encoder.Encode(output)
}

func (h HTTPAPI) deleteFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	err := h.Functions.DeleteFunction(function.ID(params.ByName("id")))
	if err != nil {
		if _, ok := err.(*function.ErrFunctionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h HTTPAPI) createSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	s := &subscription.Subscription{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(NewErrMalformedJSON(err))
		return
	}

	output, err := h.Subscriptions.CreateSubscription(s)
	if err != nil {
		if _, ok := err.(*subscription.ErrSubscriptionAlreadyExists); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*function.ErrFunctionNotFound); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*subscription.ErrSubscriptionValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*subscription.ErrPathConfict); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
		return
	}

	encoder.Encode(output)
}

func (h HTTPAPI) deleteSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	// httprouter weirdness: params are based on Request.URL.Path, not Request.URL.RawPath
	segments := strings.Split(r.URL.RawPath, "/")
	sid := segments[len(segments)-1]

	err := h.Subscriptions.DeleteSubscription(subscription.ID(sid))
	if err != nil {
		if _, ok := err.(*subscription.ErrSubscriptionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h HTTPAPI) getSubscriptions(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	subs, err := h.Subscriptions.GetAllSubscriptions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(&subscription.Subscriptions{Subscriptions: subs})
	}
}
