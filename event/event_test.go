package event_test

import (
	"testing"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/stretchr/testify/assert"
	"github.com/serverless/event-gateway/internal/zap"
)

const (
	transformationVersion string = "0.1"
)

func TestNew(t *testing.T) {
	for _, testCase := range newTests {
		result := eventpkg.New(testCase.eventType, "https://slsgateway.com", testCase.mime, testCase.payload)

		assert.NotEqual(t, result.EventID, "")
		assert.Equal(t, testCase.expectedEvent.EventType, result.EventType)
		assert.Equal(t, testCase.expectedEvent.CloudEventsVersion, result.CloudEventsVersion)
		assert.Equal(t, testCase.expectedEvent.Source, result.Source)
		assert.Equal(t, testCase.expectedEvent.ContentType, result.ContentType)
		assert.Equal(t, testCase.expectedEvent.Data, result.Data)
		assert.Equal(t, testCase.expectedEvent.Extensions, result.Extensions)
	}
}

var newTests = []struct {
	eventType     eventpkg.Type
	mime          string
	payload       interface{}
	expectedEvent eventpkg.Event
}{
	{ // not CloudEvent
		eventpkg.Type("user.created"),
		"application/json",
		[]byte("test"),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: transformationVersion,
			Source:             "https://slsgateway.com#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               []byte("test"),
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": transformationVersion,
				},
			},
		},
	},
	{ // System event
		eventpkg.Type("user.created"),
		"application/json",
		eventpkg.SystemEventReceivedData{},
		eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: transformationVersion,
			Source:             "https://slsgateway.com#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               eventpkg.SystemEventReceivedData{},
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": transformationVersion,
				},
			},
		},
	},
	{
		// valid CloudEvent
		eventpkg.Type("user.created"),
		"application/json",
		[]byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "`+ transformationVersion +`",
			"source": "https://example.com/",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: transformationVersion,
			Source:             "https://example.com/",
			ContentType:        "text/plain",
			Data:               "test",
		},
	},
	{
		// type mismatch
		eventpkg.Type("user.deleted"),
		"application/json",
		[]byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "`+ transformationVersion +`",
			"source": "https://example.com/",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.deleted"),
			CloudEventsVersion: transformationVersion,
			Source:             "https://slsgateway.com#transformationVersion=0.1",
			ContentType:        "application/json",
			Data: map[string]interface{}{
				"eventType":           "user.created",
				"cloudEventsVersion": transformationVersion,
				"source":               "https://example.com/",
				"eventID":             "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
				"contentType":         "text/plain",
				"data":                 "test",
			},
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": transformationVersion,
				},
			},
		},
	},
	{
		// invalid CloudEvent (missing required fields)
		eventpkg.Type("user.created"),
		"application/json",
		[]byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "0.1"
			}`),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: transformationVersion,
			Source:             "https://slsgateway.com#transformationVersion=0.1",
			ContentType:        "application/json",
			Data: map[string]interface{}{
				"eventType":           "user.created",
				"cloudEventsVersion": transformationVersion,
			},
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": transformationVersion,
				},
			},
		},
	},
}
