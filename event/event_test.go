package event_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/bouk/monkey"
	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	for _, testCase := range newTests {
		t.Run(testCase.name, func(t *testing.T) {
			result := eventpkg.New(testCase.eventType, testCase.mime, testCase.payload)

			assert.NotEqual(t, result.EventID, "")
			assert.Equal(t, testCase.expectedEvent.EventType, result.EventType)
			assert.Equal(t, testCase.expectedEvent.CloudEventsVersion, result.CloudEventsVersion)
			assert.Equal(t, testCase.expectedEvent.Source, result.Source)
			assert.Equal(t, testCase.expectedEvent.ContentType, result.ContentType)
			assert.Equal(t, testCase.expectedEvent.Data, result.Data)
			assert.Equal(t, testCase.expectedEvent.Extensions, result.Extensions)
		})
	}
}

func TestNew_Encoding(t *testing.T) {
	for _, testCase := range encodingTests {
		result := eventpkg.New(eventpkg.TypeName("test.event"), testCase.contentType, testCase.body)

		assert.Equal(t, testCase.expectedBody, result.Data)
	}
}

func TestFromRequest(t *testing.T) {
	patch := monkey.Patch(time.Now, func() time.Time { return testTime })
	defer patch.Unpatch()

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
				assert.Equal(t, testCase.expectedEvent.EventTime, received.EventTime, "EventTime is not equal")
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
	eventType     eventpkg.TypeName
	mime          string
	payload       interface{}
	expectedEvent eventpkg.Event
}{
	{
		name:      "not CloudEvent",
		eventType: eventpkg.TypeName("user.created"),
		mime:      "application/json",
		payload:   []byte("test"),
		expectedEvent: eventpkg.Event{
			EventType:          eventpkg.TypeName("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               []byte("test"),
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{
					"transformed":            "true",
					"transformation-version": eventpkg.TransformationVersion,
				},
			},
		},
	},
	{
		name:      "system event",
		eventType: eventpkg.TypeName("user.created"),
		mime:      "application/json",
		payload:   eventpkg.SystemEventReceivedData{},
		expectedEvent: eventpkg.Event{
			EventType:          eventpkg.TypeName("user.created"),
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               eventpkg.SystemEventReceivedData{},
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{
					"transformed":            "true",
					"transformation-version": eventpkg.TransformationVersion,
				},
			},
		},
	},
}

var testTime = time.Date(1985, time.April, 12, 23, 20, 50, 00, time.UTC) //1985-04-12T23:20:50.00Z

