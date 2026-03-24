package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stringPtr(s string) *string {
	return &s
}

func TestAchievementCacheKey(t *testing.T) {
	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.AchievementKey(achievementID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, achievementID)
}

func TestAchievementCacheBehavior(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievement := &model.SimpleAchievement{
		ID:                   achievementID,
		Name:                 "Test Achievement",
		DescriptionPending:   "Test description pending",
		DescriptionCompleted: "Test description completed",
		ImagePending:         "https://example.com/image-pending.png",
		ImageCompleted:       "https://example.com/image-completed.png",
		Points:               100,
		Hidden:               false,
		ProjectID:            "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
	}

	// Test cache set and get
	cacheKey := cache.AchievementKey(achievementID)
	c.Set(cacheKey, achievement)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "achievement should be in cache")
	require.NotNil(t, cached)

	cachedAchievement, ok := cached.(*model.SimpleAchievement)
	assert.True(t, ok, "cached value should be a *model.SimpleAchievement")
	require.NotNil(t, cachedAchievement)
	assert.Equal(t, achievementID, cachedAchievement.ID)
	assert.Equal(t, "Test Achievement", cachedAchievement.Name)
	assert.Equal(t, "Test description pending", cachedAchievement.DescriptionPending)
	assert.Equal(t, "Test description completed", cachedAchievement.DescriptionCompleted)
	assert.Equal(t, 100, cachedAchievement.Points)
}

func TestAchievementCacheExpiry(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievement := &model.SimpleAchievement{
		ID:                   achievementID,
		Name:                 "Test Achievement",
		DescriptionPending:   "Test description pending",
		DescriptionCompleted: "Test description completed",
		ImagePending:         "https://example.com/image-pending.png",
		ImageCompleted:       "https://example.com/image-completed.png",
		Points:               100,
		Hidden:               false,
		ProjectID:            "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
	}

	cacheKey := cache.AchievementKey(achievementID)
	c.Set(cacheKey, achievement)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify it's in cache
	_, ok := c.Get(cacheKey)
	assert.True(t, ok, "achievement should be in cache")

	// Test cache invalidation
	c.Delete(cacheKey)
	_, ok = c.Get(cacheKey)
	assert.False(t, ok, "achievement should not be in cache after deletion")
}

func TestMultipleAchievementsInCache(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	achievements := []model.Achievement{
		&model.SimpleAchievement{
			ID:                   "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
			Name:                 "Simple Achievement",
			DescriptionPending:   "First achievement pending",
			DescriptionCompleted: "First achievement completed",
			ImagePending:         "https://example.com/1-pending.png",
			ImageCompleted:       "https://example.com/1-completed.png",
			Points:               50,
			Hidden:               false,
			ProjectID:            "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		},
		&model.ContentAchievement{
			ID:                   "AC01K8XV6VK9ED2GBZSQ2VDTAT9T",
			Name:                 "Content Achievement",
			DescriptionPending:   "Second achievement pending",
			DescriptionCompleted: "Second achievement completed",
			ImagePending:         "https://example.com/2-pending.png",
			ImageCompleted:       "https://example.com/2-completed.png",
			Points:               100,
			Hidden:               false,
			ProjectID:            "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
			TotalItems:           3,
		},
		&model.StreakAchievement{
			ID:                   "AC01K8XV6VK9ED2GBZSQ2VDTATZZ",
			Name:                 "Streak Achievement",
			DescriptionPending:   "Third achievement pending",
			DescriptionCompleted: "Third achievement completed",
			ImagePending:         "https://example.com/3-pending.png",
			ImageCompleted:       "https://example.com/3-completed.png",
			Points:               150,
			Hidden:               false,
			ProjectID:            "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
			TotalItems:           0,
		},
	}

	// Store all achievements in cache
	for _, achievement := range achievements {
		c.Set(cache.AchievementKey(achievement.GetID()), achievement)
	}
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify all achievements can be retrieved
	for _, expectedAchievement := range achievements {
		cached, ok := c.Get(cache.AchievementKey(expectedAchievement.GetID()))
		assert.True(t, ok, "achievement %s should be in cache", expectedAchievement.GetID())
		require.NotNil(t, cached)

		cachedAchievement, ok := cached.(model.Achievement)
		assert.True(t, ok)
		assert.Equal(t, expectedAchievement.GetID(), cachedAchievement.GetID())
		assert.Equal(t, expectedAchievement.GetName(), cachedAchievement.GetName())
	}
}
