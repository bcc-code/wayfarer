package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
)

const (
	cacheTestUserID    = "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	cacheTestProjectID = "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	cacheTestTeamID    = "TM01ARZ3NDEKTSV4RRFFQ69G5FAV"
)

// newCacheOnlyResolver builds a resolver with a live cache and loaders but no
// database. Any code path that reaches the database panics inside the loader
// batch function (surfaced as a thunk error), so these tests prove the
// cache-hit paths never touch it.
func newCacheOnlyResolver(t *testing.T) *Resolver {
	t.Helper()
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	return &Resolver{
		Cache:   c,
		Loaders: loaders.NewLoaders(nil, c),
	}
}

func TestGetUserTeamInProjectCacheHit(t *testing.T) {
	r := newCacheOnlyResolver(t)

	// Membership cached as team ID; team details cached under TeamKey
	r.Cache.Set(cache.UserTeamInProjectKey(cacheTestUserID, cacheTestProjectID), cacheTestTeamID)
	r.Cache.Set(cache.TeamKey(cacheTestTeamID), &model.Team{
		ID:        cacheTestTeamID,
		ProjectID: cacheTestProjectID,
		Name:      "Cached Team",
	})
	r.Cache.Wait()

	team, err := r.getUserTeamInProject(context.Background(), cacheTestUserID, cacheTestProjectID)
	require.NoError(t, err)
	require.NotNil(t, team)
	assert.Equal(t, cacheTestTeamID, team.ID)
	assert.Equal(t, "Cached Team", team.Name)
}

func TestGetUserTeamInProjectNegativeCacheHit(t *testing.T) {
	r := newCacheOnlyResolver(t)

	// Empty string is the cached "user has no team" result
	r.Cache.Set(cache.UserTeamInProjectKey(cacheTestUserID, cacheTestProjectID), "")
	r.Cache.Wait()

	team, err := r.getUserTeamInProject(context.Background(), cacheTestUserID, cacheTestProjectID)
	require.NoError(t, err)
	assert.Nil(t, team)
}

func TestGetUserTeamInProjectCacheMissReachesDB(t *testing.T) {
	r := newCacheOnlyResolver(t)

	// Nothing cached → the loader batch function must query, which errors
	// here because this resolver has no database.
	_, err := r.getUserTeamInProject(context.Background(), cacheTestUserID, cacheTestProjectID)
	assert.Error(t, err)
}

func TestGetUserEnrolledChallengeIDsCacheHit(t *testing.T) {
	r := newCacheOnlyResolver(t)

	enrolled := map[string]bool{"CL01ARZ3NDEKTSV4RRFFQ69G5FAV": true}
	r.Cache.Set(cache.UserEnrolledChallengesKey(cacheTestUserID, cacheTestProjectID), enrolled)
	r.Cache.Wait()

	ids, err := r.getUserEnrolledChallengeIDs(context.Background(), cacheTestUserID, cacheTestProjectID)
	require.NoError(t, err)
	assert.True(t, ids["CL01ARZ3NDEKTSV4RRFFQ69G5FAV"])
	assert.Len(t, ids, 1)
}

func TestGetUserEnrolledChallengeIDsEmptyCacheHit(t *testing.T) {
	r := newCacheOnlyResolver(t)

	// The empty enrollment set is cached too (negative caching)
	r.Cache.Set(cache.UserEnrolledChallengesKey(cacheTestUserID, cacheTestProjectID), map[string]bool{})
	r.Cache.Wait()

	ids, err := r.getUserEnrolledChallengeIDs(context.Background(), cacheTestUserID, cacheTestProjectID)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestGetUserAccessibleQuizIDsCacheHit(t *testing.T) {
	r := newCacheOnlyResolver(t)

	// The cached set is project-wide, so it answers any requested quiz subset
	r.Cache.SetWithTTL(cache.UserQuizSessionAccessKey(cacheTestUserID, cacheTestProjectID),
		map[string]bool{"QZA": true}, loaders.QuizSessionAccessTTL)
	r.Cache.Wait()

	access, err := r.getUserAccessibleQuizIDs(context.Background(), cacheTestUserID, cacheTestProjectID, []string{"QZA"})
	require.NoError(t, err)
	assert.True(t, access["QZA"])
	assert.False(t, access["QZB"])

	access, err = r.getUserAccessibleQuizIDs(context.Background(), cacheTestUserID, cacheTestProjectID, []string{"QZA", "QZB", "QZC"})
	require.NoError(t, err)
	assert.True(t, access["QZA"])
	assert.False(t, access["QZB"])
}

func TestGetUserAccessibleQuizIDsEmptyRequest(t *testing.T) {
	r := newCacheOnlyResolver(t)

	// No quiz IDs → no lookup at all
	access, err := r.getUserAccessibleQuizIDs(context.Background(), cacheTestUserID, cacheTestProjectID, nil)
	require.NoError(t, err)
	assert.Empty(t, access)
}

func TestGetUserAccessibleQuizIDsStaleShapeBypassesCache(t *testing.T) {
	r := newCacheOnlyResolver(t)

	// A cached value with the pre-loader shape fails the type assertion and
	// is treated as a miss — the batch function must re-query, which errors
	// here because this resolver has no database.
	type legacyCachedQuizAccess struct {
		Covered    map[string]bool
		Accessible map[string]bool
	}
	r.Cache.SetWithTTL(cache.UserQuizSessionAccessKey(cacheTestUserID, cacheTestProjectID),
		&legacyCachedQuizAccess{
			Covered:    map[string]bool{"QZA": true},
			Accessible: map[string]bool{"QZA": true},
		}, loaders.QuizSessionAccessTTL)
	r.Cache.Wait()

	_, err := r.getUserAccessibleQuizIDs(context.Background(), cacheTestUserID, cacheTestProjectID, []string{"QZA"})
	assert.Error(t, err)
}
