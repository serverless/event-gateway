package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/event-gateway/subscription/cors"
)

// HTTPAPI exposes REST API for configuring EG
type HTTPAPI struct {
	EventTypes    event.Service
	Functions     function.Service
	Subscriptions subscription.Service
	CORSes        cors.Service
}

// EventTypesResponse is a HTTPAPI JSON response containing event types.
type EventTypesResponse struct {
	EventTypes event.Types `json:"eventTypes"`
}

// FunctionsResponse is a HTTPAPI JSON response containing functions.
type FunctionsResponse struct {
	Functions function.Functions `json:"functions"`
}

// SubscriptionsResponse is a HTTPAPI JSON response containing subscriptions.
type SubscriptionsResponse struct {
	Subscriptions subscription.Subscriptions `json:"subscriptions"`
}

// CORSResponse is a HTTPAPI JSON response containing cors configuration.
type CORSResponse struct {
	CORSes cors.CORSes `json:"cors"`
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.GET("/v1/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.Handler("GET", "/v1/metrics", promhttp.Handler())

	router.GET("/v1/spaces/:space/eventtypes", h.listEventTypes)
	router.GET("/v1/spaces/:space/eventtypes/:name", h.getEventType)
	router.POST("/v1/spaces/:space/eventtypes", h.createEventType)
	router.PUT("/v1/spaces/:space/eventtypes/:name", h.updateEventType)
	router.DELETE("/v1/spaces/:space/eventtypes/:name", h.deleteEventType)

	router.GET("/v1/spaces/:space/functions", h.listFunctions)
	router.GET("/v1/spaces/:space/functions/:id", h.getFunction)
	router.POST("/v1/spaces/:space/functions", h.createFunction)
	router.PUT("/v1/spaces/:space/functions/:id", h.updateFunction)
	router.DELETE("/v1/spaces/:space/functions/:id", h.deleteFunction)

	router.GET("/v1/spaces/:space/subscriptions", h.listSubscriptions)
	router.GET("/v1/spaces/:space/subscriptions/:id", h.getSubscription)
	router.POST("/v1/spaces/:space/subscriptions", h.createSubscription)
	router.PUT("/v1/spaces/:space/subscriptions/:id", h.updateSubscription)
	router.DELETE("/v1/spaces/:space/subscriptions/:id", h.deleteSubscription)

	router.GET("/v1/spaces/:space/cors", h.listCORS)
	router.GET("/v1/spaces/:space/cors/*id", h.getCORS)
	router.POST("/v1/spaces/:space/cors", h.createCORS)
	router.PUT("/v1/spaces/:space/cors/*id", h.updateCORS)
	router.DELETE("/v1/spaces/:space/cors/*id", h.deleteCORS)
}

func (h HTTPAPI) getEventType(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	fn, err := h.EventTypes.GetEventType(space, event.TypeName(params.ByName("name")))
	if err != nil {
		if _, ok := err.(*event.ErrEventTypeNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(fn)
	}

	metricConfigRequests.WithLabelValues(space, "eventtype", "get").Inc()
}

func (h HTTPAPI) listEventTypes(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	types, err := h.EventTypes.ListEventTypes(space)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(&EventTypesResponse{types})
	}

	metricConfigRequests.WithLabelValues(space, "eventtype", "list").Inc()
}

func (h HTTPAPI) createEventType(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	eventType := &event.Type{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(eventType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		validationErr := event.ErrEventTypeValidation{Message: err.Error()}
		encoder.Encode(&Response{Errors: []Error{{Message: validationErr.Error()}}})
		return
	}

	eventType.Space = params.ByName("space")
	output, err := h.EventTypes.CreateEventType(eventType)
	if err != nil {
		if _, ok := err.(*event.ErrEventTypeValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*event.ErrEventTypeAlreadyExists); ok {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusCreated)
		encoder.Encode(output)

		metricEventTypes.WithLabelValues(eventType.Space).Inc()
	}

	metricConfigRequests.WithLabelValues(eventType.Space, "eventtype", "create").Inc()
}

func (h HTTPAPI) updateEventType(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	eventType := &event.Type{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(eventType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		validationErr := event.ErrEventTypeValidation{Message: err.Error()}
		encoder.Encode(&Response{Errors: []Error{{Message: validationErr.Error()}}})
		return
	}

	eventType.Space = params.ByName("space")
	eventType.Name = event.TypeName(params.ByName("name"))
	output, err := h.EventTypes.UpdateEventType(eventType)
	if err != nil {
		if _, ok := err.(*event.ErrEventTypeNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else if _, ok := err.(*event.ErrEventTypeValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*event.ErrAuthorizerDoesNotExists); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusOK)
		encoder.Encode(output)
	}

	metricConfigRequests.WithLabelValues(eventType.Space, "eventtype", "update").Inc()
}

func (h HTTPAPI) deleteEventType(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	err := h.EventTypes.DeleteEventType(space, event.TypeName(params.ByName("name")))
	if err != nil {
		if _, ok := err.(*event.ErrEventTypeNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else if _, ok := err.(*event.ErrEventTypeHasSubscriptions); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusNoContent)

		metricEventTypes.WithLabelValues(space).Dec()
	}

	metricConfigRequests.WithLabelValues(space, "eventtype", "delete").Inc()
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
	fns, err := h.Functions.ListFunctions(space)
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
	output, err := h.Functions.CreateFunction(fn)
	if err != nil {
		if _, ok := err.(*function.ErrFunctionValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*function.ErrFunctionAlreadyRegistered); ok {
			w.WriteHeader(http.StatusConflict)
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
		} else if _, ok := err.(*function.ErrFunctionHasSubscriptions); ok {
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
	subs, err := h.Subscriptions.ListSubscriptions(space)
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
	fn, err := h.Subscriptions.GetSubscription(space, subscription.ID(params.ByName("id")))
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
			w.WriteHeader(http.StatusConflict)
		} else if _, ok := err.(*event.ErrEventTypeNotFound); ok {
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
	s.ID = subscription.ID(params.ByName("id"))

	output, err := h.Subscriptions.UpdateSubscription(s.ID, s)
	if err != nil {
		if _, ok := err.(*subscription.ErrInvalidSubscriptionUpdate); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*subscription.ErrSubscriptionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
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
	err := h.Subscriptions.DeleteSubscription(space, subscription.ID(params.ByName("id")))
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

func (h HTTPAPI) listCORS(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	configs, err := h.CORSes.ListCORS(space)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(&CORSResponse{configs})
	}

	metricConfigRequests.WithLabelValues(space, "cors", "list").Inc()
}

func (h HTTPAPI) getCORS(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	config, err := h.CORSes.GetCORS(space, extractCORSID(r.URL.RawPath))
	if err != nil {
		if _, ok := err.(*cors.ErrCORSNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		encoder.Encode(config)
	}

	metricConfigRequests.WithLabelValues(space, "cors", "get").Inc()
}

func (h HTTPAPI) createCORS(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	config := &cors.CORS{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(config)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		validationErr := cors.ErrCORSValidation{Message: err.Error()}
		encoder.Encode(&Response{Errors: []Error{{Message: validationErr.Error()}}})
		return
	}

	config.Space = params.ByName("space")
	output, err := h.CORSes.CreateCORS(config)
	if err != nil {
		if _, ok := err.(*cors.ErrCORSAlreadyExists); ok {
			w.WriteHeader(http.StatusConflict)
		} else if _, ok := err.(*cors.ErrCORSValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusCreated)
		encoder.Encode(output)

		metricCORS.WithLabelValues(config.Space).Inc()
	}

	metricConfigRequests.WithLabelValues(config.Space, "cors", "create").Inc()
}

func (h HTTPAPI) updateCORS(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	config := &cors.CORS{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(config)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		validationErr := cors.ErrCORSValidation{Message: err.Error()}
		encoder.Encode(&Response{Errors: []Error{{Message: validationErr.Error()}}})
		return
	}

	config.Space = params.ByName("space")
	config.ID = extractCORSID(r.URL.RawPath)

	output, err := h.CORSes.UpdateCORS(config)
	if err != nil {
		if _, ok := err.(*cors.ErrCORSNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else if _, ok := err.(*cors.ErrCORSValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusOK)
		encoder.Encode(output)
	}

	metricConfigRequests.WithLabelValues(config.Space, "cors", "update").Inc()
}

func (h HTTPAPI) deleteCORS(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	space := params.ByName("space")
	err := h.CORSes.DeleteCORS(space, extractCORSID(r.URL.RawPath))
	if err != nil {
		if _, ok := err.(*cors.ErrCORSNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		encoder.Encode(&Response{Errors: []Error{{Message: err.Error()}}})
	} else {
		w.WriteHeader(http.StatusNoContent)

		metricCORS.WithLabelValues(space).Dec()
	}

	metricConfigRequests.WithLabelValues(space, "cors", "delete").Inc()
}

// httprouter weirdness: params are based on Request.URL.Path, not Request.URL.RawPath
func extractCORSID(rawPath string) cors.ID {
	segments := strings.Split(rawPath, "/")
	id := segments[len(segments)-1]
	return cors.ID(id)
}
