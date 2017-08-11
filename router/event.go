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
	ReceivedAt uint64      `json:"receivedAt"`
	Data       interface{} `json:"data"`
	DataType   string      `json:"dataType"`
}

// HTTPSchema is a event schema used for sending events to HTTP subscriptions.
type HTTPSchema struct {
	Headers map[string][]string `json:"headers"`
	Query   map[string][]string `json:"query"`
	Body    interface{}         `json:"body"`
}

const (
	mimeJSON       = "application/json"
	mimeOctetStrem = "application/octet-stream"
)

func fromRequest(r *http.Request) (*Schema, error) {
	name := r.Header.Get("event")
	if name == "" {
		name = eventHTTP
	}

	mime := r.Header.Get("content-type")
	if mime == "" {
		mime = mimeOctetStrem
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	event := &Schema{
		Event:      name,
		ID:         uuid.NewV4().String(),
		ReceivedAt: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		DataType:   mime,
		Data:       body,
	}

	if mime == mimeJSON && len(body) > 0 {
		err := json.Unmarshal(body, &event.Data)
		if err != nil {
			return nil, err
		}
	}

	if event.Event == eventHTTP {
		event.Data = &HTTPSchema{
			Headers: r.Header,
			Query:   r.URL.Query(),
			Body:    event.Data,
		}
	}

	return event, nil
}

type event struct {
	topics  []pubsub.TopicID
	payload []byte
}
