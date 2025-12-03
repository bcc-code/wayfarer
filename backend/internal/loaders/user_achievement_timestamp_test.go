package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserAchievementKey(t *testing.T) {
	key := UserAchievementKey{
		UserID:        "US01K8XV6VK9ED2GBZSQ2VDTAT8T",
		AchievementID: "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
	}

	keyStr := key.String()
	assert.NotEmpty(t, keyStr)
	assert.Contains(t, keyStr, "US01K8XV6VK9ED2GBZSQ2VDTAT8T")
	assert.Contains(t, keyStr, "AC01K8XV6VK9ED2GBZSQ2VDTAT8T")
	assert.Equal(t, "US01K8XV6VK9ED2GBZSQ2VDTAT8T:AC01K8XV6VK9ED2GBZSQ2VDTAT8T", keyStr)

	raw := key.Raw()
	assert.Equal(t, key, raw)
}

func TestUserAchievementTimestampCacheKey(t *testing.T) {
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.UserAchievementTimestampKey(userID, achievementID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, userID)
	assert.Contains(t, cacheKey, achievementID)
}

func TestUserAchievementTimestampCacheBehavior(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievedAt := time.Now()

	// Test cache set and get
	cacheKey := cache.UserAchievementTimestampKey(userID, achievementID)
	c.Set(cacheKey, &achievedAt)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "timestamp should be in cache")
	require.NotNil(t, cached)

	cachedTimestamp, ok := cached.(*time.Time)
	assert.True(t, ok, "cached value should be a *time.Time")
	require.NotNil(t, cachedTimestamp)
	assert.Equal(t, achievedAt.Unix(), cachedTimestamp.Unix())
}

func TestUserAchievementTimestampCacheNil(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT9T" // Different achievement (not achieved)

	// Test cache set and get with nil value (user hasn't achieved this)
	cacheKey := cache.UserAchievementTimestampKey(userID, achievementID)
	var nilTimestamp *time.Time = nil
	c.Set(cacheKey, nilTimestamp)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache - nil values should be cacheable
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "nil timestamp should be in cache")
	assert.Nil(t, cached, "cached value should be nil for unachieved achievement")
}

func TestUserAchievementTimestampCacheInvalidation(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievedAt := time.Now()

	cacheKey := cache.UserAchievementTimestampKey(userID, achievementID)
	c.Set(cacheKey, &achievedAt)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify it's in cache
	_, ok := c.Get(cacheKey)
	assert.True(t, ok, "timestamp should be in cache")

	// Test cache invalidation
	c.Delete(cacheKey)
	_, ok = c.Get(cacheKey)
	assert.False(t, ok, "timestamp should not be in cache after deletion")
}

func TestMultipleUserAchievementTimestampsInCache(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievements := []struct {
		id         string
		achievedAt *time.Time
	}{
		{
			id:         "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
			achievedAt: func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
		},
		{
			id:         "AC01K8XV6VK9ED2GBZSQ2VDTAT9T",
			achievedAt: func() *time.Time { t := time.Now().Add(-48 * time.Hour); return &t }(),
		},
		{
			id:         "AC01K8XV6VK9ED2GBZSQ2VDTATZZ",
			achievedAt: nil, // Not achieved
		},
	}

	// Store all timestamps in cache
	for _, a := range achievements {
		c.Set(cache.UserAchievementTimestampKey(userID, a.id), a.achievedAt)
	}
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify all timestamps can be retrieved
	for _, a := range achievements {
		cached, ok := c.Get(cache.UserAchievementTimestampKey(userID, a.id))
		assert.True(t, ok, "timestamp for achievement %s should be in cache", a.id)

		if a.achievedAt != nil {
			cachedTs, ok := cached.(*time.Time)
			assert.True(t, ok)
			require.NotNil(t, cachedTs)
			assert.Equal(t, a.achievedAt.Unix(), cachedTs.Unix())
		} else {
			assert.Nil(t, cached, "unachieved achievement should have nil timestamp in cache")
		}
	}
}

func TestDifferentUsersAchievementTimestamps(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	user1ID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	user2ID := "US01K8XV6VK9ED2GBZSQ2VDTAT9T"

	user1Time := time.Now().Add(-24 * time.Hour)
	user2Time := time.Now().Add(-48 * time.Hour)

	// Store timestamps for both users
	c.Set(cache.UserAchievementTimestampKey(user1ID, achievementID), &user1Time)
	c.Set(cache.UserAchievementTimestampKey(user2ID, achievementID), &user2Time)
	time.Sleep(10 * time.Millisecond)

	// Verify each user has their own cached timestamp
	cached1, ok := c.Get(cache.UserAchievementTimestampKey(user1ID, achievementID))
	assert.True(t, ok)
	ts1 := cached1.(*time.Time)
	assert.Equal(t, user1Time.Unix(), ts1.Unix())

	cached2, ok := c.Get(cache.UserAchievementTimestampKey(user2ID, achievementID))
	assert.True(t, ok)
	ts2 := cached2.(*time.Time)
	assert.Equal(t, user2Time.Unix(), ts2.Unix())

	// Invalidating one user's cache shouldn't affect the other
	c.Delete(cache.UserAchievementTimestampKey(user1ID, achievementID))

	_, ok = c.Get(cache.UserAchievementTimestampKey(user1ID, achievementID))
	assert.False(t, ok, "user1's cache should be invalidated")

	_, ok = c.Get(cache.UserAchievementTimestampKey(user2ID, achievementID))
	assert.True(t, ok, "user2's cache should still be valid")
}
