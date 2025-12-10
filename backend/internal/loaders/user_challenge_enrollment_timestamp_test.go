package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserChallengeEnrollmentTimestampCacheKey(t *testing.T) {
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.UserChallengeEnrollmentKey(userID, challengeID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, userID)
	assert.Contains(t, cacheKey, challengeID)
}

func TestUserChallengeEnrollmentTimestampCacheBehavior(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	enrolledAt := time.Now()

	// Test cache set and get
	cacheKey := cache.UserChallengeEnrollmentKey(userID, challengeID)
	c.Set(cacheKey, &enrolledAt)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "timestamp should be in cache")
	require.NotNil(t, cached)

	cachedTimestamp, ok := cached.(*time.Time)
	assert.True(t, ok, "cached value should be a *time.Time")
	require.NotNil(t, cachedTimestamp)
	assert.Equal(t, enrolledAt.Unix(), cachedTimestamp.Unix())
}

func TestUserChallengeEnrollmentTimestampCacheNil(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT9T" // Different challenge (not enrolled)

	// Test cache set and get with nil value (user hasn't enrolled in this)
	cacheKey := cache.UserChallengeEnrollmentKey(userID, challengeID)
	var nilTimestamp *time.Time = nil
	c.Set(cacheKey, nilTimestamp)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache - nil values should be cacheable
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "nil timestamp should be in cache")
	assert.Nil(t, cached, "cached value should be nil for unenrolled challenge")
}

func TestUserChallengeEnrollmentTimestampCacheInvalidation(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	enrolledAt := time.Now()

	cacheKey := cache.UserChallengeEnrollmentKey(userID, challengeID)
	c.Set(cacheKey, &enrolledAt)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify it's in cache
	_, ok := c.Get(cacheKey)
	assert.True(t, ok, "timestamp should be in cache")

	// Test cache invalidation
	c.Delete(cacheKey)
	_, ok = c.Get(cacheKey)
	assert.False(t, ok, "timestamp should not be in cache after deletion")
}

func TestMultipleUserChallengeEnrollmentTimestampsInCache(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	challenges := []struct {
		id         string
		enrolledAt *time.Time
	}{
		{
			id:         "CL01K8XV6VK9ED2GBZSQ2VDTAT8T",
			enrolledAt: func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
		},
		{
			id:         "CL01K8XV6VK9ED2GBZSQ2VDTAT9T",
			enrolledAt: func() *time.Time { t := time.Now().Add(-48 * time.Hour); return &t }(),
		},
		{
			id:         "CL01K8XV6VK9ED2GBZSQ2VDTATZZ",
			enrolledAt: nil, // Not enrolled
		},
	}

	// Store all timestamps in cache
	for _, ch := range challenges {
		c.Set(cache.UserChallengeEnrollmentKey(userID, ch.id), ch.enrolledAt)
	}
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify all timestamps can be retrieved
	for _, ch := range challenges {
		cached, ok := c.Get(cache.UserChallengeEnrollmentKey(userID, ch.id))
		assert.True(t, ok, "timestamp for challenge %s should be in cache", ch.id)

		if ch.enrolledAt != nil {
			cachedTs, ok := cached.(*time.Time)
			assert.True(t, ok)
			require.NotNil(t, cachedTs)
			assert.Equal(t, ch.enrolledAt.Unix(), cachedTs.Unix())
		} else {
			assert.Nil(t, cached, "unenrolled challenge should have nil timestamp in cache")
		}
	}
}

func TestDifferentUsersChallengeEnrollmentTimestamps(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	user1ID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	user2ID := "US01K8XV6VK9ED2GBZSQ2VDTAT9T"

	user1Time := time.Now().Add(-24 * time.Hour)
	user2Time := time.Now().Add(-48 * time.Hour)

	// Store timestamps for both users
	c.Set(cache.UserChallengeEnrollmentKey(user1ID, challengeID), &user1Time)
	c.Set(cache.UserChallengeEnrollmentKey(user2ID, challengeID), &user2Time)
	time.Sleep(10 * time.Millisecond)

	// Verify each user has their own cached timestamp
	cached1, ok := c.Get(cache.UserChallengeEnrollmentKey(user1ID, challengeID))
	assert.True(t, ok)
	ts1 := cached1.(*time.Time)
	assert.Equal(t, user1Time.Unix(), ts1.Unix())

	cached2, ok := c.Get(cache.UserChallengeEnrollmentKey(user2ID, challengeID))
	assert.True(t, ok)
	ts2 := cached2.(*time.Time)
	assert.Equal(t, user2Time.Unix(), ts2.Unix())

	// Invalidating one user's cache shouldn't affect the other
	c.Delete(cache.UserChallengeEnrollmentKey(user1ID, challengeID))

	_, ok = c.Get(cache.UserChallengeEnrollmentKey(user1ID, challengeID))
	assert.False(t, ok, "user1's cache should be invalidated")

	_, ok = c.Get(cache.UserChallengeEnrollmentKey(user2ID, challengeID))
	assert.True(t, ok, "user2's cache should still be valid")
}
