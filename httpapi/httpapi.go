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

// FunctionsResponse is a HTTPAPI JSON response containing functions.
type FunctionsResponse struct {
	Functions function.Functions `json:"functions"`
}

// SubscriptionsResponse is a HTTPAPI JSON response containing subscriptions.
type SubscriptionsResponse struct {
	Subscriptions subscription.Subscriptions `json:"subscriptions"`
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.GET("/v1/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.Handler("GET", "/metrics", promhttp.Handler())

	router.GET("/v1/spaces/:space/functions", h.listFunctions)
	router.GET("/v1/spaces/:space/functions/:id", h.getFunction)
	router.POST("/v1/spaces/:space/functions", h.createFunction)
	router.PUT("/v1/spaces/:space/functions/:id", h.updateFunction)
	router.DELETE("/v1/spaces/:space/functions/:id", h.deleteFunction)

	router.GET("/v1/spaces/:space/subscriptions", h.listSubscriptions)
	router.GET("/v1/spaces/:space/subscriptions/*id", h.getSubscription)
	router.POST("/v1/spaces/:space/subscriptions", h.createSubscription)
	router.PUT("/v1/spaces/:space/subscriptions/*id", h.updateSubscription)
	router.DELETE("/v1/spaces/:space/subscriptions/*id", h.deleteSubscription)
}

func (h HTTPAPI) getFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	fn, err := h.Functions.GetFunction(space, function.ID(params.ByName("id")))
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

	metricConfigRequests.WithLabelValues(space, "function", "get").Inc()
}

func (h HTTPAPI) listFunctions(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	fns, err := h.Functions.GetFunctions(space)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(&FunctionsResponse{fns})
	}

	metricConfigRequests.WithLabelValues(space, "function", "list").Inc()
}

func (h HTTPAPI) createFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn := &function.Function{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(fn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		validationErr := function.ErrFunctionValidation{Message: err.Error()}
		encoder.Encode(&Response{Errors: []Error{{Message: validationErr.Error()}}})
		return
	}

	fn.Space = params.ByName("space")
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
	} else {
		w.WriteHeader(http.StatusCreated)
		encoder.Encode(output)

		metricFunctions.WithLabelValues(fn.Space).Inc()
	}

	metricConfigRequests.WithLabelValues(fn.Space, "function", "create").Inc()
}

func (h HTTPAPI) updateFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	fn := &function.Function{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(fn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		validationErr := function.ErrFunctionValidation{Message: err.Error()}
		encoder.Encode(&Response{Errors: []Error{{Message: validationErr.Error()}}})
		return
	}

	fn.Space = params.ByName("space")
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
	} else {
		encoder.Encode(output)
	}

	metricConfigRequests.WithLabelValues(fn.Space, "function", "update").Inc()
}

func (h HTTPAPI) deleteFunction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	err := h.Functions.DeleteFunction(space, function.ID(params.ByName("id")))
	if err != nil {
		if _, ok := err.(*function.ErrFunctionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else if _, ok := err.(*function.ErrFunctionHasSubscriptionsError); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusNoContent)

		metricFunctions.WithLabelValues(space).Dec()
	}

	metricConfigRequests.WithLabelValues(space, "function", "delete").Inc()
}

func (h HTTPAPI) listSubscriptions(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	subs, err := h.Subscriptions.GetSubscriptions(space)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(&SubscriptionsResponse{subs})
	}

	metricConfigRequests.WithLabelValues(space, "subscription", "list").Inc()
}

func (h HTTPAPI) getSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	fn, err := h.Subscriptions.GetSubscription(space, extractSubscriptionID(r.URL.RawPath))
	if err != nil {
		if _, ok := err.(*subscription.ErrSubscriptionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(fn)
	}

	metricConfigRequests.WithLabelValues(space, "subscription", "get").Inc()
}

func (h HTTPAPI) createSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	s := &subscription.Subscription{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		validationErr := subscription.ErrSubscriptionValidation{Message: err.Error()}
		encoder.Encode(&Response{Errors: []Error{{Message: validationErr.Error()}}})
		return
	}

	s.Space = params.ByName("space")
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
	} else {
		w.WriteHeader(http.StatusCreated)
		encoder.Encode(output)

		metricSubscriptions.WithLabelValues(s.Space).Inc()
	}

	metricConfigRequests.WithLabelValues(s.Space, "subscription", "create").Inc()
}

func (h HTTPAPI) updateSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	s := &subscription.Subscription{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		validationErr := subscription.ErrSubscriptionValidation{Message: err.Error()}
		encoder.Encode(&Response{Errors: []Error{{Message: validationErr.Error()}}})
		return
	}

	s.Space = params.ByName("space")
	s.ID = extractSubscriptionID(r.URL.RawPath)
	output, err := h.Subscriptions.UpdateSubscription(s.ID, s)
	if err != nil {
		if _, ok := err.(*subscription.ErrInvalidSubscriptionUpdate); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*subscription.ErrSubscriptionNotFound); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*function.ErrFunctionNotFound); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*subscription.ErrSubscriptionValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusOK)
		encoder.Encode(output)
	}

	metricConfigRequests.WithLabelValues(s.Space, "subscription", "update").Inc()
}

func (h HTTPAPI) deleteSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	err := h.Subscriptions.DeleteSubscription(space, extractSubscriptionID(r.URL.RawPath))
	if err != nil {
		if _, ok := err.(*subscription.ErrSubscriptionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusNoContent)

		metricSubscriptions.WithLabelValues(space).Dec()
	}

	metricConfigRequests.WithLabelValues(space, "subscription", "delete").Inc()
}

// httprouter weirdness: params are based on Request.URL.Path, not Request.URL.RawPath
func extractSubscriptionID(rawPath string) subscription.ID {
	segments := strings.Split(rawPath, "/")
	sid := segments[len(segments)-1]
	return subscription.ID(sid)
}
