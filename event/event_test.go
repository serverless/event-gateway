package event_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
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
		t.Run(testCase.name, func(t *testing.T) {
			url, _ := url.Parse("http://example.com")
			received, err := eventpkg.FromRequest(&http.Request{
				URL:    url,
				Header: testCase.requestHeaders,
				Body:   ioutil.NopCloser(bytes.NewReader(testCase.requestBody)),
			})

			if err != nil {
				assert.Equal(t, testCase.expectedError, err)
			} else {
				assert.Equal(t, testCase.expectedEvent.EventType, received.EventType, "EventType is not equal")
				assert.Equal(t, testCase.expectedEvent.Source, received.Source, "Source is not equal")
				assert.Equal(t, testCase.expectedEvent.CloudEventsVersion, received.CloudEventsVersion, "CloudEventsVersion is not equal")
				assert.Equal(t, testCase.expectedEvent.ContentType, received.ContentType, "ContentType is not equal")
				assert.Equal(t, testCase.expectedEvent.SchemaURL, received.SchemaURL, "SchemaURL is not equal")
				assert.Equal(t, testCase.expectedEvent.Extensions, received.Extensions, "Extensions is not equal")
				assert.EqualValues(t, testCase.expectedEvent.Data, received.Data, "Data is not equal")
			}
		})
	}
}

var newTests = []struct {
	name          string
	eventType     eventpkg.Type
	mime          string
	payload       interface{}
	expectedEvent eventpkg.Event
}{
	{
		name:      "not CloudEvent",
		eventType: eventpkg.Type("user.created"),
		mime:      "application/json",
		payload:   []byte("test"),
		expectedEvent: eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               []byte("test"),
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{
					"transformed":            true,
					"transformation-version": eventpkg.TransformationVersion,
				},
			},
		},
	},
	{
		name:      "system event",
		eventType: eventpkg.Type("user.created"),
		mime:      "application/json",
		payload:   eventpkg.SystemEventReceivedData{},
		expectedEvent: eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               eventpkg.SystemEventReceivedData{},
			Extensions: map[string]interface{}{
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
	name           string
	requestHeaders http.Header
	requestBody    []byte
	expectedEvent  *eventpkg.Event
	expectedError  error
}{
	{
		name:           "valid CloudEvent",
		requestHeaders: http.Header{"Content-Type": []string{"application/cloudevents+json"}},
		requestBody: []byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "0.1",
			"source": "http://example.com",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "http://example.com",
			ContentType:        "text/plain",
			Data:               "test",
		},
	},
	// {
	// 	name:           "error when invalid CloudEvent",
	// 	requestHeaders: http.Header{"Content-Type": []string{"application/cloudevents+json"}},
	// 	requestBody:    []byte(`{"eventType": "invalid"}`),
	// 	expectedEvent: &eventpkg.Event{
	// 		EventType:          eventpkg.TypeHTTPRequest,
	// 		CloudEventsVersion: "0.1",
	// 		Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
	// 		ContentType:        "application/cloudevents+json",
	// 		Data: &eventpkg.HTTPRequestData{
	// 			Headers: map[string]string{"Content-Type": "application/cloudevents+json"},
	// 			Body:    []byte(`{"eventType": "invalid"}`),
	// 			Query:   map[string][]string{}},
	// 		Extensions: map[string]interface{}{
	// 			"eventgateway": map[string]interface{}{"transformed": true, "transformation-version": "0.1"}},
	// 	},
	// },
	{
		name: "valid CloudEvent in binary mode",
		requestHeaders: http.Header{
			"Content-Type":          []string{"text/plain"},
			"Ce-Eventtype":          []string{"myevent"},
			"Ce-Eventtypeversion":   []string{"0.1beta"},
			"Ce-Cloudeventsversion": []string{"0.1"},
			"Ce-Source":             []string{"https://example.com"},
			"Ce-Eventid":            []string{"778d495b-a29e-48f9-a438-a26de1e33515"},
			"Ce-Schemaurl":          []string{"https://example.com"},
			"Ce-X-MyExtension":      []string{"ding"},
		},
		requestBody: []byte("hey there"),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("myevent"),
			CloudEventsVersion: "0.1",
			Source:             "https://example.com",
			ContentType:        "text/plain",
			SchemaURL:          "https://example.com",
			Data:               []byte("hey there"),
			Extensions:         map[string]interface{}{"myExtension": "ding"},
		},
	},
	{
		name: "legacy mode event",
		requestHeaders: http.Header{
			"Event": []string{"myevent"},
		},
		requestBody: []byte("hey there"),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("myevent"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/octet-stream",
			Data:               []byte("hey there"),
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{"transformed": true, "transformation-version": "0.1"}},
		},
	},
	{
		name: "valid CloudEvent in legacy mode",
		requestHeaders: http.Header{
			"Content-Type": []string{"application/json"},
			"Event":        []string{"user.created"},
		},
		requestBody: []byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "0.1",
			"source": "http://example.com",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "http://example.com",
			ContentType:        "text/plain",
			Data:               "test",
		},
	},
	{
		name: "invalid CloudEvent in legacy mode",
		requestHeaders: http.Header{
			"Content-Type": []string{"application/json"},
			"Event":        []string{"user.created"},
		},
		requestBody: []byte(`{
			"eventType": "user.created"
		}`),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               map[string]interface{}{"eventType": "user.created"},
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{"transformed": true, "transformation-version": "0.1"}},
		},
	},
	{
		name: "regular HTTP request",
		requestHeaders: http.Header{
			"Content-Type": []string{"application/json"},
		},
		requestBody: []byte("hey there"),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.TypeHTTPRequest,
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/cloudevents+json",
			Data: &eventpkg.HTTPRequestData{
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Query: map[string][]string{},
				Body:  []byte("hey there"),
			},
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{"transformed": true, "transformation-version": "0.1"}},
		},
	},
}