var encodingTests = []struct {
	name         string
	body         []byte
	contentType  string
	expectedBody interface{}
}{
	{
		"unrecogninzed content type",
		[]byte("some=thing"),
		"application/octet-stream",
		[]byte("some=thing"),
	},
	{
		"form content type",
		[]byte("some=thing"),
		"application/x-www-form-urlencoded",
		"some=thing",
	},
	{
		"multipart form content type",
		[]byte("--X-INSOMNIA-BOUNDARY\r\nContent-Disposition: form-data; name=\"some\"\r\n\r\nthing\r\n--X-INSOMNIA-BOUNDARY--\r\n"),
		"multipart/form-data; boundary=X-INSOMNIA-BOUNDARY",
		"--X-INSOMNIA-BOUNDARY\r\nContent-Disposition: form-data; name=\"some\"\r\n\r\nthing\r\n--X-INSOMNIA-BOUNDARY--\r\n",
	},
	{
		"JSON content type",
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
			"eventTime": "1985-04-12T23:20:50.00Z",
			"cloudEventsVersion": "0.1",
			"source": "http://example.com",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),

		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.TypeName("user.created"),
			EventTime:          &testTime,
			CloudEventsVersion: "0.1",
			Source:             "http://example.com",
			ContentType:        "text/plain",
			Data:               "test",
		},
	},
	{
		name:           "valid CloudEvent with invalid event time",
		requestHeaders: http.Header{"Content-Type": []string{"application/cloudevents+json"}},
		requestBody: []byte(`{
			"eventType": "user.created",
			"eventTime": "nottime",
			"cloudEventsVersion": "0.1",
			"source": "http://example.com",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`),
		expectedError: &time.ParseError{Layout: "\"2006-01-02T15:04:05Z07:00\"", Value: "\"nottime\"", LayoutElem: "2006", ValueElem: "nottime\"", Message: ""},
	},
	{
		name:           "error if invalid CloudEvent",
		requestHeaders: http.Header{"Content-Type": []string{"application/cloudevents+json"}},
		requestBody:    []byte(`{"eventType": "invalid"}`),
		expectedError: &eventpkg.ErrParsingCloudEvent{
			Message: "Key: 'Event.CloudEventsVersion' Error:Field validation for 'CloudEventsVersion' failed on the 'required' tag\n" +
				"Key: 'Event.Source' Error:Field validation for 'Source' failed on the 'uri' tag\n" +
				"Key: 'Event.EventID' Error:Field validation for 'EventID' failed on the 'required' tag"},
	},
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
			EventType:          eventpkg.TypeName("myevent"),
			CloudEventsVersion: "0.1",
			Source:             "https://example.com",
			ContentType:        "text/plain",
			SchemaURL:          "https://example.com",
			Data:               []byte("hey there"),
			Extensions:         map[string]interface{}{"myExtension": "ding"},
		},
	},
	{
		name: "valid CloudEvent in binary mode with valid event time",
		requestHeaders: http.Header{
			"Content-Type":          []string{"text/plain"},
			"Ce-Eventtype":          []string{"myevent"},
			"Ce-Eventtypeversion":   []string{"0.1beta"},
			"Ce-Cloudeventsversion": []string{"0.1"},
			"Ce-Source":             []string{"https://example.com"},
			"Ce-Eventid":            []string{"778d495b-a29e-48f9-a438-a26de1e33515"},
			"Ce-Eventtime":          []string{"1985-04-12T23:20:50.00Z"},
			"Ce-Schemaurl":          []string{"https://example.com"},
			"Ce-X-MyExtension":      []string{"ding"},
		},
		requestBody: []byte("hey there"),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.TypeName("myevent"),
			EventTime:          &testTime,
			CloudEventsVersion: "0.1",
			Source:             "https://example.com",
			ContentType:        "text/plain",
			SchemaURL:          "https://example.com",
			Data:               []byte("hey there"),
			Extensions:         map[string]interface{}{"myExtension": "ding"},
		},
	},
	{
		name: "valid CloudEvent in binary mode with invalid event time",
		requestHeaders: http.Header{
			"Content-Type":          []string{"text/plain"},
			"Ce-Eventtype":          []string{"myevent"},
			"Ce-Eventtypeversion":   []string{"0.1beta"},
			"Ce-Cloudeventsversion": []string{"0.1"},
			"Ce-Source":             []string{"https://example.com"},
			"Ce-Eventid":            []string{"778d495b-a29e-48f9-a438-a26de1e33515"},
			"Ce-Eventtime":          []string{"nottime"},
			"Ce-Schemaurl":          []string{"https://example.com"},
			"Ce-X-MyExtension":      []string{"ding"},
		},
		requestBody:   []byte("hey there"),
		expectedError: &time.ParseError{Layout: "2006-01-02T15:04:05Z07:00", Value: "nottime", LayoutElem: "2006", ValueElem: "nottime", Message: ""},
	},
	{
		name: "error if invalid CloudEvent in binary mode",
		requestHeaders: http.Header{
			"Content-Type":          []string{"text/plain"},
			"Ce-Eventtype":          []string{"myevent"},
			"Ce-Cloudeventsversion": []string{"0.1"},
			"Ce-Source":             []string{"NOT URI"},
			"Ce-Eventid":            []string{"778d495b-a29e-48f9-a438-a26de1e33515"},
		},
		requestBody: []byte(`{"eventType": "invalid"}`),
		expectedError: &eventpkg.ErrParsingCloudEvent{
			Message: "Key: 'Event.Source' Error:Field validation for 'Source' failed on the 'uri' tag"},
	},
	{
		name: "legacy mode event",
		requestHeaders: http.Header{
			"Event": []string{"myevent"},
		},
		requestBody: []byte("hey there"),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.TypeName("myevent"),
			EventTime:          &testTime,
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/octet-stream",
			Data:               []byte("hey there"),
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{"transformed": "true", "transformation-version": "0.1"}},
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
			EventType:          eventpkg.TypeName("user.created"),
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
			EventType:          eventpkg.TypeName("user.created"),
			EventTime:          &testTime,
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data:               map[string]interface{}{"eventType": "user.created"},
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{"transformed": "true", "transformation-version": "0.1"}},
		},
	},
	{
		name: "regular HTTP JSON request",
		requestHeaders: http.Header{
			"Content-Type": []string{"application/json"},
		},
		requestBody: []byte(`{"key": "value"}`),
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.TypeHTTPRequest,
			EventTime:          &testTime,
			CloudEventsVersion: "0.1",
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/json",
			Data: &eventpkg.HTTPRequestData{
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Query: map[string][]string{},
				Body: map[string]interface{}{
					"key": "value",
				},
			},
			Extensions: map[string]interface{}{
				"eventgateway": map[string]interface{}{"transformed": "true", "transformation-version": "0.1"}},
		},
	},
}
