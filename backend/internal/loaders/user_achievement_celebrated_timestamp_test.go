package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserAchievementCelebratedTimestampCacheKey(t *testing.T) {
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.UserAchievementCelebratedTimestampKey(userID, achievementID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, userID)
	assert.Contains(t, cacheKey, achievementID)
	assert.Contains(t, cacheKey, "celebrated")

	// Ensure it differs from the achievedAt cache key
	achievedKey := cache.UserAchievementTimestampKey(userID, achievementID)
	assert.NotEqual(t, cacheKey, achievedKey)
}

func TestUserAchievementCelebratedTimestampCacheBehavior(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	celebratedAt := time.Now()

	// Test cache set and get
	cacheKey := cache.UserAchievementCelebratedTimestampKey(userID, achievementID)
	c.Set(cacheKey, &celebratedAt)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "timestamp should be in cache")
	require.NotNil(t, cached)

	cachedTimestamp, ok := cached.(*time.Time)
	assert.True(t, ok, "cached value should be a *time.Time")
	require.NotNil(t, cachedTimestamp)
	assert.Equal(t, celebratedAt.Unix(), cachedTimestamp.Unix())
}

func TestUserAchievementCelebratedTimestampCacheNil(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT9T"

	// Test cache set and get with nil value (user hasn't celebrated this)
	cacheKey := cache.UserAchievementCelebratedTimestampKey(userID, achievementID)
	var nilTimestamp *time.Time = nil
	c.Set(cacheKey, nilTimestamp)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache - nil values should be cacheable
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "nil timestamp should be in cache")
	assert.Nil(t, cached, "cached value should be nil for uncelebrated achievement")
}

func TestUserAchievementCelebratedTimestampCacheInvalidation(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	celebratedAt := time.Now()

	cacheKey := cache.UserAchievementCelebratedTimestampKey(userID, achievementID)
	c.Set(cacheKey, &celebratedAt)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify it's in cache
	_, ok := c.Get(cacheKey)
	assert.True(t, ok, "timestamp should be in cache")

	// Test cache invalidation
	c.Delete(cacheKey)
	_, ok = c.Get(cacheKey)
	assert.False(t, ok, "timestamp should not be in cache after deletion")
}

func TestCelebratedTimestampIndependentFromAchievedTimestamp(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"

	achievedAt := time.Now().Add(-24 * time.Hour)
	celebratedAt := time.Now()

	// Store both timestamps
	c.Set(cache.UserAchievementTimestampKey(userID, achievementID), &achievedAt)
	c.Set(cache.UserAchievementCelebratedTimestampKey(userID, achievementID), &celebratedAt)
	time.Sleep(10 * time.Millisecond)

	// Verify both are stored independently
	cachedAchieved, ok := c.Get(cache.UserAchievementTimestampKey(userID, achievementID))
	assert.True(t, ok)
	ts1 := cachedAchieved.(*time.Time)
	assert.Equal(t, achievedAt.Unix(), ts1.Unix())

	cachedCelebrated, ok := c.Get(cache.UserAchievementCelebratedTimestampKey(userID, achievementID))
	assert.True(t, ok)
	ts2 := cachedCelebrated.(*time.Time)
	assert.Equal(t, celebratedAt.Unix(), ts2.Unix())

	// Invalidating celebrated shouldn't affect achieved
	c.Delete(cache.UserAchievementCelebratedTimestampKey(userID, achievementID))

	_, ok = c.Get(cache.UserAchievementCelebratedTimestampKey(userID, achievementID))
	assert.False(t, ok, "celebrated cache should be invalidated")

	_, ok = c.Get(cache.UserAchievementTimestampKey(userID, achievementID))
	assert.True(t, ok, "achieved cache should still be valid")
}
