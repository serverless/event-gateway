package pathtree

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/stretchr/testify/assert"
)

func TestAddRoute(t *testing.T) {
	t.Run("root conflict", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/", function.ID("testid1"))

		err := tree.AddRoute("/", function.ID("testid2"))

		assert.EqualError(t, err, "route / conflicts with existing route")
	})

	t.Run("static route conflict", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/a", function.ID("testid1"))

		err := tree.AddRoute("/a", function.ID("testid2"))

		assert.EqualError(t, err, "route /a conflicts with existing route")
	})

	t.Run("parametrized route conflict", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/:foo", function.ID("testid1"))

		err := tree.AddRoute("/:bar", function.ID("testid2"))

		assert.EqualError(t, err, `parameter with different name ("foo") already defined: for route: /:bar`)
	})

	t.Run("parameterized and static route conflict", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/:foo", function.ID("testid1"))

		err := tree.AddRoute("/bar", function.ID("testid2"))

		assert.EqualError(t, err, `parameter with different name ("foo") already defined: for route: /bar`)
	})

	t.Run("wildcard conflict", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/*foo", function.ID("testid1"))

		err := tree.AddRoute("/*bar", function.ID("testid2"))

		assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /*bar`)
	})

	t.Run("wildcard and paramerized route conflict", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/*foo", function.ID("testid1"))

		err := tree.AddRoute("/bar", function.ID("testid2"))
		assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /bar`)

		err = tree.AddRoute("/:bar", function.ID("testid2"))
		assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /:bar`)
	})

	t.Run("paramerized route and wildcard conflict", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/:foo", function.ID("testid1"))

		err := tree.AddRoute("/*bar", function.ID("testid2"))
		assert.EqualError(t, err, `wildcard with different name ("foo") already defined: for route: /*bar`)
	})

	t.Run("wildcard not last", func(t *testing.T) {
		tree := NewNode()

		err := tree.AddRoute("/*foo/bar", function.ID("testid1"))

		assert.EqualError(t, err, "wildcard must be the last parameter")
	})
}

func TestResolve(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/", function.ID("testid"))

		functionID, _ := tree.Resolve("/")

		assert.Equal(t, function.ID("testid"), functionID)
	})

	t.Run("no root", func(t *testing.T) {
		tree := NewNode()

		functionID, _ := tree.Resolve("/")

		assert.Nil(t, functionID)
	})

	t.Run("static", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/a", function.ID("testid1"))
		tree.AddRoute("/b", function.ID("testid2"))
		tree.AddRoute("/a/b", function.ID("testid3"))
		tree.AddRoute("/d/e/f", function.ID("testid4"))

		functionID, _ := tree.Resolve("/a")
		assert.Equal(t, function.ID("testid1"), functionID)

		functionID, _ = tree.Resolve("/b")
		assert.Equal(t, function.ID("testid2"), functionID)

		functionID, _ = tree.Resolve("/a/b")
		assert.Equal(t, function.ID("testid3"), functionID)

		functionID, _ = tree.Resolve("/d/e/f")
		assert.Equal(t, function.ID("testid4"), functionID)
	})

	t.Run("no path", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/a/b/c/d", function.ID("testid1"))

		functionID, _ := tree.Resolve("/b")
		assert.Nil(t, functionID)
		functionID, _ = tree.Resolve("/a/b")
		assert.Nil(t, functionID)
	})

	t.Run("trailing slash", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/a/", function.ID("testid1"))

		functionID, _ := tree.Resolve("/a")
		assert.Nil(t, functionID)
		functionID, _ = tree.Resolve("/a/")
		assert.Equal(t, function.ID("testid1"), functionID)
	})

	t.Run("param", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/:name", function.ID("testid1"))
		tree.AddRoute("/:name/:id", function.ID("testid2"))

		functionID, params := tree.Resolve("/foo")
		assert.Equal(t, function.ID("testid1"), functionID)
		assert.EqualValues(t, Params{"name": "foo"}, params)

		functionID, params = tree.Resolve("/foo/1")
		assert.Equal(t, function.ID("testid2"), functionID)
		assert.EqualValues(t, Params{"name": "foo", "id": "1"}, params)
	})

	t.Run("param doesn't match", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/:name", function.ID("testid1"))

		functionID, _ := tree.Resolve("/foo/bar/baz")
		assert.Nil(t, functionID)
	})

	t.Run("param and static", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/:name/bar/:id", function.ID("testid1"))

		functionID, params := tree.Resolve("/foo/bar/baz")
		assert.Equal(t, function.ID("testid1"), functionID)
		assert.EqualValues(t, Params{"name": "foo", "id": "baz"}, params)
	})

	t.Run("wildcard", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/*foo", function.ID("testid1"))

		functionID, params := tree.Resolve("/foo/bar/baz")
		assert.Equal(t, function.ID("testid1"), functionID)
		assert.EqualValues(t, Params{"foo": "foo/bar/baz"}, params)
	})
}

func TestDeleteRoute_Root(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/", function.ID("testid"))
		tree.DeleteRoute("/")

		functionID, _ := tree.Resolve("/")

		assert.Nil(t, functionID)
	})

	t.Run("static", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/a/b/c", function.ID("testid1"))
		tree.DeleteRoute("/a/b/c")

		functionID, _ := tree.Resolve("/a/b/c")

		assert.Nil(t, functionID)
	})

	t.Run("static without child", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/a", function.ID("testid1"))
		tree.AddRoute("/a/b", function.ID("testid2"))
		tree.DeleteRoute("/a")

		functionID, _ := tree.Resolve("/a")
		assert.Nil(t, functionID)
		functionID, _ = tree.Resolve("/a/b")
		assert.Equal(t, function.ID("testid2"), functionID)
	})

	t.Run("param with child", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/:foo", function.ID("testid1"))
		tree.AddRoute("/:foo/bar", function.ID("testid2"))

		tree.DeleteRoute("/:foo")

		functionID, _ := tree.Resolve("/a")
		assert.Nil(t, functionID)
		functionID, _ = tree.Resolve("/a/bar")
		assert.Equal(t, function.ID("testid2"), functionID)
	})

	t.Run("non existing route", func(t *testing.T) {
		tree := NewNode()

		err := tree.DeleteRoute("/a")
		assert.EqualError(t, err, "unable to delete node non existing node")
	})

	t.Run("param and static", func(t *testing.T) {
		tree := NewNode()
		tree.AddRoute("/:foo", function.ID("testid1"))
		tree.DeleteRoute("/:foo")
		tree.AddRoute("/a", function.ID("testid2"))

		functionID, _ := tree.Resolve("/a")
		assert.Equal(t, function.ID("testid2"), functionID)
	})
}
