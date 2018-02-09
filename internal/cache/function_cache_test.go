package cache

import (
	"testing"

	"github.com/serverless/event-gateway/api"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestFunctionCacheModified(t *testing.T) {
	fcache := newFunctionCache(zap.NewNop())

	fcache.Modified("testfunc1", []byte(`{"functionId":"testfunc1"}`))
	fcache.Modified("testfunc2", []byte(`{"functionId":"testfunc2"}`))

	id1 := api.FunctionID("testfunc1")
	id2 := api.FunctionID("testfunc2")
	assert.Equal(t, &api.Function{ID: id1}, fcache.cache[id1])
	assert.Equal(t, &api.Function{ID: id2}, fcache.cache[id2])
}

func TestFunctionCacheModified_WrongPayload(t *testing.T) {
	fcache := newFunctionCache(zap.NewNop())

	fcache.Modified("testfunc1", []byte(`not json`))

	assert.Equal(t, map[api.FunctionID]*api.Function{}, fcache.cache)
}

func TestFunctionCacheModifiedDeleted(t *testing.T) {
	fcache := newFunctionCache(zap.NewNop())

	fcache.Modified("testfunc1", []byte(`{"functionId":"testfunc1"}`))
	fcache.Modified("testfunc2", []byte(`{"functionId":"testfunc2"}`))
	fcache.Deleted("testfunc2", []byte(`{"functionId":"testfunc2"}`))

	id1 := api.FunctionID("testfunc1")
	assert.Equal(t, map[api.FunctionID]*api.Function{id1: &api.Function{ID: id1}}, fcache.cache)
}
