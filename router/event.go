package router

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/satori/go.uuid"
	"github.com/serverless/event-gateway/subscriptions"
)

// Event is a default event structure. All data that passes through the Event Gateway is formatted as an Event, based on this schema.
type Event struct {
	Event      string      `json:"event"`
	ID         string      `json:"id"`
	ReceivedAt uint64      `json:"receivedAt"`
	Data       interface{} `json:"data"`
	DataType   string      `json:"dataType"`
}

// NewEvent return new instance of Event.
func NewEvent(name, mime string, payload interface{}) *Event {
	return &Event{
		Event:      name,
		ID:         uuid.NewV4().String(),
		ReceivedAt: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		DataType:   mime,
		Data:       payload,
	}
}

// HTTPEvent is a event schema used for sending events to HTTP subscriptions.
type HTTPEvent struct {
	Headers map[string][]string `json:"headers"`
	Query   map[string][]string `json:"query"`
	Body    interface{}         `json:"body"`
	Path    string              `json:"path"`
	Method  string              `json:"method"`
}

// HTTPResponse is a response schema returned by subscribed function in case of HTTP event.
type HTTPResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

const (
	mimeJSON       = "application/json"
	mimeOctetStrem = "application/octet-stream"
)

func fromRequest(r *http.Request) (*Event, error) {
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

	event := NewEvent(name, mime, body)

	if mime == mimeJSON && len(body) > 0 {
		err := json.Unmarshal(body, &event.Data)
		if err != nil {
			return nil, err
		}
	}

	if event.Event == eventHTTP {
		event.Data = &HTTPEvent{
			Headers: r.Header,
			Query:   r.URL.Query(),
			Body:    event.Data,
			Path:    r.URL.Path,
			Method:  r.Method,
		}
	}

	return event, nil
}

type event struct {
	topic   subscriptions.TopicID
	payload []byte
}
