package pathtree

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/stretchr/testify/assert"
)

func TestResolve_Root(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/", "default", function.ID("testid"))

	space, functionID, _ := tree.Resolve("/")

	assert.Equal(t, function.ID("testid"), *functionID)
	assert.Equal(t, "default", space)
}

func TestResolve_NoRoot(t *testing.T) {
	tree := NewNode()

	space, functionID, _ := tree.Resolve("/")

	assert.Nil(t, functionID)
	assert.Equal(t, "", space)
}

func TestResolve_Static(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a", "default", function.ID("testid1"))
	tree.AddRoute("/b", "default", function.ID("testid2"))
	tree.AddRoute("/a/b", "default", function.ID("testid3"))
	tree.AddRoute("/d/e/f", "default", function.ID("testid4"))

	space, functionID, _ := tree.Resolve("/a")
	assert.Equal(t, function.ID("testid1"), *functionID)
	assert.Equal(t, "default", space)

	space, functionID, _ = tree.Resolve("/b")
	assert.Equal(t, function.ID("testid2"), *functionID)
	assert.Equal(t, "default", space)

	space, functionID, _ = tree.Resolve("/a/b")
	assert.Equal(t, function.ID("testid3"), *functionID)
	assert.Equal(t, "default", space)

	space, functionID, _ = tree.Resolve("/d/e/f")
	assert.Equal(t, function.ID("testid4"), *functionID)
	assert.Equal(t, "default", space)
}

func TestResolve_StaticConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a", "default", function.ID("testid1"))

	err := tree.AddRoute("/a", "default", function.ID("testid2"))

	assert.EqualError(t, err, "route /a conflicts with existing route")
}

func TestResolve_NoPath(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/b/c/d", "default", function.ID("testid1"))

	space, functionID, _ := tree.Resolve("/b")
	assert.Nil(t, functionID)
	assert.Equal(t, "", space)
	space, functionID, _ = tree.Resolve("/a/b")
	assert.Nil(t, functionID)
	assert.Equal(t, "", space)
}

func TestResolve_TrailingSlash(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/", "default", function.ID("testid1"))

	_, functionID, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	_, functionID, _ = tree.Resolve("/a/")
	assert.Equal(t, function.ID("testid1"), *functionID)
}

func TestResolve_Param(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name", "default", function.ID("testid1"))
	tree.AddRoute("/:name/:id", "default", function.ID("testid2"))

	_, functionID, params := tree.Resolve("/foo")
	assert.Equal(t, function.ID("testid1"), *functionID)
	assert.EqualValues(t, Params{"name": "foo"}, params)

	_, functionID, params = tree.Resolve("/foo/1")
	assert.Equal(t, function.ID("testid2"), *functionID)
	assert.EqualValues(t, Params{"name": "foo", "id": "1"}, params)
}

func TestResolve_ParamNoMatch(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name", "default", function.ID("testid1"))

	_, functionID, _ := tree.Resolve("/foo/bar/baz")
	assert.Nil(t, functionID)
}

func TestResolve_ParamAndStatic(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name/bar/:id", "default", function.ID("testid1"))

	_, functionID, params := tree.Resolve("/foo/bar/baz")
	assert.Equal(t, function.ID("testid1"), *functionID)
	assert.EqualValues(t, Params{"name": "foo", "id": "baz"}, params)
}

func TestResolve_ParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", "default", function.ID("testid1"))

	err := tree.AddRoute("/:bar", "default", function.ID("testid2"))

	assert.EqualError(t, err, `parameter with different name ("foo") already defined: for route: /:bar`)
}

func TestResolve_ParamStaticConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", "default", function.ID("testid1"))

	err := tree.AddRoute("/bar", "default", function.ID("testid2"))

	assert.EqualError(t, err, `parameter with different name ("foo") already defined: for route: /bar`)
}

func TestResolve_StaticParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/foo/:bar", "default", function.ID("testid1"))

	err := tree.AddRoute("/:bar", "default", function.ID("testid2"))

	assert.EqualError(t, err, "static route already defined for route: /:bar")
}

func TestResolve_Wildcard(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", "default", function.ID("testid1"))

	_, functionID, params := tree.Resolve("/foo/bar/baz")
	assert.Equal(t, function.ID("testid1"), *functionID)
	assert.EqualValues(t, Params{"foo": "foo/bar/baz"}, params)
}

func TestResolve_WildcardNotLast(t *testing.T) {
	tree := NewNode()

	err := tree.AddRoute("/*foo/bar", "default", function.ID("testid1"))

	assert.EqualError(t, err, "wildcard must be the last parameter")
}

func TestResolve_WildcardConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", "default", function.ID("testid1"))

	err := tree.AddRoute("/*bar", "default", function.ID("testid2"))

	assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /*bar`)
}

func TestResolve_WildcardParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", "default", function.ID("testid1"))

	err := tree.AddRoute("/bar", "default", function.ID("testid2"))
	assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /bar`)

	err = tree.AddRoute("/:bar", "default", function.ID("testid2"))
	assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /:bar`)
}

func TestResolve_ParamWildcardConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", "default", function.ID("testid1"))

	err := tree.AddRoute("/*bar", "default", function.ID("testid2"))
	assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /*bar`)
}

func TestDeleteRoute_Root(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/", "default", function.ID("testid"))
	tree.DeleteRoute("/")

	_, functionID, _ := tree.Resolve("/")

	assert.Nil(t, functionID)
}

func TestDeleteRoute_Static(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/b/c", "default", function.ID("testid1"))
	tree.DeleteRoute("/a/b/c")

	_, functionID, _ := tree.Resolve("/a/b/c")

	assert.Nil(t, functionID)
}

func TestDeleteRoute_StaticWithChild(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a", "default", function.ID("testid1"))
	tree.AddRoute("/a/b", "default", function.ID("testid2"))
	tree.DeleteRoute("/a")

	_, functionID, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	_, functionID, _ = tree.Resolve("/a/b")
	assert.Equal(t, function.ID("testid2"), *functionID)
}

func TestDeleteRoute_ParamWithChild(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", "default", function.ID("testid1"))
	tree.AddRoute("/:foo/bar", "default", function.ID("testid2"))

	tree.DeleteRoute("/:foo")

	_, functionID, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	_, functionID, _ = tree.Resolve("/a/bar")
	assert.Equal(t, function.ID("testid2"), *functionID)
}

func TestDeleteRoute_NonExisting(t *testing.T) {
	tree := NewNode()

	err := tree.DeleteRoute("/a")
	assert.EqualError(t, err, "unable to delete node non existing node")
}

func TestDeleteRoute_DeleteParamAddStatic(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", "default", function.ID("testid1"))
	tree.DeleteRoute("/:foo")
	tree.AddRoute("/a", "default", function.ID("testid2"))

	_, functionID, _ := tree.Resolve("/a")
	assert.Equal(t, function.ID("testid2"), *functionID)
}
