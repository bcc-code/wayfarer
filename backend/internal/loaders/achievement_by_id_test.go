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
		ID:          achievementID,
		Name:        "Test Achievement",
		Description: "Test description",
		Image:       stringPtr("https://example.com/image.png"),
		Points:      100,
		Hidden:      false,
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
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
	assert.Equal(t, "Test description", cachedAchievement.Description)
	assert.Equal(t, 100, cachedAchievement.Points)
}

func TestAchievementCacheExpiry(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	achievementID := "AC01K8XV6VK9ED2GBZSQ2VDTAT8T"
	achievement := &model.SimpleAchievement{
		ID:          achievementID,
		Name:        "Test Achievement",
		Description: "Test description",
		Image:       stringPtr("https://example.com/image.png"),
		Points:      100,
		Hidden:      false,
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
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

func TestSimpleAchievementModel(t *testing.T) {
	achievement := &model.SimpleAchievement{
		ID:          "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
		Name:        "First Achievement",
		Description: "Complete your first challenge",
		Image:       stringPtr("https://example.com/achievement.png"),
		Points:      50,
		Hidden:      false,
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
	}

	assert.Equal(t, "AC01K8XV6VK9ED2GBZSQ2VDTAT8T", achievement.ID)
	assert.Equal(t, "First Achievement", achievement.Name)
	assert.Equal(t, "Complete your first challenge", achievement.Description)
	assert.Equal(t, 50, achievement.Points)
	assert.False(t, achievement.Hidden)
}

func TestContentAchievementModel(t *testing.T) {
	achievement := &model.ContentAchievement{
		ID:          "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
		Name:        "Content Achievement",
		Description: "Complete all required content",
		Image:       stringPtr("https://example.com/content.png"),
		Points:      100,
		Hidden:      false,
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		TotalItems:  2,
		// Items will be populated by resolver
	}

	assert.Equal(t, "AC01K8XV6VK9ED2GBZSQ2VDTAT8T", achievement.ID)
	assert.Equal(t, "Content Achievement", achievement.Name)
	assert.Equal(t, 2, achievement.TotalItems)
}

func TestStreakAchievementModel(t *testing.T) {
	achievement := &model.StreakAchievement{
		ID:           "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
		Name:         "Streak Achievement",
		Description:  "Maintain a 7-day streak",
		Image:        stringPtr("https://example.com/streak.png"),
		Points:       200,
		Hidden:       false,
		ProjectID:    "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		StreakID:     "SK01K8XV6VK9ED2GBZSQ2VDTAT8T",
		NeededStreak: 7,
	}

	assert.Equal(t, "AC01K8XV6VK9ED2GBZSQ2VDTAT8T", achievement.ID)
	assert.Equal(t, "Streak Achievement", achievement.Name)
	assert.Equal(t, 7, achievement.NeededStreak)
	assert.Equal(t, "SK01K8XV6VK9ED2GBZSQ2VDTAT8T", achievement.StreakID)
	assert.Equal(t, 200, achievement.Points)
}

func TestMultipleAchievementsInCache(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	achievements := []model.Achievement{
		&model.SimpleAchievement{
			ID:          "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
			Name:        "Simple Achievement",
			Description: "First achievement",
			Image:       stringPtr("https://example.com/1.png"),
			Points:      50,
			Hidden:      false,
			ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		},
		&model.ContentAchievement{
			ID:          "AC01K8XV6VK9ED2GBZSQ2VDTAT9T",
			Name:        "Content Achievement",
			Description: "Second achievement",
			Image:       stringPtr("https://example.com/2.png"),
			Points:      100,
			Hidden:      false,
			ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
			TotalItems:  3,
		},
		&model.StreakAchievement{
			ID:           "AC01K8XV6VK9ED2GBZSQ2VDTATZZ",
			Name:         "Streak Achievement",
			Description:  "Third achievement",
			Image:        stringPtr("https://example.com/3.png"),
			Points:       150,
			Hidden:       false,
			ProjectID:    "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
			StreakID:     "SK01K8XV6VK9ED2GBZSQ2VDTAT8T",
			NeededStreak: 7,
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

func TestHiddenAchievement(t *testing.T) {
	achievement := &model.SimpleAchievement{
		ID:          "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
		Name:        "Secret Achievement",
		Description: "Hidden achievement",
		Image:       stringPtr("https://example.com/secret.png"),
		Points:      500,
		Hidden:      true,
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
	}

	assert.True(t, achievement.Hidden)
	assert.Equal(t, 500, achievement.Points)
}
