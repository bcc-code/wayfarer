package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/cache"
)

// TestUserProjectPointsKey_Format verifies the cache key format
func TestUserProjectPointsKey_Format(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		projectID string
		expected  string
	}{
		{
			name:      "standard IDs",
			userID:    "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			projectID: "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
			expected:  "userprojectpoints:US01ARZ3NDEKTSV4RRFFQ69G5FAV:PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		},
		{
			name:      "different IDs",
			userID:    "US01DIFFERENT123456789ABCDEF",
			projectID: "PR01ANOTHER789012345678GHIJK",
			expected:  "userprojectpoints:US01DIFFERENT123456789ABCDEF:PR01ANOTHER789012345678GHIJK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.UserProjectPointsKey(tt.userID, tt.projectID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUserProjectPointsCache_SetAndGet verifies cache set and get operations
func TestUserProjectPointsCache_SetAndGet(t *testing.T) {
	testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	points := 150

	cacheKey := cache.UserProjectPointsKey(userID, projectID)

	// Initially not in cache
	_, ok := testCache.Get(cacheKey)
	assert.False(t, ok, "key should not be in cache initially")

	// Set value
	testCache.Set(cacheKey, points)
	testCache.Wait()

	// Get value
	cached, ok := testCache.Get(cacheKey)
	require.True(t, ok, "key should be in cache after set")
	cachedPoints, ok := cached.(int)
	require.True(t, ok, "cached value should be an int")
	assert.Equal(t, points, cachedPoints)
}

// TestUserProjectPointsCache_Invalidation verifies cache invalidation on user change
func TestUserProjectPointsCache_Invalidation(t *testing.T) {
	testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	points := 200

	cacheKey := cache.UserProjectPointsKey(userID, projectID)

	// Set value
	testCache.Set(cacheKey, points)
	testCache.Wait()

	// Verify it's in cache
	_, ok := testCache.Get(cacheKey)
	require.True(t, ok, "key should be in cache after set")

	// Invalidate user cache
	testCache.InvalidateUser(userID)

	// Verify it's removed
	_, ok = testCache.Get(cacheKey)
	assert.False(t, ok, "key should be removed from cache after user invalidation")
}

// TestUserProjectPointsCache_InvalidationDoesNotAffectOtherUsers verifies isolation
func TestUserProjectPointsCache_InvalidationDoesNotAffectOtherUsers(t *testing.T) {
	testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	user1ID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	user2ID := "US02ARZ3NDEKTSV4RRFFQ69G5FAV"
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"

	cacheKey1 := cache.UserProjectPointsKey(user1ID, projectID)
	cacheKey2 := cache.UserProjectPointsKey(user2ID, projectID)

	// Set values for both users
	testCache.Set(cacheKey1, 100)
	testCache.Set(cacheKey2, 200)
	testCache.Wait()

	// Invalidate only user1
	testCache.InvalidateUser(user1ID)

	// Verify user1's cache is removed
	_, ok := testCache.Get(cacheKey1)
	assert.False(t, ok, "user1's key should be removed from cache")

	// Verify user2's cache is still there
	cached, ok := testCache.Get(cacheKey2)
	require.True(t, ok, "user2's key should still be in cache")
	cachedPoints, ok := cached.(int)
	require.True(t, ok, "cached value should be an int")
	assert.Equal(t, 200, cachedPoints)
}

// TestUserProjectPointsCache_MultipleProjects verifies caching across projects
func TestUserProjectPointsCache_MultipleProjects(t *testing.T) {
	testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	project1ID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	project2ID := "PR02ARZ3NDEKTSV4RRFFQ69G5FAV"

	cacheKey1 := cache.UserProjectPointsKey(userID, project1ID)
	cacheKey2 := cache.UserProjectPointsKey(userID, project2ID)

	// Set different points for different projects
	testCache.Set(cacheKey1, 100)
	testCache.Set(cacheKey2, 250)
	testCache.Wait()

	// Verify both are cached correctly
	cached1, ok := testCache.Get(cacheKey1)
	require.True(t, ok)
	assert.Equal(t, 100, cached1.(int))

	cached2, ok := testCache.Get(cacheKey2)
	require.True(t, ok)
	assert.Equal(t, 250, cached2.(int))

	// Invalidating user should clear both
	testCache.InvalidateUser(userID)

	_, ok = testCache.Get(cacheKey1)
	assert.False(t, ok, "project1 cache should be cleared")

	_, ok = testCache.Get(cacheKey2)
	assert.False(t, ok, "project2 cache should be cleared")
}
