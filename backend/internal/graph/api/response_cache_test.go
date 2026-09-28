package api

import (
	"strings"
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
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
	q := "query { currentProject { id } }"

	t.Run("deterministic for equal inputs", func(t *testing.T) {
		vars := map[string]any{"a": 1, "b": "x"}
		same := map[string]any{"b": "x", "a": 1} // map order must not matter
		assert.Equal(t, responseCacheKey(q, vars, "en", ""), responseCacheKey(q, same, "en", ""))
	})

	t.Run("query, language, variables and user all separate entries", func(t *testing.T) {
		base := responseCacheKey(q, nil, "en", "")
		assert.NotEqual(t, base, responseCacheKey("query { currentProject { name } }", nil, "en", ""))
		assert.NotEqual(t, base, responseCacheKey(q, nil, "de", ""))
		assert.NotEqual(t, base, responseCacheKey(q, map[string]any{"first": 5}, "en", ""))
		assert.NotEqual(t,
			responseCacheKey(q, map[string]any{"first": 5}, "en", ""),
			responseCacheKey(q, map[string]any{"first": 50}, "en", ""))
		assert.NotEqual(t, base, responseCacheKey(q, nil, "en", "US01"))
		assert.NotEqual(t,
			responseCacheKey(q, nil, "en", "US01"),
			responseCacheKey(q, nil, "en", "US02"))
	})

	t.Run("key shape routes to the right invalidation prefixes", func(t *testing.T) {
		shared := responseCacheKey(q, nil, "en", "")
		user := responseCacheKey(q, nil, "en", "US01")
		assert.True(t, strings.HasPrefix(shared, cache.PrefixGQLResponseShared))
		assert.True(t, strings.HasPrefix(user, cache.GQLResponseUserPrefix("US01")))
		assert.True(t, strings.HasPrefix(shared, cache.PrefixGQLResponse))
		assert.True(t, strings.HasPrefix(user, cache.PrefixGQLResponse))
	})
}

func parseOp(t *testing.T, query string) (*ast.OperationDefinition, ast.FragmentDefinitionList) {
	t.Helper()
	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	require.NoError(t, err)
	require.Len(t, doc.Operations, 1)
	return doc.Operations[0], doc.Fragments
}

func TestOperationShareable(t *testing.T) {
	t.Run("pure branding and leaderboard standings are shareable", func(t *testing.T) {
		op, frags := parseOp(t, `query {
			myCurrentProject {
				id
				name
				branding { logoImage { url width height blurhash } rounding }
				leaderboard(entityType: PERSONS, first: 10) {
					totalCount
					edges { node { id name score rank tags } }
				}
			}
		}`)
		assert.True(t, operationShareable(op, frags))
	})

	t.Run("leaderboard me is per-user", func(t *testing.T) {
		op, frags := parseOp(t, `query {
			myCurrentProject { leaderboard(entityType: PERSONS) { me { id rank } } }
		}`)
		assert.False(t, operationShareable(op, frags))
	})

	t.Run("user-scoped project children are per-user", func(t *testing.T) {
		for _, q := range []string{
			`query { myCurrentProject { myTeam { joinCode } } }`,
			`query { myCurrentProject { myPoints } }`,
			`query { myCurrentProject { activeChallenges { id } } }`,
			`query { myCurrentProject { completedChallenges { id } } }`,
			`query { myCurrentProject { activeChallengesCount } }`,
			`query { myCurrentProject { journal { totalCount } } }`,
		} {
			op, frags := parseOp(t, q)
			assert.False(t, operationShareable(op, frags), q)
		}
	})

	t.Run("unknown fields conservatively force per-user", func(t *testing.T) {
		op, frags := parseOp(t, `query { myCurrentProject { someFutureField } }`)
		assert.False(t, operationShareable(op, frags))
	})

	t.Run("fragments are walked", func(t *testing.T) {
		op, frags := parseOp(t, `query { myCurrentProject { ...F } }
			fragment F on Project { myPoints }`)
		assert.False(t, operationShareable(op, frags))

		op, frags = parseOp(t, `query { myCurrentProject { ...G } }
			fragment G on Project { id name }`)
		assert.True(t, operationShareable(op, frags))
	})
}
