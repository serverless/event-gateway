package cache

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/libkv"
	"github.com/serverless/event-gateway/providers/http"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestFunctionCacheModified(t *testing.T) {
	t.Run("added", func(t *testing.T) {
		fcache := newFunctionCache(zap.NewNop())

		fcache.Modified(
			"default/testfunc1",
			[]byte(`{"functionId":"testfunc1","space":"default","type":"http","provider":{"url":"http://e.io"}}`),
		)
		fcache.Modified(
			"default/testfunc2",
			[]byte(`{"functionId":"testfunc2","space":"default","type":"http","provider":{"url":"http://e.io"}}`),
		)

		id1 := function.ID("testfunc1")
		id2 := function.ID("testfunc2")
		assert.Equal(
			t,
			&function.Function{ID: id1, Space: "default", ProviderType: http.Type, Provider: &http.HTTP{URL: "http://e.io"}},
			fcache.cache[libkv.FunctionKey{Space: "default", ID: id1}],
		)
		assert.Equal(
			t,
			&function.Function{ID: id2, Space: "default", ProviderType: http.Type, Provider: &http.HTTP{URL: "http://e.io"}},
			fcache.cache[libkv.FunctionKey{Space: "default", ID: id2}],
		)
	})

	t.Run("wrong payload", func(t *testing.T) {
		fcache := newFunctionCache(zap.NewNop())

		fcache.Modified("default/testfunc1", []byte(`not json`))

		assert.Equal(t, map[libkv.FunctionKey]*function.Function{}, fcache.cache)
	})

	t.Run("deleted", func(t *testing.T) {
		fcache := newFunctionCache(zap.NewNop())

		fcache.Modified(
			"default/testfunc1",
			[]byte(`{"functionId":"testfunc1","space":"default","type":"http","provider":{"url":"http://e.io"}}`),
		)
		fcache.Modified("default/testfunc2", []byte(`{"functionId":"testfunc2"}`))
		fcache.Deleted("default/testfunc2", []byte(`{"functionId":"testfunc2"}`))

		fid := function.ID("testfunc1")
		expected := map[libkv.FunctionKey]*function.Function{
			libkv.FunctionKey{Space: "default", ID: fid}: &function.Function{
				ID:           fid,
				Space:        "default",
				ProviderType: http.Type,
				Provider:     &http.HTTP{URL: "http://e.io"},
			},
		}
		assert.Equal(t, expected, fcache.cache)
	})
}
