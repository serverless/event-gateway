package event_test

import (
	"testing"
	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/stretchr/testify/assert"
	"encoding/json"
)

var exampleEvent = map[string]interface{}{
	"event-type": "user.created",
	"event-type-version": "",
	"cloud-events-version": "0.1",
	"source": "https://slsgateway.com/",
	"event-id": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
	"event-time": "2018-04-03T18:48:16.537683692+04:00",
	"schema-url": "",
	"content-type": "application/x-www-form-urlencoded",
	"extensions": nil,
	"data": map[string]interface{}{
		"headers": map[string]interface{}{
			"Accept": "*/*",
			"Content-Length": "11",
			"Content-Type": "application/x-www-form-urlencoded",
			"User-Agent": "insomnia/5.14.9",
		},
		"query": map[string]interface{}{},
		"body": "file=adsasd",
		"host": "localhost:4000",
		"path": "/something/users/1",
		"method": "POST",
		"params": map[string]interface{}{
			"id": "1",
		},
	},
}

func TestNew_CustomToCloudEvent(t *testing.T) {
	testevent, err := json.Marshal(exampleEvent)
	assert.Nil(t, err)
	event := eventpkg.New(eventpkg.Type(exampleEvent["event-type"].(string)), exampleEvent["content-type"].(string), testevent)
	assert.NotNil(t, event)
	assert.NotNil(t, event.Data)
}

// Will not parse payload to CloudEvent because payload event-type is different from request event-type
func TestNew_CustomToCloudEventEventType(t *testing.T){
	exampleEvent["source"] = ""
	testevent, err := json.Marshal(exampleEvent)
	assert.Nil(t, err)
	event := eventpkg.New(eventpkg.Type("myevent"), exampleEvent["content-type"].(string), testevent)
	assert.NotEqual(t, exampleEvent["event-id"], event.EventID)
}

// Will not parse payload to CloudEvent because it will fail on "source" validation
func TestNew_CustomToCloudEventSourceValidation(t *testing.T){
	exampleEvent["source"] = ""
	testevent, err := json.Marshal(exampleEvent)
	assert.Nil(t, err)
	event := eventpkg.New(eventpkg.Type(exampleEvent["event-type"].(string)), exampleEvent["content-type"].(string), testevent)
	assert.NotEqual(t, exampleEvent["event-id"], event.EventID)
}
