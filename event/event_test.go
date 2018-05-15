package event_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/internal/zap"
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
		assert.Equal(t, testCase.expectedEvent.Extensions, result.Extensions)
	}
}

func TestNew_Encoding(t *testing.T) {
	for _, testCase := range encodingTests {
		result := eventpkg.New(eventpkg.Type("test.event"), testCase.contentType, testCase.body)

		assert.Equal(t, testCase.expectedBody, result.Data)
	}
}

func TestFromRequest(t *testing.T) {
	for _, testCase := range fromRequestTests {
		u, err := url.Parse(testCase.url)
		assert.Nil(t, err)

		h := http.Header{}
		h.Add("Content-Type", testCase.contentType)

		for key, val := range testCase.headers {
			h.Add(key, val)
		}

		e, err := eventpkg.FromRequest(&http.Request{
			URL:    u,
			Header: h,
			Body:   ioutil.NopCloser(bytes.NewReader(testCase.body)),
		})

		assert.Nil(t, err, fmt.Sprintf("event.FromRequest threw an error: %v", err))
		assert.Equal(t, testCase.expectedEvent.EventType, e.EventType, "EventType is not equal")
		assert.Equal(t, testCase.expectedEvent.Source, e.Source, "Source is not equal")
		assert.Equal(t, testCase.expectedEvent.CloudEventsVersion, e.CloudEventsVersion, "CloudEventsVersion is not equal")
		assert.Equal(t, testCase.expectedEvent.ContentType, e.ContentType, "ContentType is not equal")
		if expectedHRD, ok := testCase.expectedEvent.Data.(*eventpkg.HTTPRequestData); ok {
			actualHRD, ok := e.Data.(*eventpkg.HTTPRequestData)
			assert.True(t, ok, "actual Event.Data is not HTTPRequestData")
			for key, _ := range expectedHRD.Headers {
				assert.Equal(t, expectedHRD.Headers[key], actualHRD.Headers[key], fmt.Sprintf("Expected header %s is not equal", key))
			}
			assert.Equal(t, expectedHRD.Body, actualHRD.Body, "Body is not equal")
		} else {
			assert.Equal(t, testCase.expectedEvent.Data, e.Data, "Event.Data is not equal")
		}
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
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               []byte("test"),
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed":            true,
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
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               eventpkg.SystemEventReceivedData{},
			Extensions: zap.MapStringInterface{
				"eventgateway": map[string]interface{}{
					"transformed":            true,
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

var fromRequestTests = []struct {
	url           string
	contentType   string
	headers       map[string]string
	body          []byte
	expectedEvent *eventpkg.Event
}{
	// Valid CloudEvent with application/json content-type
	{
		url:         "https://example.com/myspace",
		contentType: "application/json; charset=utf-8",
		body: []byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "/mysource",
			ContentType:        "text/plain",
			Data: &eventpkg.HTTPRequestData{
				Headers: map[string]string{"content-type": "application/json; charset=utf-8"},
				Body:    "test",
			},
		},
	},
	// Valid CloudEvent with application/cloudevents+json content-type
	{
		url:         "https://example.com/myspace",
		contentType: "application/cloudevents+json",
		body: []byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "/mysource",
			ContentType:        "text/plain",
			Data: &eventpkg.HTTPRequestData{
				Headers: map[string]string{"content-type": "application/cloudevents+json"},
				Body:    "test",
			},
		},
	},
	// invalid CloudEvent
	{
		url:         "https://example.com/myspace",
		contentType: "application/cloudevents+json",
		body: []byte(`{
			"eventType": 
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("http"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/cloudevents+json",
			Data: &eventpkg.HTTPRequestData{
				Headers: map[string]string{"content-type": "application/cloudevents+json"},
				Body: []byte(`{
			"eventType": 
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
			},
		},
	},
	// Empty content type
	{
		url:         "https://example.com/myspace",
		contentType: "",
		body: []byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"}`),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.TypeHTTP,
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/octet-stream",
			Data: &eventpkg.HTTPRequestData{
				Headers: map[string]string{"content-type": ""},
				Body: []byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"}`),
			},
		},
	},
	// CloudEvent from headers
	{
		url: "https://example.com/myspace",
		headers: map[string]string{
			"CE-EventType":          "myevent",
			"CE-EventTypeVersion":   "0.1beta",
			"CE-CloudEventsVersion": "0.1",
			"CE-Source":             "https://example.com",
			"CE-EventID":            "778d495b-a29e-48f9-a438-a26de1e33515",
		},
		body: []byte("hey there"),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("myevent"),
			CloudEventsVersion: "0.1",
			Source:             "https://example.com",
			Data: &eventpkg.HTTPRequestData{
				Headers: map[string]string{
					"Ce-Eventid":            "778d495b-a29e-48f9-a438-a26de1e33515",
					"Ce-Eventtype":          "myevent",
					"Ce-Eventtypeversion":   "0.1beta",
					"Ce-Cloudeventsversion": "0.1",
					"Ce-Source":             "https://example.com",
				},
				Body: []byte("hey there"),
			},
		},
	},
	// Custom event
	{
		headers: map[string]string{
			"Event": "myevent",
		},
		body: []byte("hey there"),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("myevent"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/octet-stream",
			Data:               []byte("hey there"),
		},
	},
	// Valid custom CloudEvent with application/cloudevents+json content-type
	{
		contentType: "application/cloudevents+json",
		body: []byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		headers: map[string]string{
			"Event": "myevent",
		},
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "/mysource",
			ContentType:        "text/plain",
			Data:               "test",
		},
	},
	// invalid custom CloudEvent with application/cloudevents+json content-type
	{
		contentType: "application/cloudevents+json",
		body: []byte(`{
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		headers: map[string]string{
			"Event": "myevent",
		},
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("myevent"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/cloudevents+json",
			Data: map[string]interface{}{
				"cloudEventsVersion": "0.1",
				"source":             "/mysource",
				"eventID":            "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
				"contentType":        "text/plain",
				"data":               "test",
			},
		},
	},
}
