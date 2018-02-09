package pathtree

import (
	"testing"

	"github.com/serverless/event-gateway/api"
	"github.com/stretchr/testify/assert"
)

func TestResolve_Root(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/", api.FunctionID("testid"), &api.CORS{Origins: []string{"http://example.com"}})

	functionID, _, corsConfig := tree.Resolve("/")

	assert.Equal(t, api.FunctionID("testid"), *functionID)
	assert.Equal(t, &api.CORS{Origins: []string{"http://example.com"}}, corsConfig)
}

func TestResolve_NoRoot(t *testing.T) {
	tree := NewNode()

	functionID, _, _ := tree.Resolve("/")

	assert.Nil(t, functionID)
}

func TestResolve_Static(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a", api.FunctionID("testid1"), nil)
	tree.AddRoute("/b", api.FunctionID("testid2"), nil)
	tree.AddRoute("/a/b", api.FunctionID("testid3"), nil)
	tree.AddRoute("/d/e/f", api.FunctionID("testid4"), nil)

	functionID, _, _ := tree.Resolve("/a")
	assert.Equal(t, api.FunctionID("testid1"), *functionID)

	functionID, _, _ = tree.Resolve("/b")
	assert.Equal(t, api.FunctionID("testid2"), *functionID)

	functionID, _, _ = tree.Resolve("/a/b")
	assert.Equal(t, api.FunctionID("testid3"), *functionID)

	functionID, _, _ = tree.Resolve("/d/e/f")
	assert.Equal(t, api.FunctionID("testid4"), *functionID)
}

func TestResolve_StaticConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a", api.FunctionID("testid1"), nil)

	err := tree.AddRoute("/a", api.FunctionID("testid2"), nil)

	assert.EqualError(t, err, "route /a conflicts with existing route")
}

func TestResolve_NoPath(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/b/c/d", api.FunctionID("testid1"), nil)

	functionID, _, _ := tree.Resolve("/b")
	assert.Nil(t, functionID)
	functionID, _, _ = tree.Resolve("/a/b")
	assert.Nil(t, functionID)
}

func TestResolve_TrailingSlash(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/", api.FunctionID("testid1"), nil)

	functionID, _, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	functionID, _, _ = tree.Resolve("/a/")
	assert.Equal(t, api.FunctionID("testid1"), *functionID)
}

func TestResolve_Param(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name", api.FunctionID("testid1"), nil)
	tree.AddRoute("/:name/:id", api.FunctionID("testid2"), nil)

	functionID, params, _ := tree.Resolve("/foo")
	assert.Equal(t, api.FunctionID("testid1"), *functionID)
	assert.EqualValues(t, Params{"name": "foo"}, params)

	functionID, params, _ = tree.Resolve("/foo/1")
	assert.Equal(t, api.FunctionID("testid2"), *functionID)
	assert.EqualValues(t, Params{"name": "foo", "id": "1"}, params)
}

func TestResolve_ParamNoMatch(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name", api.FunctionID("testid1"), nil)

	functionID, _, _ := tree.Resolve("/foo/bar/baz")
	assert.Nil(t, functionID)
}

func TestResolve_ParamAndStatic(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:name/bar/:id", api.FunctionID("testid1"), nil)

	functionID, params, _ := tree.Resolve("/foo/bar/baz")
	assert.Equal(t, api.FunctionID("testid1"), *functionID)
	assert.EqualValues(t, Params{"name": "foo", "id": "baz"}, params)
}

func TestResolve_ParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", api.FunctionID("testid1"), nil)

	err := tree.AddRoute("/:bar", api.FunctionID("testid2"), nil)

	assert.EqualError(t, err, `parameter with different name ("foo") already defined: for route: /:bar`)
}

func TestResolve_ParamStaticConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", api.FunctionID("testid1"), nil)

	err := tree.AddRoute("/bar", api.FunctionID("testid2"), nil)

	assert.EqualError(t, err, `parameter with different name ("foo") already defined: for route: /bar`)
}

func TestResolve_StaticParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/foo/:bar", api.FunctionID("testid1"), nil)

	err := tree.AddRoute("/:bar", api.FunctionID("testid2"), nil)

	assert.EqualError(t, err, "static route already defined for route: /:bar")
}

func TestResolve_Wildcard(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", api.FunctionID("testid1"), nil)

	functionID, params, _ := tree.Resolve("/foo/bar/baz")
	assert.Equal(t, api.FunctionID("testid1"), *functionID)
	assert.EqualValues(t, Params{"foo": "foo/bar/baz"}, params)
}

func TestResolve_WildcardNotLast(t *testing.T) {
	tree := NewNode()

	err := tree.AddRoute("/*foo/bar", api.FunctionID("testid1"), nil)

	assert.EqualError(t, err, "wildcard must be the last parameter")
}

func TestResolve_WildcardConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", api.FunctionID("testid1"), nil)

	err := tree.AddRoute("/*bar", api.FunctionID("testid2"), nil)

	assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /*bar`)
}

func TestResolve_WildcardParamConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/*foo", api.FunctionID("testid1"), nil)

	err := tree.AddRoute("/bar", api.FunctionID("testid2"), nil)
	assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /bar`)

	err = tree.AddRoute("/:bar", api.FunctionID("testid2"), nil)
	assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /:bar`)
}

func TestResolve_ParamWildcardConflict(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", api.FunctionID("testid1"), nil)

	err := tree.AddRoute("/*bar", api.FunctionID("testid2"), nil)
	assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /*bar`)
}

func TestDeleteRoute_Root(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/", api.FunctionID("testid"), nil)
	tree.DeleteRoute("/")

	functionID, _, _ := tree.Resolve("/")

	assert.Nil(t, functionID)
}

func TestDeleteRoute_Static(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a/b/c", api.FunctionID("testid1"), nil)
	tree.DeleteRoute("/a/b/c")

	functionID, _, _ := tree.Resolve("/a/b/c")

	assert.Nil(t, functionID)
}

func TestDeleteRoute_StaticWithChild(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/a", api.FunctionID("testid1"), nil)
	tree.AddRoute("/a/b", api.FunctionID("testid2"), nil)
	tree.DeleteRoute("/a")

	functionID, _, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	functionID, _, _ = tree.Resolve("/a/b")
	assert.Equal(t, api.FunctionID("testid2"), *functionID)
}

func TestDeleteRoute_ParamWithChild(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", api.FunctionID("testid1"), nil)
	tree.AddRoute("/:foo/bar", api.FunctionID("testid2"), nil)

	tree.DeleteRoute("/:foo")

	functionID, _, _ := tree.Resolve("/a")
	assert.Nil(t, functionID)
	functionID, _, _ = tree.Resolve("/a/bar")
	assert.Equal(t, api.FunctionID("testid2"), *functionID)
}

func TestDeleteRoute_NonExisting(t *testing.T) {
	tree := NewNode()

	err := tree.DeleteRoute("/a")
	assert.EqualError(t, err, "unable to delete node non existing node")
}

func TestDeleteRoute_DeleteParamAddStatic(t *testing.T) {
	tree := NewNode()
	tree.AddRoute("/:foo", api.FunctionID("testid1"), nil)
	tree.DeleteRoute("/:foo")
	tree.AddRoute("/a", api.FunctionID("testid2"), nil)

	functionID, _, _ := tree.Resolve("/a")
	assert.Equal(t, api.FunctionID("testid2"), *functionID)
}
