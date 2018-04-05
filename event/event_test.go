package event_test

import (
	"testing"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	for _, testCase := range newTests {
		result := eventpkg.New(testCase.eventType, testCase.mime, testCase.payload)

		assert.NotEqual(t, result.EventID, "")
		assert.Equal(t, testCase.expectedEvent.EventType, result.EventType)
		assert.Equal(t, testCase.expectedEvent.CloudEventsVersion, result.CloudEventsVersion)
		assert.Equal(t, testCase.expectedEvent.Source, result.Source)
		assert.Equal(t, testCase.expectedEvent.ContentType, result.ContentType)
		assert.Equal(t, testCase.expectedEvent.Data, result.Data)
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
			CloudEventsVersion: "0.1",
			Source:             "https://slsgateway.com/",
			ContentType:        "application/json",
			Data:               []byte("test"),
		},
	},
	{ // System event
		eventpkg.Type("user.created"),
		"application/json",
		eventpkg.SystemEventReceivedData{},
		eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "https://slsgateway.com/",
			ContentType:        "application/json",
			Data:               eventpkg.SystemEventReceivedData{},
		},
	},
	{
		// valid CloudEvent
		eventpkg.Type("user.created"),
		"application/json",
		[]byte(`{
			"event-type": "user.created",
			"cloud-events-version": "0.1",
			"source": "https://example.com/",
			"event-id": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"content-type": "text/plain",
			"data": "test"
			}`),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
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
			"event-type": "user.created",
			"cloud-events-version": "0.1",
			"source": "https://example.com/",
			"event-id": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"content-type": "text/plain",
			"data": "test"
			}`),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.deleted"),
			CloudEventsVersion: "0.1",
			Source:             "https://slsgateway.com/",
			ContentType:        "application/json",
			Data: map[string]interface{}{
				"event-type":           "user.created",
				"cloud-events-version": "0.1",
				"source":               "https://example.com/",
				"event-id":             "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
				"content-type":         "text/plain",
				"data":                 "test",
			},
		},
	},
	{
		// invalid CloudEvent (missing required fields)
		eventpkg.Type("user.created"),
		"application/json",
		[]byte(`{
			"event-type": "user.created",
			"cloud-events-version": "0.1"
			}`),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "https://slsgateway.com/",
			ContentType:        "application/json",
			Data: map[string]interface{}{
				"event-type":           "user.created",
				"cloud-events-version": "0.1",
			},
		},
	},
}
