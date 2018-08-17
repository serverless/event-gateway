package queue

import (
	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/subscription"
)

type Queue interface {
	Push(subscriptionID subscription.ID, event eventpkg.Event) error
	Pull() (*eventpkg.Event, error)
	MarkAsDelivered(subscriptionID subscription.ID, eventID string) error
}
