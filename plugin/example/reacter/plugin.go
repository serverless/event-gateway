package main

import (
	"encoding/gob"
	"log"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/plugin"
	"github.com/serverless/event-gateway/subscription"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(event.SystemEventReceivedData{})
	gob.Register(event.SystemFunctionInvokingData{})
}

// Simple plugin demonstrating how to build plugins.
type Simple struct{}

// Subscriptions return list of events that plugin listens to.
func (s *Simple) Subscriptions() []plugin.Subscription {
	return []plugin.Subscription{
		plugin.Subscription{
			EventType: event.SystemEventReceivedType,
			Type:      subscription.TypeSync,
		},
	}
}

// React is called for every event that plugin subscribed to.
func (s *Simple) React(instance event.Event) error {
	switch instance.EventType {
	case event.SystemEventReceivedType:
		received := instance.Data.(event.SystemEventReceivedData)
		log.Printf("received %s for event: %q", event.SystemEventReceivedType, received.Event.EventType)
		break
	}

	return nil
}
