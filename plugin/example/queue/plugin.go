package main

import (
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/subscription"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(event.SystemEventReceivedData{})
	gob.Register(event.SystemFunctionInvokingData{})
}

type Filesystem struct{}

func (f Filesystem) Push(subscriptionID subscription.ID, event event.Event) error {
	byt, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("event_"+event.EventID+".json", byt, 0644)
}

func (f Filesystem) Pull() (*event.Event, error) {
	log.Printf("PULL")
	return nil, nil
}

func (f Filesystem) MarkAsDelivered(subscriptionID subscription.ID, eventID string) error {
	log.Printf("MARK AS DELIVERED")
	return nil
}
