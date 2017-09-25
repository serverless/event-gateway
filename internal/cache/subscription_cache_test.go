package cache

import (
	"testing"

	"github.com/serverless/event-gateway/functions"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestSubscriptionCacheModified(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "test.event", "functionId": "testfunc1", "path": "/"}`))
	scache.Modified("testsub2", []byte(`{"subscriptionId":"testsub2", "event": "test.event", "functionId": "testfunc2", "path": "/"}`))

	assert.Equal(
		t,
		[]functions.FunctionID{functions.FunctionID("testfunc1"), functions.FunctionID("testfunc2")},
		scache.eventToFunctions["/"]["test.event"],
	)
}

func TestSubscriptionCacheModified_WrongPayload(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub", []byte(`not json`))

	assert.Equal(t, []functions.FunctionID(nil), scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedDeleted(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "test.event", "functionId": "testfunc1", "path": "/"}`))
	scache.Modified("testsub2", []byte(`{"subscriptionId":"testsub2", "event": "test.event", "functionId": "testfunc2", "path": "/"}`))
	scache.Deleted("testsub1", []byte(`{"subscriptionId":"testsub1", "event": "test.event", "functionId": "testfunc1", "path": "/"}`))

	assert.Equal(t, []functions.FunctionID{functions.FunctionID("testfunc2")}, scache.eventToFunctions["/"]["test.event"])
}

func TestSubscriptionCacheModifiedDeletedLast(t *testing.T) {
	scache := newSubscriptionCache(zap.NewNop())

	scache.Modified("testsub", []byte(`{"subscriptionId":"testsub", "event": "test.event", "functionId": "testfunc", "path": "/"}`))
	scache.Deleted("testsub", []byte(`{"subscriptionId":"testsub", "event": "test.event", "functionId": "testfunc", "path": "/"}`))

	assert.Equal(t, []functions.FunctionID(nil), scache.eventToFunctions["/"]["test.event"])
}
