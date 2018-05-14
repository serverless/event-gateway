package event_test

import (
	"testing"

	"io/ioutil"
	"net/http"
	"strings"

	"net/url"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/stretchr/testify/assert"
)

func TestFromRequest(t *testing.T) {
	for _, testCase := range fromRequestTests {
		u, err := url.Parse(testCase.url)
		assert.Nil(t, err)

		h := http.Header{}
		h.Add("Content-Type", testCase.contentType)

		e, err := eventpkg.FromRequest(&http.Request{
			URL:    u,
			Header: h,
			Body:   ioutil.NopCloser(strings.NewReader(testCase.body)),
		})

		assert.Nil(t, err)
		assert.Equal(t, testCase.expectedEvent.EventType, e.EventType)
		assert.Equal(t, testCase.expectedEvent.Source, e.Source)
		assert.Equal(t, testCase.expectedEvent.ContentType, e.ContentType)
		assert.Equal(t, testCase.expectedEvent.Data, e.Data)
	}
}

var fromRequestTests = []struct {
	url           string
	contentType   string
	body          string
	expectedEvent *eventpkg.Event
}{
	{
		url:         "https://something.eventgateway.com/myspace",
		contentType: "application/cloudevents+json",
		body: `{
			"eventType": "user.created",
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"
			}`,
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.Type("user.created"),
			CloudEventsVersion: eventpkg.TransformationVersion,
			Source:             "/mysource",
			ContentType:        "text/plain",
			Data: &eventpkg.HTTPEvent{
				Headers: map[string]string{"Content-Type": "application/cloudevents+json"},
				Query:   map[string][]string{},
				Body:    "test",
				Host:    "",
				Path:    "/myspace",
				Method:  "",
				Params:  map[string]string(nil),
			},
		},
	},
	{
		url:         "https://something.eventgateway.com/myspace",
		contentType: "",
		body: `{
			"eventType": "user.created",
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"}`,
		expectedEvent: &eventpkg.Event{
			EventType:          eventpkg.TypeHTTP,
			CloudEventsVersion: eventpkg.TransformationVersion,
			Source:             "https://serverless.com/event-gateway/#transformationVersion=0.1",
			ContentType:        "application/octet-stream",
			Data: &eventpkg.HTTPEvent{
				Headers: map[string]string{"Content-Type": ""},
				Query:   map[string][]string{},
				Body: []byte(`{
			"eventType": "user.created",
			"cloudEventsVersion": "` + eventpkg.TransformationVersion + `",
			"source": "/mysource",
			"eventID": "6f6ada3b-0aa2-4b3c-989a-91ffc6405f11",
			"contentType": "text/plain",
			"data": "test"}`),
				Host:   "",
				Path:   "/myspace",
				Method: "",
				Params: map[string]string(nil),
			},
		},
	},
}
