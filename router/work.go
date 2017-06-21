package router

import (
	"github.com/serverless/event-gateway/pubsub"
)

type work struct {
	topic   pubsub.TopicID
	payload []byte
}
