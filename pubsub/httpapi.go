package pubsub

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/util/httpapi"
)

// HTTPAPI for pubsub sub-service
type HTTPAPI struct {
	PubSub *PubSub
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.POST("/v1/subscriptions", h.createSubscription)
	router.DELETE("/v1/subscriptions/:subscriptionID", h.deleteSubscription)
	router.GET("/v1/subscriptions", h.getSubscriptions)
}

func (h HTTPAPI) createSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	s := &Subscription{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(s)

	output, err := h.PubSub.CreateSubscription(s)
	if err != nil {
		if _, ok := err.(*ErrorSubscriptionAlreadyExists); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorFunctionNotFound); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorSubscriptionValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(output)
	}
}

func (h HTTPAPI) deleteSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	err := h.PubSub.DeleteSubscription(SubscriptionID(params.ByName("subscriptionID")))
	if err != nil {
		if _, ok := err.(*ErrorSubscriptionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	}
}

func (h HTTPAPI) getSubscriptions(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	subs, err := h.PubSub.GetAllSubscriptions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(&subscriptions{subs})
	}
}

type subscriptions struct {
	Subscriptions []*Subscription `json:"subscriptions"`
}
