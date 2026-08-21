package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/vektah/gqlparser/v2/ast"
)

// ResponseCache is a gqlgen extension that serves whole cached responses for
// queries whose result is identical for every caller. Field resolution,
// dataloaders and JSON marshalling are all skipped on a hit.
//
// Only queries whose every top-level selection is in cacheableQueryFields are
// eligible — those resolvers must not read the calling user's identity. The
// cache key is the raw query text plus the request language, so any change in
// selection shape or Accept-Language gets its own entry. Entries expire by
// TTL only; keep it short enough that admin edits (e.g. branding) surface
// quickly.
type ResponseCache struct {
	Cache *cache.CacheWithRegistry
	TTL   time.Duration
}

// cacheableQueryFields are top-level query fields whose responses do not
// depend on the calling user. Do NOT add a field here unless its resolver
// (and every child resolver reachable from it) ignores the user in context.
var cacheableQueryFields = map[string]bool{
	"currentProject":   true,
	"myCurrentProject": true, // despite the name, resolves the global current project
	"__typename":       true,
}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
} = ResponseCache{}

func (ResponseCache) ExtensionName() string { return "ResponseCache" }

func (ResponseCache) Validate(graphql.ExecutableSchema) error { return nil }

// cacheableOperation reports whether the operation's result is user-independent.
func cacheableOperation(op *ast.OperationDefinition) bool {
	if op == nil || op.Operation != ast.Query || len(op.SelectionSet) == 0 {
		return false
	}
	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok || !cacheableQueryFields[field.Name] {
			return false // fragment spreads and unknown fields opt out
		}
	}
	return true
}

func responseCacheKey(rawQuery, lang string) string {
	sum := sha256.Sum256([]byte(rawQuery))
	return "gqlresponse:" + hex.EncodeToString(sum[:]) + ":" + lang
}

func (r ResponseCache) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	oc := graphql.GetOperationContext(ctx)
	if r.Cache == nil || oc == nil || !cacheableOperation(oc.Operation) {
		return next(ctx)
	}

	key := responseCacheKey(oc.RawQuery, middleware.GetLanguage(ctx))
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
