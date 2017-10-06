package main

import (
	"encoding/gob"
	"log"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/plugin"
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
			Type:      plugin.Sync,
		},
	}
}

// React is called for every event that plugin subscribed to.
func (s *Simple) React(instance event.Event) error {
	switch instance.Type {
	case event.SystemEventReceivedType:
		received := instance.Data.(event.SystemEventReceivedData)
		log.Printf("received gateway.received.event for event: %q", received.Event.Type)
		break
	}

	return nil
}
