package cache

import (
	"testing"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/libkv"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestEventTypeCacheModified(t *testing.T) {
	typesCache := newEventTypeCache(zap.NewNop())

	typesCache.Modified("default/user.created", []byte(`{"name":"user.created","space":"default"}`))
	typesCache.Modified("default/user.deleted", []byte(`{"name":"user.deleted","space":"default"}`))

	name1 := eventpkg.TypeName("user.created")
	name2 := eventpkg.TypeName("user.deleted")
	assert.Equal(t,
		&eventpkg.Type{Name: name1, Space: "default"},
		typesCache.cache[libkv.EventTypeKey{Space: "default", Name: name1}])
	assert.Equal(t,
		&eventpkg.Type{Name: name2, Space: "default"},
		typesCache.cache[libkv.EventTypeKey{Space: "default", Name: name2}])
}

func TestEventTypeCacheModified_WrongPayload(t *testing.T) {
	typesCache := newEventTypeCache(zap.NewNop())

	typesCache.Modified("default/user.created", []byte(`not json`))

	assert.Equal(t, map[libkv.EventTypeKey]*eventpkg.Type{}, typesCache.cache)
}

func TestEventTypeCacheModifiedDeleted(t *testing.T) {
	typesCache := newEventTypeCache(zap.NewNop())

	typesCache.Modified("default/user.created", []byte(`{"name":"user.created","space":"default"}`))
	typesCache.Modified("default/user.deleted", []byte(`{"name":"user.deleted","space":"default"}`))
	typesCache.Deleted("default/user.deleted", []byte(`{"name":"user.deleted","space":"default"}`))

	name := eventpkg.TypeName("user.created")
	expected := map[libkv.EventTypeKey]*eventpkg.Type{
		libkv.EventTypeKey{Space: "default", Name: name}: &eventpkg.Type{
			Name:  name,
			Space: "default",
		},
	}
	assert.Equal(t, expected, typesCache.cache)
}
