package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamCacheKey(t *testing.T) {
	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.TeamKey(teamID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, teamID)
}

func TestTeamCacheBehavior(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	team := &model.Team{
		ID:          teamID,
		Name:        "Test Team",
		Description: "Test description",
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
	}

	// Test cache set and get
	cacheKey := cache.TeamKey(teamID)
	c.Set(cacheKey, team)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "team should be in cache")
	require.NotNil(t, cached)

	cachedTeam, ok := cached.(*model.Team)
	assert.True(t, ok, "cached value should be a *model.Team")
	require.NotNil(t, cachedTeam)
	assert.Equal(t, teamID, cachedTeam.ID)
	assert.Equal(t, "Test Team", cachedTeam.Name)
	assert.Equal(t, "Test description", cachedTeam.Description)
	assert.Equal(t, "PR01K8XV6J9H7BAEV49ZFVYS8R1K", cachedTeam.ProjectID)
}

func TestTeamCacheExpiry(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	team := &model.Team{
		ID:          teamID,
		Name:        "Test Team",
		Description: "Test description",
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
	}

	cacheKey := cache.TeamKey(teamID)
	c.Set(cacheKey, team)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify it's in cache
	_, ok := c.Get(cacheKey)
	assert.True(t, ok, "team should be in cache")

	// Test cache invalidation
	c.Delete(cacheKey)
	_, ok = c.Get(cacheKey)
	assert.False(t, ok, "team should not be in cache after deletion")
}

func TestTeamModel(t *testing.T) {
	team := &model.Team{
		ID:          "TM01K8XV6VK9ED2GBZSQ2VDTAT8T",
		Name:        "Team Alpha",
		Description: "First team description",
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
	}

	assert.Equal(t, "TM01K8XV6VK9ED2GBZSQ2VDTAT8T", team.ID)
	assert.Equal(t, "Team Alpha", team.Name)
	assert.Equal(t, "First team description", team.Description)
	assert.Equal(t, "PR01K8XV6J9H7BAEV49ZFVYS8R1K", team.ProjectID)
}

func TestTeamModelWithEmptyDescription(t *testing.T) {
	team := &model.Team{
		ID:          "TM01K8XV6VK9ED2GBZSQ2VDTAT8T",
		Name:        "Team Alpha",
		Description: "",
		ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
	}

	assert.Equal(t, "", team.Description)
	assert.NotNil(t, team.Description) // Should be empty string, not nil
}

func TestMultipleTeamsInCache(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	teams := []*model.Team{
		{
			ID:          "TM01K8XV6VK9ED2GBZSQ2VDTAT8T",
			Name:        "Team Alpha",
			Description: "First team",
			ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		},
		{
			ID:          "TM01K8XV6VK9ED2GBZSQ2VDTAT9T",
			Name:        "Team Beta",
			Description: "Second team",
			ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		},
		{
			ID:          "TM01K8XV6VK9ED2GBZSQ2VDTATZZ",
			Name:        "Team Gamma",
			Description: "Third team",
			ProjectID:   "PR01K8XV6J9H7BAEV49ZFVYS8R1K",
		},
	}

	// Store all teams in cache
	for _, team := range teams {
		c.Set(cache.TeamKey(team.ID), team)
	}
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify all teams can be retrieved
	for _, expectedTeam := range teams {
		cached, ok := c.Get(cache.TeamKey(expectedTeam.ID))
		assert.True(t, ok, "team %s should be in cache", expectedTeam.ID)

		cachedTeam, ok := cached.(*model.Team)
		assert.True(t, ok)
		assert.Equal(t, expectedTeam.ID, cachedTeam.ID)
		assert.Equal(t, expectedTeam.Name, cachedTeam.Name)
	}
}
