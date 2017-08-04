package router

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/satori/go.uuid"
	"github.com/serverless/event-gateway/pubsub"
)

// Schema is a default event schema. All data that passes through the Event Gateway is formatted as an Event, based on this schema.
type Schema struct {
	Event      string      `json:"event"`
	ID         string      `json:"id"`
	ReceivedAt uint        `json:"receivedAt"`
	Data       interface{} `json:"data"`
	DataType   string      `json:"dataType"`
}

// HTTPSchema is a event schema used for sending events to HTTP subscriptions.
type HTTPSchema struct {
	Headers map[string][]string `json:"headers"`
	Query   map[string][]string `json:"query"`
	Data    interface{}         `json:"data"`
}

const (
	mimeJSON       = "application/json"
	mimeOctetStrem = "application/octet-stream"
)

func transform(event string, r *http.Request) ([]byte, error) {
	mime := r.Header.Get("content-type")
	if mime == "" {
		mime = mimeOctetStrem
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	instance := &Schema{
		Event:      event,
		ID:         uuid.NewV4().String(),
		ReceivedAt: uint(time.Now().UnixNano() / int64(time.Millisecond)),
		DataType:   mime,
		Data:       payload,
	}

	if mime == mimeJSON && len(payload) > 0 {
		err := json.Unmarshal(payload, &instance.Data)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(instance)
}

func transformHTTP(r *http.Request) ([]byte, error) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	instance := &HTTPSchema{
		Headers: r.Header,
		Query:   r.URL.Query(),
		Data:    payload,
	}

	if r.Header.Get("content-type") == mimeJSON && len(payload) > 0 {
		err := json.Unmarshal(payload, &instance.Data)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(instance)
}

type event struct {
	topics  []pubsub.TopicID
	payload []byte
}
