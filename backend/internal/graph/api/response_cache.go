package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/vektah/gqlparser/v2/ast"
)

// ResponseCache is a gqlgen extension that serves whole cached responses.
// Field resolution, dataloaders and JSON marshalling are all skipped on a hit.
//
// Only queries whose every top-level selection is in cacheableQueryFields are
// eligible. Eligible queries are cached under one of two key shapes:
//
//   - shared: every selected field (walked recursively, fragments included)
//     is in shareableFields, so the response is identical for every caller
//     and one entry serves all users.
//   - per-user: the selection touches any user-scoped field (myTeam,
//     myPoints, activeChallenges, leaderboard.me, ...) or any field not
//     explicitly known to be user-independent. The key then includes the
//     calling user's ID, so a cached response is only ever replayed to the
//     user it was produced for.
//
// The key always includes a hash of the operation variables and the request
// language. Entries expire by TTL; additionally, per-user entries are dropped
// when that user runs any mutation (read-your-own-writes) and when the user's
// caches are invalidated (cache.InvalidateUser and the enrollment
// invalidation), and every response entry is dropped when a project is
// invalidated, so admin edits surface immediately.
type ResponseCache struct {
	Cache *cache.CacheWithRegistry
	TTL   time.Duration
}

// cacheableQueryFields are top-level query fields eligible for response
// caching. This is a size/juice gate, not a correctness gate: correctness
// comes from the shared/per-user key split below.
var cacheableQueryFields = map[string]bool{
	"currentProject":   true,
	"myCurrentProject": true,
	"__typename":       true,
}

// shareableFields are fields (anywhere in the selection) whose resolvers are
// verified to not read the calling user's identity. A selection consisting
// solely of these fields produces a byte-identical response for every caller
// and may be shared cross-user. Any field NOT in this set — including fields
// added to the schema later — conservatively forces a per-user cache entry.
// Do NOT add a field here unless its resolver (and every child resolver
// reachable from it) ignores the user in context. "me" must never be added.
var shareableFields = map[string]bool{
	"__typename": true,

	// Project scalars (from the project row, no user in context)
	"id": true, "name": true, "description": true,
	"startDate": true, "endDate": true, "archivedAt": true,
	"rules": true, "infoMessage": true, "infoMessageStart": true, "infoMessageEnd": true,
	"markdown": true, "html": true,

	// Branding subtree
	"branding": true, "logoImage": true, "bannerImage": true,
	"url": true, "width": true, "height": true, "blurhash": true,
	"rounding": true, "colors": true, "light": true, "dark": true,
	"accent": true, "accentContrast": true, "onAccent": true,
	"backgroundDefault": true, "backgroundRaised": true, "backgroundIndent": true,
	"textDefault": true, "textMuted": true, "textHint": true,
	"shadowDefault": true, "shadowBlank": true, "borderDefault": true,

	// Leaderboard subtree WITHOUT `me` (global standings are the same for
	// everyone; `me` is per-user and is deliberately absent from this set)
	"leaderboard": true, "totalCount": true, "edges": true, "node": true, "cursor": true,
	"pageInfo": true, "hasNextPage": true, "hasPreviousPage": true,
	"startCursor": true, "endCursor": true,
	"score": true, "rank": true, "tags": true,
}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
} = ResponseCache{}

func (ResponseCache) ExtensionName() string { return "ResponseCache" }

func (ResponseCache) Validate(graphql.ExecutableSchema) error { return nil }

// cacheableOperation reports whether the operation is eligible for response
// caching at all (query with only whitelisted top-level fields).
func cacheableOperation(op *ast.OperationDefinition) bool {
	if op == nil || op.Operation != ast.Query || len(op.SelectionSet) == 0 {
		return false
	}
	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok || !cacheableQueryFields[field.Name] {
			return false // top-level fragment spreads and unknown fields opt out
		}
	}
	return true
}

// selectionShareable reports whether every field in the selection (fragments
// resolved) is verified user-independent. Unknown fields return false.
func selectionShareable(set ast.SelectionSet, fragments ast.FragmentDefinitionList) bool {
	for _, sel := range set {
		switch s := sel.(type) {
		case *ast.Field:
			if !shareableFields[s.Name] {
				return false
			}
			if !selectionShareable(s.SelectionSet, fragments) {
				return false
			}
		case *ast.FragmentSpread:
			frag := fragments.ForName(s.Name)
			if frag == nil || !selectionShareable(frag.SelectionSet, fragments) {
				return false
			}
		case *ast.InlineFragment:
			if !selectionShareable(s.SelectionSet, fragments) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// operationShareable checks the children of the (already validated) top-level
// fields — the top-level names themselves are the eligibility gate, their
// subtrees decide shared vs per-user.
func operationShareable(op *ast.OperationDefinition, fragments ast.FragmentDefinitionList) bool {
	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			return false
		}
		if !selectionShareable(field.SelectionSet, fragments) {
			return false
		}
	}
	return true
}

// responseCacheKey builds the cache key. userID is empty for shared entries.
// encoding/json marshals map keys in sorted order, so the variables hash is
// deterministic for equal variable sets.
func responseCacheKey(rawQuery string, variables map[string]any, lang, userID string) string {
	qsum := sha256.Sum256([]byte(rawQuery))
	varsJSON, err := json.Marshal(variables)
	if err != nil {
		varsJSON = []byte("unmarshalable")
	}
	vsum := sha256.Sum256(varsJSON)
	suffix := hex.EncodeToString(qsum[:]) + ":" + hex.EncodeToString(vsum[:]) + ":" + lang
	if userID == "" {
		return cache.PrefixGQLResponseShared + suffix
	}
	return cache.GQLResponseUserPrefix(userID) + suffix
}

func (r ResponseCache) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	oc := graphql.GetOperationContext(ctx)
	if r.Cache == nil || oc == nil || oc.Operation == nil {
		return next(ctx)
	}

	// Read-your-own-writes: any successful mutation drops the caller's
	// per-user response entries, so the next read reflects the write even
	// inside the TTL window. Shared entries contain no user state and are
	// unaffected by per-user mutations.
	if oc.Operation.Operation == ast.Mutation {
		resp := next(ctx)
		if userID, ok := middleware.GetUserID(ctx); ok && userID != "" {
			if resp == nil || len(resp.Errors) == 0 {
				r.Cache.DeletePrefix(cache.GQLResponseUserPrefix(userID))
			}
		}
		return resp
	}

	if !cacheableOperation(oc.Operation) {
		return next(ctx)
	}

	var fragments ast.FragmentDefinitionList
	if oc.Doc != nil {
		fragments = oc.Doc.Fragments
	}
	userID := ""
	if !operationShareable(oc.Operation, fragments) {
		uid, ok := middleware.GetUserID(ctx)
		if !ok || uid == "" {
			// User-scoped selection without an authenticated user: nothing
			// sensible to key on, skip caching.
			return next(ctx)
		}
		userID = uid
	}

	key := responseCacheKey(oc.RawQuery, oc.Variables, middleware.GetLanguage(ctx), userID)
	if cached, ok := r.Cache.Get(key); ok {
		if resp, ok := cached.(*graphql.Response); ok && resp != nil {
			return resp
		}
	}

	resp := next(ctx)
	if resp != nil && len(resp.Errors) == 0 && resp.Data != nil {
		r.Cache.SetWithTTL(key, resp, r.TTL)
	}
	return resp
}
