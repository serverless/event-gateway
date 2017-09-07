package pathtree

import (
	"testing"

	"github.com/serverless/event-gateway/functions"
	"github.com/stretchr/testify/assert"
)

func TestResolve_Root(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/", functions.FunctionID("testid"))

	functionID, _ := tree.Resolve("/")

	assert.Equal(t, functions.FunctionID("testid"), *functionID)
}

func TestResolve_NoRoot(t *testing.T) {
	tree := NewNode()

	functionID, _ := tree.Resolve("/")

	assert.Nil(t, functionID)
}

func TestResolve_Static(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a", functions.FunctionID("testid1"))
	tree.AddRoute("/b", functions.FunctionID("testid2"))
	tree.AddRoute("/a/b", functions.FunctionID("testid3"))
	tree.AddRoute("/d/e/f", functions.FunctionID("testid4"))

	functionID, _ := tree.Resolve("/a")
	assert.Equal(t, functions.FunctionID("testid1"), *functionID)

	functionID, _ = tree.Resolve("/b")
	assert.Equal(t, functions.FunctionID("testid2"), *functionID)

	functionID, _ = tree.Resolve("/a/b")
	assert.Equal(t, functions.FunctionID("testid3"), *functionID)

	functionID, _ = tree.Resolve("/d/e/f")
	assert.Equal(t, functions.FunctionID("testid4"), *functionID)
}

func TestResolve_NoPath(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/b/c/d", functions.FunctionID("testid1"))

	functionID, _ := tree.Resolve("/b")
	assert.Nil(t, functionID)
	functionID, _ = tree.Resolve("/a/b")
	assert.Nil(t, functionID)
}

func TestResolve_TrailingSlash(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/", functions.FunctionID("testid1"))

	functionID, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	functionID, _ = tree.Resolve("/a/")
	assert.Equal(t, functions.FunctionID("testid1"), *functionID)
}

func TestResolve_Param(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name", functions.FunctionID("testid1"))
	tree.AddRoute("/:name/:id", functions.FunctionID("testid2"))

	functionID, params := tree.Resolve("/foo")
	assert.Equal(t, functions.FunctionID("testid1"), *functionID)
	assert.EqualValues(t, Params{"name": "foo"}, params)

	functionID, params = tree.Resolve("/foo/1")
	assert.Equal(t, functions.FunctionID("testid2"), *functionID)
	assert.EqualValues(t, Params{"name": "foo", "id": "1"}, params)
}

func TestResolve_ParamNoMatch(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name", functions.FunctionID("testid1"))

	functionID, _ := tree.Resolve("/foo/bar/baz")
	assert.Nil(t, functionID)
}

func TestResolve_ParamAndStatic(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name/bar/:id", functions.FunctionID("testid1"))

	functionID, params := tree.Resolve("/foo/bar/baz")
	assert.Equal(t, functions.FunctionID("testid1"), *functionID)
	assert.EqualValues(t, Params{"name": "foo", "id": "baz"}, params)
}

func TestResolve_ParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", functions.FunctionID("testid1"))

	assert.Panics(t, func() { tree.AddRoute("/:bar", functions.FunctionID("testid2")) })
}

func TestResolve_StaticParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", functions.FunctionID("testid1"))

	assert.Panics(t, func() { tree.AddRoute("/bar", functions.FunctionID("testid2")) })
}

func TestResolve_StaticParamConflictDiffLevels(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo/:bar", functions.FunctionID("testid1"))

	assert.Panics(t, func() { tree.AddRoute("/baz", functions.FunctionID("testid2")) })
}

func TestResolve_Wildcard(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", functions.FunctionID("testid1"))

	functionID, params := tree.Resolve("/foo/bar/baz")
	assert.Equal(t, functions.FunctionID("testid1"), *functionID)
	assert.EqualValues(t, Params{"foo": "foo/bar/baz"}, params)
}

func TestResolve_WildcardNotLast(t *testing.T) {
	tree := NewNode()

	assert.Panics(t, func() { tree.AddRoute("/*foo/bar", functions.FunctionID("testid1")) })
}

func TestResolve_WildcardConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", functions.FunctionID("testid1"))

	assert.Panics(t, func() { tree.AddRoute("/*bar", functions.FunctionID("testid2")) })
}

func TestResolve_WildcardParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", functions.FunctionID("testid1"))

	assert.Panics(t, func() { tree.AddRoute("/bar", functions.FunctionID("testid2")) })
	assert.Panics(t, func() { tree.AddRoute("/:baz", functions.FunctionID("testid2")) })
}

func TestResolve_ParamWildcardConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", functions.FunctionID("testid1"))

	assert.Panics(t, func() { tree.AddRoute("/*bar", functions.FunctionID("testid2")) })
}

func TestDeleteRoute_Root(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/", functions.FunctionID("testid"))
	tree.DeleteRoute("/")

	functionID, _ := tree.Resolve("/")

	assert.Nil(t, functionID)
}

func TestDeleteRoute_Static(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/b/c", functions.FunctionID("testid1"))
	tree.DeleteRoute("/a/b/c")

	functionID, _ := tree.Resolve("/a/b/c")

	assert.Nil(t, functionID)
}

func TestDeleteRoute_StaticWithChild(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a", functions.FunctionID("testid1"))
	tree.AddRoute("/a/b", functions.FunctionID("testid2"))
	tree.DeleteRoute("/a")

	functionID, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	functionID, _ = tree.Resolve("/a/b")
	assert.Equal(t, functions.FunctionID("testid2"), *functionID)
}

func TestDeleteRoute_ParamWithChild(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", functions.FunctionID("testid1"))
	tree.AddRoute("/:foo/bar", functions.FunctionID("testid2"))

	tree.DeleteRoute("/:foo")

	functionID, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	functionID, _ = tree.Resolve("/a/bar")
	assert.Equal(t, functions.FunctionID("testid2"), *functionID)
}

func TestDeleteRoute_NonExisting(t *testing.T) {
	tree := NewNode()

	assert.Panics(t, func() { tree.DeleteRoute("/a") })
}

func TestDeleteRoute_DeleteParamAddStatic(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", functions.FunctionID("testid1"))
	tree.DeleteRoute("/:foo")
	tree.AddRoute("/a", functions.FunctionID("testid2"))

	functionID, _ := tree.Resolve("/a")
	assert.Equal(t, functions.FunctionID("testid2"), *functionID)
}
