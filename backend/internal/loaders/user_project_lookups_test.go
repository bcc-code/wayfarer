package loaders

import (
	"context"
	"errors"
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCache(t *testing.T) *cache.CacheWithRegistry {
	t.Helper()
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	return c
}

func TestBatchUserProjectLookupCacheHitSkipsFetch(t *testing.T) {
	c := newTestCache(t)

	key := UserProjectKey{UserID: "US01K8XV6VK9ED2GBZSQ2VDTAT8T", ProjectID: "PR01K8XV6VK9ED2GBZSQ2VDTAT8T"}
	c.Set(cache.UserTeamInProjectKey(key.UserID, key.ProjectID), "TM01K8XV6VK9ED2GBZSQ2VDTAT8T")
	c.Wait()

	results := batchUserProjectLookup(context.Background(), c, []UserProjectKey{key},
		cache.UserTeamInProjectKey,
		func(ctx context.Context, projectID string, userIDs []string) (map[string]string, error) {
			t.Fatal("fetch must not be called on a cache hit")
			return nil, nil
		},
		func() string { return "" },
		func(k, v string) { c.Set(k, v) },
	)

	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	assert.Equal(t, "TM01K8XV6VK9ED2GBZSQ2VDTAT8T", results[0].Data)
}

func TestBatchUserProjectLookupGroupsByProjectAndCachesNegatives(t *testing.T) {
	c := newTestCache(t)

	keys := []UserProjectKey{
		{UserID: "US01AAAAAAAAAAAAAAAAAAAAAAAA", ProjectID: "PR01AAAAAAAAAAAAAAAAAAAAAAAA"},
		{UserID: "US01BBBBBBBBBBBBBBBBBBBBBBBB", ProjectID: "PR01AAAAAAAAAAAAAAAAAAAAAAAA"},
		{UserID: "US01AAAAAAAAAAAAAAAAAAAAAAAA", ProjectID: "PR01BBBBBBBBBBBBBBBBBBBBBBBB"},
		// Duplicate key must not produce a duplicate fetch entry
		{UserID: "US01AAAAAAAAAAAAAAAAAAAAAAAA", ProjectID: "PR01AAAAAAAAAAAAAAAAAAAAAAAA"},
	}

	fetchCalls := map[string][]string{}
	results := batchUserProjectLookup(context.Background(), c, keys,
		cache.UserTeamInProjectKey,
		func(ctx context.Context, projectID string, userIDs []string) (map[string]string, error) {
			fetchCalls[projectID] = userIDs
			if projectID == "PR01AAAAAAAAAAAAAAAAAAAAAAAA" {
				return map[string]string{"US01AAAAAAAAAAAAAAAAAAAAAAAA": "TM01AAAAAAAAAAAAAAAAAAAAAAAA"}, nil
			}
			return map[string]string{}, nil
		},
		func() string { return "" },
		func(k, v string) { c.Set(k, v) },
	)

	require.Len(t, results, 4)
	for _, r := range results {
		require.NoError(t, r.Error)
	}
	// One grouped fetch per distinct project, with deduped user IDs
	require.Len(t, fetchCalls, 2)
	assert.ElementsMatch(t, []string{"US01AAAAAAAAAAAAAAAAAAAAAAAA", "US01BBBBBBBBBBBBBBBBBBBBBBBB"},
		fetchCalls["PR01AAAAAAAAAAAAAAAAAAAAAAAA"])
	assert.ElementsMatch(t, []string{"US01AAAAAAAAAAAAAAAAAAAAAAAA"},
		fetchCalls["PR01BBBBBBBBBBBBBBBBBBBBBBBB"])

	// Results in key order, missing pairs get the empty value
	assert.Equal(t, "TM01AAAAAAAAAAAAAAAAAAAAAAAA", results[0].Data)
	assert.Equal(t, "", results[1].Data)
	assert.Equal(t, "", results[2].Data)
	assert.Equal(t, "TM01AAAAAAAAAAAAAAAAAAAAAAAA", results[3].Data)

	// Negative result was cached too
	c.Wait()
	cached, ok := c.Get(cache.UserTeamInProjectKey("US01BBBBBBBBBBBBBBBBBBBBBBBB", "PR01AAAAAAAAAAAAAAAAAAAAAAAA"))
	require.True(t, ok, "negative result should be cached")
	assert.Equal(t, "", cached)
}

func TestBatchUserProjectLookupFetchErrorScopedToProject(t *testing.T) {
	c := newTestCache(t)

	keys := []UserProjectKey{
		{UserID: "US01AAAAAAAAAAAAAAAAAAAAAAAA", ProjectID: "PR01FAILFAILFAILFAILFAILFAIL"},
		{UserID: "US01AAAAAAAAAAAAAAAAAAAAAAAA", ProjectID: "PR01OKOKOKOKOKOKOKOKOKOKOKOK"},
	}

	fetchErr := errors.New("db down")
	results := batchUserProjectLookup(context.Background(), c, keys,
		cache.UserEnrolledChallengesKey,
		func(ctx context.Context, projectID string, userIDs []string) (map[string]map[string]bool, error) {
			if projectID == "PR01FAILFAILFAILFAILFAILFAIL" {
				return nil, fetchErr
			}
			return map[string]map[string]bool{
				"US01AAAAAAAAAAAAAAAAAAAAAAAA": {"CL01AAAAAAAAAAAAAAAAAAAAAAAA": true},
			}, nil
		},
		func() map[string]bool { return map[string]bool{} },
		func(k string, v map[string]bool) { c.Set(k, v) },
	)

	require.Len(t, results, 2)
	assert.ErrorIs(t, results[0].Error, fetchErr)
	require.NoError(t, results[1].Error)
	assert.True(t, results[1].Data["CL01AAAAAAAAAAAAAAAAAAAAAAAA"])

	// Errored keys must not be cached
	c.Wait()
	_, ok := c.Get(cache.UserEnrolledChallengesKey("US01AAAAAAAAAAAAAAAAAAAAAAAA", "PR01FAILFAILFAILFAILFAILFAIL"))
	assert.False(t, ok)
}

func TestBatchUserProjectLookupStaleCacheShapeIsMiss(t *testing.T) {
	c := newTestCache(t)

	key := UserProjectKey{UserID: "US01AAAAAAAAAAAAAAAAAAAAAAAA", ProjectID: "PR01AAAAAAAAAAAAAAAAAAAAAAAA"}
	// Simulate a pre-refactor cache entry with a different value shape
	type legacyShape struct{ Covered, Accessible map[string]bool }
	c.SetWithTTL(cache.UserQuizSessionAccessKey(key.UserID, key.ProjectID), &legacyShape{}, QuizSessionAccessTTL)
	c.Wait()

	fetched := false
	results := batchUserProjectLookup(context.Background(), c, []UserProjectKey{key},
		cache.UserQuizSessionAccessKey,
		func(ctx context.Context, projectID string, userIDs []string) (map[string]map[string]bool, error) {
			fetched = true
			return map[string]map[string]bool{key.UserID: {"QZ01AAAAAAAAAAAAAAAAAAAAAAAA": true}}, nil
		},
		func() map[string]bool { return map[string]bool{} },
		func(k string, v map[string]bool) { c.SetWithTTL(k, v, QuizSessionAccessTTL) },
	)

	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	assert.True(t, fetched, "stale-shaped cache entry must be treated as a miss")
	assert.True(t, results[0].Data["QZ01AAAAAAAAAAAAAAAAAAAAAAAA"])
}
