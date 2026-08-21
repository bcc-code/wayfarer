package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
)

func opWithFields(opType ast.Operation, fields ...string) *ast.OperationDefinition {
	sels := ast.SelectionSet{}
	for _, f := range fields {
		sels = append(sels, &ast.Field{Name: f})
	}
	return &ast.OperationDefinition{Operation: opType, SelectionSet: sels}
}

func TestCacheableOperation(t *testing.T) {
	t.Run("currentProject query is cacheable", func(t *testing.T) {
		assert.True(t, cacheableOperation(opWithFields(ast.Query, "currentProject")))
		assert.True(t, cacheableOperation(opWithFields(ast.Query, "myCurrentProject")))
		assert.True(t, cacheableOperation(opWithFields(ast.Query, "myCurrentProject", "__typename")))
	})

	t.Run("user-dependent fields are not cacheable", func(t *testing.T) {
		assert.False(t, cacheableOperation(opWithFields(ast.Query, "me")))
		assert.False(t, cacheableOperation(opWithFields(ast.Query, "currentProject", "me")))
	})

	t.Run("mutations are never cacheable", func(t *testing.T) {
		assert.False(t, cacheableOperation(opWithFields(ast.Mutation, "currentProject")))
	})

	t.Run("fragment spreads opt out", func(t *testing.T) {
		op := &ast.OperationDefinition{
			Operation:    ast.Query,
			SelectionSet: ast.SelectionSet{&ast.FragmentSpread{Name: "F"}},
		}
		assert.False(t, cacheableOperation(op))
	})

	t.Run("nil and empty operations are not cacheable", func(t *testing.T) {
		assert.False(t, cacheableOperation(nil))
		assert.False(t, cacheableOperation(opWithFields(ast.Query)))
	})
}

func TestResponseCacheKey(t *testing.T) {
	base := responseCacheKey("query { currentProject { id } }", "en")
	assert.Equal(t, base, responseCacheKey("query { currentProject { id } }", "en"))
	assert.NotEqual(t, base, responseCacheKey("query { currentProject { id } }", "de"))
	assert.NotEqual(t, base, responseCacheKey("query { currentProject { name } }", "en"))
}
