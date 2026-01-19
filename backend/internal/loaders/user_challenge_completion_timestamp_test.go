package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserChallengeKey(t *testing.T) {
	key := UserChallengeKey{
		UserID:      "US01K8XV6VK9ED2GBZSQ2VDTAT8T",
		ChallengeID: "CL01K8XV6VK9ED2GBZSQ2VDTAT8T",
	}

	keyStr := key.String()
	assert.NotEmpty(t, keyStr)
	assert.Contains(t, keyStr, "US01K8XV6VK9ED2GBZSQ2VDTAT8T")
	assert.Contains(t, keyStr, "CL01K8XV6VK9ED2GBZSQ2VDTAT8T")
	assert.Equal(t, "US01K8XV6VK9ED2GBZSQ2VDTAT8T:CL01K8XV6VK9ED2GBZSQ2VDTAT8T", keyStr)
}

func TestUserChallengeCompletionTimestampCacheKey(t *testing.T) {
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.UserChallengeCompletionKey(userID, challengeID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, userID)
	assert.Contains(t, cacheKey, challengeID)
}

func TestUserChallengeCompletionTimestampCacheBehavior(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	completedAt := time.Now()

	// Test cache set and get
	cacheKey := cache.UserChallengeCompletionKey(userID, challengeID)
	c.Set(cacheKey, &completedAt)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "timestamp should be in cache")
	require.NotNil(t, cached)

	cachedTimestamp, ok := cached.(*time.Time)
	assert.True(t, ok, "cached value should be a *time.Time")
	require.NotNil(t, cachedTimestamp)
	assert.Equal(t, completedAt.Unix(), cachedTimestamp.Unix())
}

func TestUserChallengeCompletionTimestampCacheNil(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT9T" // Different challenge (not completed)

	// Test cache set and get with nil value (user hasn't completed this)
	cacheKey := cache.UserChallengeCompletionKey(userID, challengeID)
	var nilTimestamp *time.Time = nil
	c.Set(cacheKey, nilTimestamp)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache - nil values should be cacheable
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "nil timestamp should be in cache")
	assert.Nil(t, cached, "cached value should be nil for uncompleted challenge")
}

func TestUserChallengeCompletionTimestampCacheInvalidation(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	completedAt := time.Now()

	cacheKey := cache.UserChallengeCompletionKey(userID, challengeID)
	c.Set(cacheKey, &completedAt)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify it's in cache
	_, ok := c.Get(cacheKey)
	assert.True(t, ok, "timestamp should be in cache")

	// Test cache invalidation
	c.Delete(cacheKey)
	_, ok = c.Get(cacheKey)
	assert.False(t, ok, "timestamp should not be in cache after deletion")
}

func TestMultipleUserChallengeCompletionTimestampsInCache(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challenges := []struct {
		id          string
		completedAt *time.Time
	}{
		{
			id:          "CL01K8XV6VK9ED2GBZSQ2VDTAT8T",
			completedAt: func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
		},
		{
			id:          "CL01K8XV6VK9ED2GBZSQ2VDTAT9T",
			completedAt: func() *time.Time { t := time.Now().Add(-48 * time.Hour); return &t }(),
		},
		{
			id:          "CL01K8XV6VK9ED2GBZSQ2VDTATZZ",
			completedAt: nil, // Not completed
		},
	}

	// Store all timestamps in cache
	for _, ch := range challenges {
		c.Set(cache.UserChallengeCompletionKey(userID, ch.id), ch.completedAt)
	}
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify all timestamps can be retrieved
	for _, ch := range challenges {
		cached, ok := c.Get(cache.UserChallengeCompletionKey(userID, ch.id))
		assert.True(t, ok, "timestamp for challenge %s should be in cache", ch.id)

		if ch.completedAt != nil {
			cachedTs, ok := cached.(*time.Time)
			assert.True(t, ok)
			require.NotNil(t, cachedTs)
			assert.Equal(t, ch.completedAt.Unix(), cachedTs.Unix())
		} else {
			assert.Nil(t, cached, "uncompleted challenge should have nil timestamp in cache")
		}
	}
}

func TestDifferentUsersChallengeCompletionTimestamps(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	user1ID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	user2ID := "US01K8XV6VK9ED2GBZSQ2VDTAT9T"

	user1Time := time.Now().Add(-24 * time.Hour)
	user2Time := time.Now().Add(-48 * time.Hour)

	// Store timestamps for both users
	c.Set(cache.UserChallengeCompletionKey(user1ID, challengeID), &user1Time)
	c.Set(cache.UserChallengeCompletionKey(user2ID, challengeID), &user2Time)
	time.Sleep(10 * time.Millisecond)

	// Verify each user has their own cached timestamp
	cached1, ok := c.Get(cache.UserChallengeCompletionKey(user1ID, challengeID))
	assert.True(t, ok)
	ts1 := cached1.(*time.Time)
	assert.Equal(t, user1Time.Unix(), ts1.Unix())

	cached2, ok := c.Get(cache.UserChallengeCompletionKey(user2ID, challengeID))
	assert.True(t, ok)
	ts2 := cached2.(*time.Time)
	assert.Equal(t, user2Time.Unix(), ts2.Unix())

	// Invalidating one user's cache shouldn't affect the other
	c.Delete(cache.UserChallengeCompletionKey(user1ID, challengeID))

	_, ok = c.Get(cache.UserChallengeCompletionKey(user1ID, challengeID))
	assert.False(t, ok, "user1's cache should be invalidated")

	_, ok = c.Get(cache.UserChallengeCompletionKey(user2ID, challengeID))
	assert.True(t, ok, "user2's cache should still be valid")
}
