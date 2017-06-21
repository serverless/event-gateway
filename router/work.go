package router

import (
	"github.com/serverless/event-gateway/pubsub"
)

type work struct {
	topics  []pubsub.TopicID
	payload []byte
}
