package cache

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/libkv"
	"github.com/serverless/event-gateway/subscription"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestSubscriptionCacheModifiedEvents(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc1",
		"path": "/"}`))
	scache.Modified("testsub2", []byte(`{
		"subscriptionId":
		"testsub2",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc2",
		"path": "/"}`))

	expected := []libkv.FunctionKey{
		libkv.FunctionKey{Space: "space1", ID: "testfunc1"},
		libkv.FunctionKey{Space: "space1", ID: "testfunc2"}}
	assert.Equal(t, expected, scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedSyncSubscription(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"type":"sync",
		"space": "default",
		"event": "http.request",
		"functionId": "testfunc1",
		"path": "/a",
		"method": "GET"}`))
	scache.Modified("testsub2", []byte(`{
		"subscriptionId":"testsub2",
		"type":"sync",
		"space": "default",
		"event": "http.request",
		"functionId": "testfunc2",
		"path": "/b",
		"method": "GET"}`))

	space, id, _, _ := scache.endpoints["GET"].Resolve("/a")
	assert.Equal(t, function.ID("testfunc1"), *id)
	assert.Equal(t, "default", space)
}

func TestSubscriptionCacheModifiedCORSConfiguration(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"type": "sync",
		"eventType": "http.request",
		"functionId": "testfunc1",
		"path": "/a",
		"method": "GET",
		"cors": {
			"origins": ["http://example.com"]
		}}`))

	_, _, _, corsConfig := scache.endpoints["GET"].Resolve("/a")
	assert.Equal(t, &subscription.CORS{Origins: []string{"http://example.com"}}, corsConfig)
}

func TestSubscriptionCacheModifiedEventsWrongPayload(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub", []byte(`not json`))

	assert.Equal(t, []libkv.FunctionKey(nil), scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedEventsDeleted(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc1",
		"path": "/"}`))
	scache.Modified("testsub2", []byte(`{
		"subscriptionId":"testsub2",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc2",
		"path": "/"}`))
	scache.Deleted("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc1",
		"path": "/"}`))

	assert.Equal(t, []libkv.FunctionKey{{Space: "space1", ID: function.ID("testfunc2")}}, scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedHTTPSubscriptionDeleted(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"type": "sync",
		"eventType": "http.request",
		"functionId": "testfunc1",
		"path": "/",
		"method": "GET"}`))
	scache.Deleted("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"type": "sync",
		"eventType": "http.request",
		"functionId": "testfunc1",
		"path": "/",
		"method": "GET"}`))

	space, id, _, _ := scache.endpoints["GET"].Resolve("/")
	assert.Nil(t, id)
	assert.Equal(t, "", space)
}

func TestSubscriptionCacheModifiedEventsDeletedLast(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub", []byte(`{
		"subscriptionId":"testsub",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc",
		"path": "/"}`))
	scache.Deleted("testsub", []byte(`{
		"subscriptionId":"testsub",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc",
		"path": "/"}`))

	assert.Equal(t, []libkv.FunctionKey(nil), scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedInvokable(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"space": "space1",
		"type": "async",
		"eventType": "invoke",
		"functionId": "testfunc1",
		"path": "/"}`))
	scache.Modified("testsub2", []byte(`{
		"subscriptionId":"testsub2",
		"space": "space1",
		"type": "async",
		"eventType": "invoke",
		"functionId": "testfunc2",
		"path": "/"}`))

	_, exists := scache.invokable["/"][libkv.FunctionKey{Space: "space1", ID: function.ID("testfunc1")}]
	assert.Equal(t, true, exists)
	_, exists = scache.invokable["/"][libkv.FunctionKey{Space: "space1", ID: function.ID("testfunc2")}]
	assert.Equal(t, true, exists)
}

func TestSubscriptionCacheModifiedInvokableDeleted(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"space": "space1",
		"type": "sync",
		"event": "invoke",
		"functionId": "testfunc1",
		"path": "/"}`))
	scache.Deleted("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"space": "space1",
		"type": "sync",
		"eventType": "invoke",
		"functionId": "testfunc1",
		"path": "/"}`))

	_, exists := scache.invokable["/"][libkv.FunctionKey{Space: "space1", ID: function.ID("testfunc1")}]
	assert.Equal(t, false, exists)
}
