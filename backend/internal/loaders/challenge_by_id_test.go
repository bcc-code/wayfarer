package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeCacheKey(t *testing.T) {
	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.ChallengeKey(challengeID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, challengeID)
}

func TestChallengeCacheBehavior(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	publishedAt := time.Now().Add(-24 * time.Hour)
	endTime := time.Now().Add(7 * 24 * time.Hour)

	challenge := &model.Challenge{
		ID:          challengeID,
		Name:        "Test Challenge",
		Description: scalars.HTML("Test description"),
		Image:       stringPtr("https://example.com/image.png"),
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		EventID:     nil,
		URL:         stringPtr("https://example.com/challenge"),
		ButtonText:  "Start Challenge",
		PublishedAt: &scalars.DateTime{Time: publishedAt},
		EndTime:     &scalars.DateTime{Time: endTime},
	}

	// Test cache set and get
	cacheKey := cache.ChallengeKey(challengeID)
	c.Set(cacheKey, challenge)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "challenge should be in cache")
	require.NotNil(t, cached)

	cachedChallenge, ok := cached.(*model.Challenge)
	assert.True(t, ok, "cached value should be a *model.Challenge")
	require.NotNil(t, cachedChallenge)
	assert.Equal(t, challengeID, cachedChallenge.ID)
	assert.Equal(t, "Test Challenge", cachedChallenge.Name)
	assert.Equal(t, scalars.HTML("Test description"), cachedChallenge.Description)
	assert.Equal(t, "https://example.com/challenge", *cachedChallenge.URL)
	assert.Equal(t, "Start Challenge", cachedChallenge.ButtonText)
	assert.NotNil(t, cachedChallenge.EndTime)
}

func TestChallengeCacheExpiry(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	challengeID := "CL01K8XV6VK9ED2GBZSQ2VDTAT8T"
	publishedAt := time.Now().Add(-24 * time.Hour)

	challenge := &model.Challenge{
		ID:          challengeID,
		Name:        "Test Challenge",
		Description: scalars.HTML("Test description"),
		Image:       stringPtr("https://example.com/image.png"),
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		URL:         stringPtr("https://example.com/challenge"),
		ButtonText:  "Start Challenge",
		PublishedAt: &scalars.DateTime{Time: publishedAt},
	}

	cacheKey := cache.ChallengeKey(challengeID)
	c.Set(cacheKey, challenge)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify it's in cache
	_, ok := c.Get(cacheKey)
	assert.True(t, ok, "challenge should be in cache")

	// Test cache invalidation
	c.Delete(cacheKey)
	_, ok = c.Get(cacheKey)
	assert.False(t, ok, "challenge should not be in cache after deletion")
}

func TestChallengeModel(t *testing.T) {
	publishedAt := time.Now().Add(-48 * time.Hour)
	endTime := time.Now().Add(14 * 24 * time.Hour)
	eventID := "EV01K8XV6VK9ED2GBZSQ2VDTAT8T"

	challenge := &model.Challenge{
		ID:          "CL01K8XV6VK9ED2GBZSQ2VDTAT8T",
		Name:        "Daily Bible Reading",
		Description: scalars.HTML("<p>Read the daily passage</p>"),
		Image:       stringPtr("https://example.com/bible.png"),
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		EventID:     &eventID,
		URL:         stringPtr("https://example.com/bible-reading"),
		ButtonText:  "Read Now",
		PublishedAt: &scalars.DateTime{Time: publishedAt},
		EndTime:     &scalars.DateTime{Time: endTime},
	}

	assert.Equal(t, "CL01K8XV6VK9ED2GBZSQ2VDTAT8T", challenge.ID)
	assert.Equal(t, "Daily Bible Reading", challenge.Name)
	assert.Equal(t, scalars.HTML("<p>Read the daily passage</p>"), challenge.Description)
	assert.Equal(t, "https://example.com/bible-reading", *challenge.URL)
	assert.Equal(t, "Read Now", challenge.ButtonText)
	assert.NotNil(t, challenge.EventID)
	assert.Equal(t, "EV01K8XV6VK9ED2GBZSQ2VDTAT8T", *challenge.EventID)
	assert.NotNil(t, challenge.EndTime)
}

func TestChallengeModelWithoutEndTime(t *testing.T) {
	publishedAt := time.Now().Add(-24 * time.Hour)

	challenge := &model.Challenge{
		ID:          "CL01K8XV6VK9ED2GBZSQ2VDTAT8T",
		Name:        "Ongoing Challenge",
		Description: scalars.HTML("Challenge with no end time"),
		Image:       stringPtr("https://example.com/ongoing.png"),
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		EventID:     nil,
		URL:         stringPtr("https://example.com/ongoing"),
		ButtonText:  "Participate",
		PublishedAt: &scalars.DateTime{Time: publishedAt},
		EndTime:     nil,
	}

	assert.Equal(t, "CL01K8XV6VK9ED2GBZSQ2VDTAT8T", challenge.ID)
	assert.Equal(t, "Ongoing Challenge", challenge.Name)
	assert.Nil(t, challenge.EventID)
	assert.Nil(t, challenge.EndTime)
}

func TestMultipleChallengesInCache(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	publishedAt := time.Now().Add(-24 * time.Hour)
	endTime := time.Now().Add(7 * 24 * time.Hour)

	challenges := []*model.Challenge{
		{
			ID:          "CL01K8XV6VK9ED2GBZSQ2VDTAT8T",
			Name:        "Challenge 1",
			Description: scalars.HTML("First challenge"),
			Image:       stringPtr("https://example.com/1.png"),
			ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
			URL:         stringPtr("https://example.com/challenge1"),
			ButtonText:  "Start",
			PublishedAt: &scalars.DateTime{Time: publishedAt},
			EndTime:     &scalars.DateTime{Time: endTime},
		},
		{
			ID:          "CL01K8XV6VK9ED2GBZSQ2VDTAT9T",
			Name:        "Challenge 2",
			Description: scalars.HTML("Second challenge"),
			Image:       stringPtr("https://example.com/2.png"),
			ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
			URL:         stringPtr("https://example.com/challenge2"),
			ButtonText:  "Begin",
			PublishedAt: &scalars.DateTime{Time: publishedAt},
			EndTime:     nil,
		},
		{
			ID:          "CL01K8XV6VK9ED2GBZSQ2VDTATZZ",
			Name:        "Challenge 3",
			Description: scalars.HTML("Third challenge"),
			Image:       stringPtr("https://example.com/3.png"),
			ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
			URL:         stringPtr("https://example.com/challenge3"),
			ButtonText:  "Go",
			PublishedAt: &scalars.DateTime{Time: publishedAt},
			EndTime:     &scalars.DateTime{Time: endTime},
		},
	}

	// Store all challenges in cache
	for _, challenge := range challenges {
		c.Set(cache.ChallengeKey(challenge.ID), challenge)
	}
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify all challenges can be retrieved
	for _, expectedChallenge := range challenges {
		cached, ok := c.Get(cache.ChallengeKey(expectedChallenge.ID))
		assert.True(t, ok, "challenge %s should be in cache", expectedChallenge.ID)
		require.NotNil(t, cached)

		cachedChallenge, ok := cached.(*model.Challenge)
		assert.True(t, ok)
		assert.Equal(t, expectedChallenge.ID, cachedChallenge.ID)
		assert.Equal(t, expectedChallenge.Name, cachedChallenge.Name)
		assert.Equal(t, expectedChallenge.URL, cachedChallenge.URL)
	}
}
