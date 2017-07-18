package router

import (
	"github.com/serverless/event-gateway/pubsub"
)

type event struct {
	topics  []pubsub.TopicID
	payload []byte
}
