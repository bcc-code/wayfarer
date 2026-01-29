package loaders

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"time"
)

func TestUserContentProgressCacheKey(t *testing.T) {
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.UserContentProgressKey(userID, achievementID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, userID)
	assert.Contains(t, cacheKey, achievementID)
}

func TestUserContentProgressCacheBehavior(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"

	now := time.Now()
	progress := []*sqlc.UserContentProgress{
		{
			UserID:            userID,
			AchievementID:     achievementID,
			ExternalContentID: "EC01K8XV6VK9ED2GBZSQ2VDTAT8T",
			CompletedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			UserID:            userID,
			AchievementID:     achievementID,
			ExternalContentID: "EC01K8XV6VK9ED2GBZSQ2VDTAT9T",
			CompletedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	cacheKey := cache.UserContentProgressKey(userID, achievementID)
	c.Set(cacheKey, progress)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "progress should be in cache")
	require.NotNil(t, cached)

	cachedProgress, ok := cached.([]*sqlc.UserContentProgress)
	assert.True(t, ok, "cached value should be []*sqlc.UserContentProgress")
	require.Len(t, cachedProgress, 2)
	assert.Equal(t, "EC01K8XV6VK9ED2GBZSQ2VDTAT8T", cachedProgress[0].ExternalContentID)
	assert.Equal(t, "EC01K8XV6VK9ED2GBZSQ2VDTAT9T", cachedProgress[1].ExternalContentID)
}

func TestUserContentProgressCacheEmpty(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT9T" // Different achievement (no progress)

	emptyProgress := []*sqlc.UserContentProgress{}
	cacheKey := cache.UserContentProgressKey(userID, achievementID)
	c.Set(cacheKey, emptyProgress)
	time.Sleep(10 * time.Millisecond)

	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "empty progress should be in cache")
	require.NotNil(t, cached)

	cachedProgress, ok := cached.([]*sqlc.UserContentProgress)
	assert.True(t, ok, "cached value should be []*sqlc.UserContentProgress")
	assert.Empty(t, cachedProgress)
}

func TestUserContentProgressCacheInvalidation(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"

	now := time.Now()
	progress := []*sqlc.UserContentProgress{
		{
			UserID:            userID,
			AchievementID:     achievementID,
			ExternalContentID: "EC01K8XV6VK9ED2GBZSQ2VDTAT8T",
			CompletedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	cacheKey := cache.UserContentProgressKey(userID, achievementID)
	c.Set(cacheKey, progress)
	time.Sleep(10 * time.Millisecond)

	_, ok := c.Get(cacheKey)
	assert.True(t, ok, "progress should be in cache")

	c.Delete(cacheKey)
	_, ok = c.Get(cacheKey)
	assert.False(t, ok, "progress should not be in cache after deletion")
}

func TestDifferentUsersContentProgress(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	user1ID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	user2ID := "US01K8XV6VK9ED2GBZSQ2VDTAT9T"

	now := time.Now()
	user1Progress := []*sqlc.UserContentProgress{
		{
			UserID:            user1ID,
			AchievementID:     achievementID,
			ExternalContentID: "EC01K8XV6VK9ED2GBZSQ2VDTAT8T",
			CompletedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		},
	}
	user2Progress := []*sqlc.UserContentProgress{
		{
			UserID:            user2ID,
			AchievementID:     achievementID,
			ExternalContentID: "EC01K8XV6VK9ED2GBZSQ2VDTAT8T",
			CompletedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			UserID:            user2ID,
			AchievementID:     achievementID,
			ExternalContentID: "EC01K8XV6VK9ED2GBZSQ2VDTAT9T",
			CompletedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	c.Set(cache.UserContentProgressKey(user1ID, achievementID), user1Progress)
	c.Set(cache.UserContentProgressKey(user2ID, achievementID), user2Progress)
	time.Sleep(10 * time.Millisecond)

	cached1, ok := c.Get(cache.UserContentProgressKey(user1ID, achievementID))
	assert.True(t, ok)
	progress1 := cached1.([]*sqlc.UserContentProgress)
	assert.Len(t, progress1, 1)

	cached2, ok := c.Get(cache.UserContentProgressKey(user2ID, achievementID))
	assert.True(t, ok)
	progress2 := cached2.([]*sqlc.UserContentProgress)
	assert.Len(t, progress2, 2)

	// Invalidating one user's cache shouldn't affect the other
	c.Delete(cache.UserContentProgressKey(user1ID, achievementID))

	_, ok = c.Get(cache.UserContentProgressKey(user1ID, achievementID))
	assert.False(t, ok, "user1's cache should be invalidated")

	_, ok = c.Get(cache.UserContentProgressKey(user2ID, achievementID))
	assert.True(t, ok, "user2's cache should still be valid")
}

func TestUserContentProgressKeyString(t *testing.T) {
	key := UserAchievementKey{
		UserID:        "US01K8XV6VK9ED2GBZSQ2VDTAT8T",
		AchievementID: "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
	}

	// Verify the key string format is consistent (used for grouping DB results)
	assert.Equal(t, "US01K8XV6VK9ED2GBZSQ2VDTAT8T:AC01K8XV6VK9ED2GBZSQ2VDTAT8T", key.String())
}
