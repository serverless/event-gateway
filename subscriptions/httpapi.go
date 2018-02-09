package subscriptions

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/internal/httpapi"
)

// HTTPAPI for subscriptions sub-service
type HTTPAPI struct {
	Subscriptions *Subscriptions
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.POST("/v1/subscriptions", h.createSubscription)
	router.DELETE("/v1/subscriptions/*subscriptionID", h.deleteSubscription)
	router.GET("/v1/subscriptions", h.getSubscriptions)
}

func (h HTTPAPI) createSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	s := &Subscription{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(httpapi.NewErrMalformedJSON(err))
		return
	}

	output, err := h.Subscriptions.CreateSubscription(s)
	if err != nil {
		if _, ok := err.(*ErrSubscriptionAlreadyExists); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrFunctionNotFound); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrSubscriptionValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrPathConfict); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{ Message: err.Error() }}})
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

	err := h.Subscriptions.DeleteSubscription(SubscriptionID(sid))
	if err != nil {
		if _, ok := err.(*ErrSubscriptionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{ Message: err.Error() }}})
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
		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{ Message: err.Error() }}})
	} else {
		encoder.Encode(&subscriptions{subs})
	}
}

type subscriptions struct {
	Subscriptions []*Subscription `json:"subscriptions"`
}
