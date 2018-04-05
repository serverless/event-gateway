package event_test

import (
	"testing"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/stretchr/testify/assert"
	"github.com/serverless/event-gateway/internal/zap"
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
		assert.Equal(t, testCase.expectedEvent.Extensions, result.Extensions)
	}
}

func TestNew_Encoding(t *testing.T) {
	for _, testCase := range encodingTests {
		result := eventpkg.New(eventpkg.Type("test.event"), testCase.contentType, testCase.body)

		assert.Equal(t, testCase.expectedBody, result.Data)
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
			CloudEventsVersion: eventpkg.TransformationVersion,
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               []byte("test"),
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": eventpkg.TransformationVersion,
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
			CloudEventsVersion: eventpkg.TransformationVersion,
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               eventpkg.SystemEventReceivedData{},
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": eventpkg.TransformationVersion,
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
			"cloudEventsVersion": "`+ eventpkg.TransformationVersion +`",
			"source": "https://example.com/",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: eventpkg.TransformationVersion,
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
			"cloudEventsVersion": "`+ eventpkg.TransformationVersion +`",
			"source": "https://example.com/",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		eventpkg.Event{
			EventType:          eventpkg.Type("user.deleted"),
			CloudEventsVersion: eventpkg.TransformationVersion,
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data: map[string]interface{}{
				"eventType":           "user.created",
				"cloudEventsVersion": eventpkg.TransformationVersion,
				"source":               "https://example.com/",
				"eventID":             "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
				"contentType":         "text/plain",
				"data":                 "test",
			},
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": eventpkg.TransformationVersion,
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
			CloudEventsVersion: eventpkg.TransformationVersion,
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data: map[string]interface{}{
				"eventType":           "user.created",
				"cloudEventsVersion": eventpkg.TransformationVersion,
			},
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed": true,
					"transformation-version": eventpkg.TransformationVersion,
				},
			},
		},
	},
}

var encodingTests = []struct {
	body         []byte
	contentType  string
	expectedBody interface{}
}{
	{
		[]byte("some=thing"),
		"application/octet-stream",
		[]byte("some=thing"),
	},
	{
		[]byte("some=thing"),
		"application/x-www-form-urlencoded",
		"some=thing",
	},
	{
		[]byte("--X-INSOMNIA-BOUNDARY\r\nContent-Disposition: form-data; name=\"some\"\r\n\r\nthing\r\n--X-INSOMNIA-BOUNDARY--\r\n"),
		"multipart/form-data; boundary=X-INSOMNIA-BOUNDARY",
		"--X-INSOMNIA-BOUNDARY\r\nContent-Disposition: form-data; name=\"some\"\r\n\r\nthing\r\n--X-INSOMNIA-BOUNDARY--\r\n",
	},
	{
		[]byte(`{"hello": "world"}`),
		"application/json",
		map[string]interface{}{"hello": "world"},
	},
}
