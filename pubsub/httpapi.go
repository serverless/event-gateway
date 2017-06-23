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
	router.DELETE("/v0/gateway/api/topic/:id", h.deleteTopic)
	router.GET("/v0/gateway/api/topic", h.getTopics)
}

func (h HTTPAPI) createTopic(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	t := &Topic{}
	dec := json.NewDecoder(r.Body)
	dec.Decode(t)

	output, err := h.PubSub.Create(t)
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

	err := h.PubSub.Delete(TopicID(params.ByName("id")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	}
}

func (h HTTPAPI) getTopics(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	tps, err := h.PubSub.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(&httpapi.Error{Error: err.Error()})
	} else {
		encoder.Encode(&topics{tps})
	}
}

type topics struct {
	Topics []*Topic `json:"topics"`
}
