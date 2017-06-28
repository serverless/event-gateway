package pubsub

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/httpapi"
)

// HTTPAPI for pubsub sub-service
type HTTPAPI struct {
	PubSub *PubSub
}

// RegisterRoutes register HTTP API routes
func (h HTTPAPI) RegisterRoutes(router *httprouter.Router) {
	router.POST("/v0/gateway/api/topic", h.createTopic)
	router.DELETE("/v0/gateway/api/topic/:topicID", h.deleteTopic)
	router.GET("/v0/gateway/api/topic", h.getTopics)

	router.POST("/v0/gateway/api/topic/:topicID/subscription", h.createSubscription)
	router.DELETE("/v0/gateway/api/topic/:topicID/subscription/:subscriptionID", h.deleteSubscription)
	router.GET("/v0/gateway/api/topic/:topicID/subscription", h.getSubscriptions)

	router.POST("/v0/gateway/api/topic/:topicID/publisher", h.createPublisher)
	router.DELETE("/v0/gateway/api/topic/:topicID/publisher/:publisherID", h.deletePublisher)
	router.GET("/v0/gateway/api/topic/:topicID/publisher", h.getPublishers)
}

func (h HTTPAPI) createTopic(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	t := &Topic{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(t)

	output, err := h.PubSub.CreateTopic(t)
	if err != nil {
		if _, ok := err.(*ErrorAlreadyExists); ok {
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

func (h HTTPAPI) deleteTopic(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	err := h.PubSub.DeleteTopic(TopicID(params.ByName("topicID")))
	if err != nil {
		if _, ok := err.(*ErrorNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Error{Error: err.Error()})
	}
}

func (h HTTPAPI) getTopics(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	tps, err := h.PubSub.GetAllTopics()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(&topics{tps})
	}
}

func (h HTTPAPI) createSubscription(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	s := &Subscription{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(s)

	output, err := h.PubSub.CreateSubscription(TopicID(params.ByName("topicID")), s)
	if err != nil {
		if _, ok := err.(*ErrorSubscriptionAlreadyExists); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorNotFound); ok {
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

	subs, err := h.PubSub.GetAllSubscriptions(TopicID(params.ByName("topicID")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(&subscriptions{subs})
	}
}

func (h HTTPAPI) createPublisher(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	p := &Publisher{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(p)

	output, err := h.PubSub.CreatePublisher(TopicID(params.ByName("topicID")), p)
	if err != nil {
		if _, ok := err.(*ErrorPublisherAlreadyExists); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorNotFound); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorFunctionNotFound); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else if _, ok := err.(*ErrorPublisherValidation); ok {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(output)
	}
}

func (h HTTPAPI) deletePublisher(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	err := h.PubSub.DeletePublisher(PublisherID(params.ByName("publisherID")))
	if err != nil {
		if _, ok := err.(*ErrorSubscriptionNotFound); ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	}
}

func (h HTTPAPI) getPublishers(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	pubs, err := h.PubSub.GetAllPublishers(TopicID(params.ByName("topicID")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(&publishers{pubs})
	}
}

type topics struct {
	Topics []*Topic `json:"topics"`
}

type subscriptions struct {
	Subscriptions []*Subscription `json:"subscriptions"`
}

type publishers struct {
	Publishers []*Publisher `json:"publishers"`
}
