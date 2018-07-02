package cache

import (
	"testing"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/libkv"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestSubscriptionCacheModified(t *testing.T) {
	t.Run("async added", func(t *testing.T) {
		scache := newSubscriptionCache(zap.NewNop())

		scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc1",
		"method": "GET",
		"path": "/"}`))
		scache.Modified("testsub2", []byte(`{
		"subscriptionId":
		"testsub2",
		"space": "space1",
		"type": "async",
		"eventType": "test.event",
		"functionId": "testfunc2",
		"method": "GET",
		"path": "/"}`))

		expected := []libkv.FunctionKey{
			libkv.FunctionKey{Space: "space1", ID: "testfunc1"},
			libkv.FunctionKey{Space: "space1", ID: "testfunc2"},
		}
		assert.Equal(t, expected, scache.async["GET"]["/"]["test.event"])
	})

	t.Run("sync added", func(t *testing.T) {
		scache := newSubscriptionCache(zap.NewNop())

		scache.Modified("testsub1", []byte(`{
		"subscriptionId":"testsub1",
		"type":"sync",
		"space": "default",
		"eventType": "http.request",
		"functionId": "testfunc1",
		"path": "/a",
		"method": "GET"}`))
		scache.Modified("testsub2", []byte(`{
		"subscriptionId":"testsub2",
		"type":"sync",
		"space": "default",
		"eventType": "http.request",
		"functionId": "testfunc2",
		"path": "/b",
		"method": "GET"}`))

		value, _ := scache.sync["GET"][eventpkg.TypeHTTPRequest].Resolve("/a")
		key := value.(libkv.FunctionKey)
		assert.Equal(t, function.ID("testfunc1"), key.ID)
		assert.Equal(t, "default", key.Space)
	})

	t.Run("wrong payload", func(t *testing.T) {
		scache := newSubscriptionCache(zap.NewNop())

		scache.Modified("testsub", []byte(`not json`))

		assert.Equal(t, []libkv.FunctionKey(nil), scache.async["POST"]["/"]["test.event"])
	})

	t.Run("async deleted", func(t *testing.T) {
		scache := newSubscriptionCache(zap.NewNop())

		scache.Modified("testsub1", []byte(`{
			"subscriptionId":"testsub1",
			"space": "space1",
			"type": "async",
			"eventType": "test.event",
			"functionId": "testfunc1",
			"method": "POST",
			"path": "/"}`))
		scache.Modified("testsub2", []byte(`{
			"subscriptionId":"testsub2",
			"space": "space1",
			"type": "async",
			"eventType": "test.event",
			"functionId": "testfunc2",
			"method": "POST",
			"path": "/"}`))
		scache.Deleted("testsub1", []byte(`{
			"subscriptionId":"testsub1",
			"space": "space1",
			"type": "async",
			"eventType": "test.event",
			"functionId": "testfunc1",
			"method": "POST",
			"path": "/"}`))

		assert.Equal(t, []libkv.FunctionKey{{Space: "space1", ID: function.ID("testfunc2")}}, scache.async["POST"]["/"]["test.event"])
	})

	t.Run("sync deleted", func(t *testing.T) {
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

		value, _ := scache.sync["GET"][eventpkg.TypeHTTPRequest].Resolve("/")
		assert.Nil(t, value)
	})

	t.Run("async deleted last", func(t *testing.T) {
		scache := newSubscriptionCache(zap.NewNop())

		scache.Modified("testsub", []byte(`{
			"subscriptionId":"testsub",
			"space": "space1",
			"type": "async",
			"eventType": "test.event",
			"functionId": "testfunc",
			"method": "POST",
			"path": "/"}`))
		scache.Deleted("testsub", []byte(`{
			"subscriptionId":"testsub",
			"space": "space1",
			"type": "async",
			"eventType": "test.event",
			"functionId": "testfunc",
			"method": "POST",
			"path": "/"}`))

		assert.Equal(t, []libkv.FunctionKey(nil), scache.async["POST"]["/"]["test.event"])
	})
}
