package router

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"
	"github.com/serverless/event-gateway/pubsub"
)

// Schema is a default event schema. All data that passes through the Event Gateway is formatted as an Event, based on this schema.
type Schema struct {
	Event      string      `json:"event"`
	ID         string      `json:"id"`
	ReceivedAt uint        `json:"receivedAt"`
	Encoding   string      `json:"encoding"`
	Data       interface{} `json:"data"`
}

const (
	encodingJSON   = "json"
	encodingBinary = "binary"
)

func transform(event, encoding string, payload []byte) ([]byte, error) {
	if encoding == "" {
		encoding = encodingBinary
	}

	instance := &Schema{
		Event:      event,
		ID:         uuid.NewV4().String(),
		ReceivedAt: uint(time.Now().UnixNano() / int64(time.Millisecond)),
		Encoding:   encoding,
		Data:       payload,
	}

	if encoding == encodingJSON {
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
