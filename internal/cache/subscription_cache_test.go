package cache

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/subscription"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestSubscriptionCacheModifiedEvents(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "test.event", "functionId": "testfunc1", "path": "/"}`))
	scache.Modified("testsub2", []byte(`{"subscriptionId":"testsub2", "event": "test.event", "functionId": "testfunc2", "path": "/"}`))

	assert.Equal(
		t,
		[]function.ID{function.ID("testfunc1"), function.ID("testfunc2")},
		scache.eventToFunctions["/"]["test.event"],
	)
}

func TestSubscriptionCacheModifiedHTTPSubscription(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "http", "functionId": "testfunc1", "path": "/a", "method": "GET"}`))
	scache.Modified("testsub2", []byte(`{"subscriptionId":"testsub2", "event": "http", "functionId": "testfunc2", "path": "/b", "method": "GET"}`))

	id, _, _ := scache.endpoints["GET"].Resolve("/a")
	assert.Equal(t, function.ID("testfunc1"), *id)
}

func TestSubscriptionCacheModifiedCORSConfiguration(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"event": "http",
		"functionId": "testfunc1",
		"path": "/a",
		"method": "GET",
		"cors": {
			"origins": ["http://example.com"]
		}
	}`))

	_, _, corsConfig := scache.endpoints["GET"].Resolve("/a")
	assert.Equal(t, &subscription.CORS{Origins: []string{"http://example.com"}}, corsConfig)
}

func TestSubscriptionCacheModifiedEventsWrongPayload(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub", []byte(`not json`))

	assert.Equal(t, []function.ID(nil), scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedEventsDeleted(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "test.event", "functionId": "testfunc1", "path": "/"}`))
	scache.Modified("testsub2", []byte(`{"subscriptionId":"testsub2", "event": "test.event", "functionId": "testfunc2", "path": "/"}`))
	scache.Deleted("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "test.event", "functionId": "testfunc1", "path": "/"}`))

	assert.Equal(t, []function.ID{function.ID("testfunc2")}, scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedHTTPSubscriptionDeleted(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "http", "functionId": "testfunc1", "path": "/", "method": "GET"}`))
	scache.Deleted("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "http", "functionId": "testfunc1", "path": "/", "method": "GET"}`))

	id, _, _ := scache.endpoints["GET"].Resolve("/")
	assert.Nil(t, id)
}

func TestSubscriptionCacheModifiedEventsDeletedLast(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub", []byte(`{"subscriptionId":"testsub", "event": "test.event", "functionId": "testfunc", "path": "/"}`))
	scache.Deleted("testsub", []byte(`{"subscriptionId":"testsub", "event": "test.event", "functionId": "testfunc", "path": "/"}`))

	assert.Equal(t, []function.ID(nil), scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedInvokable(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "invoke", "functionId": "testfunc1", "path": "/"}`))
	scache.Modified("testsub2", []byte(`{"subscriptionId":"testsub2", "event": "invoke", "functionId": "testfunc2", "path": "/"}`))

	_, exists := scache.invokable["/"][function.ID("testfunc1")]
	assert.Equal(t, true, exists)
	_, exists = scache.invokable["/"][function.ID("testfunc2")]
	assert.Equal(t, true, exists)
}

func TestSubscriptionCacheModifiedInvokableDeleted(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "invoke", "functionId": "testfunc1", "path": "/"}`))
	scache.Deleted("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "invoke", "functionId": "testfunc1", "path": "/"}`))

	_, exists := scache.invokable["/"][function.ID("testfunc1")]
	assert.Equal(t, false, exists)
}
