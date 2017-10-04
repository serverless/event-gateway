package router

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	eventpkg "github.com/serverless/event-gateway/event"
)

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

func fromRequest(r *http.Request) (*eventpkg.Event, error) {
	eventType := eventpkg.Type(r.Header.Get("event"))
	if eventType == "" {
		eventType = eventpkg.TypeHTTP
	}

	mime := r.Header.Get("content-type")
	if mime == "" {
		mime = mimeOctetStrem
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	event := eventpkg.NewEvent(eventType, mime, body)

	if mime == mimeJSON && len(body) > 0 {
		err := json.Unmarshal(body, &event.Data)
		if err != nil {
			return nil, err
		}
	}

	if event.Type == eventpkg.TypeHTTP {
		event.Data = &eventpkg.HTTPEvent{
			Headers: r.Header,
			Query:   r.URL.Query(),
			Body:    event.Data,
			Path:    r.URL.Path,
			Method:  r.Method,
		}
	}

	return event, nil
}

type workEvent struct {
	path  string
	event eventpkg.Event
}
