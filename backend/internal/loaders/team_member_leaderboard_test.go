package loaders

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamMemberLeaderboardCacheKey(t *testing.T) {
	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.TeamMemberLeaderboardKey(teamID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, teamID)
	assert.Contains(t, cacheKey, "leaderboard")
}

func TestTeamMemberLeaderboardTeamLeadTagsCacheKey(t *testing.T) {
	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, teamID)
	assert.Contains(t, cacheKey, "teamlead")
}

func TestTeamMemberLeaderboardTeamLeadTagsCacheKeyUniqueness(t *testing.T) {
	teamA := "TM01K8XV6VK9ED2GBZSQ2VDTAT8A"
	teamB := "TM01K8XV6VK9ED2GBZSQ2VDTAT8B"

	cacheKeyA := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamA)
	cacheKeyB := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamB)

	assert.NotEqual(t, cacheKeyA, cacheKeyB, "different teams should have different cache keys")
}

func TestCachedLeaderboardEntryWithoutTags(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	entries := []cachedLeaderboardEntry{
		{
			ID:          "US01K8XV6VK9ED2GBZSQ2VDTAT8A",
			Name:        "User A",
			Description: "Church A",
			Score:       100,
			Rank:        1,
		},
		{
			ID:          "US01K8XV6VK9ED2GBZSQ2VDTAT8B",
			Name:        "User B",
			Description: "Church B",
			Score:       90,
			Rank:        2,
		},
	}

	// Cache leaderboard data without tags
	cacheKey := cache.TeamMemberLeaderboardKey(teamID)
	c.Set(cacheKey, entries)
	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Retrieve from cache
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "leaderboard should be in cache")
	require.NotNil(t, cached)

	cachedEntries, ok := cached.([]cachedLeaderboardEntry)
	assert.True(t, ok, "cached value should be []cachedLeaderboardEntry")
	assert.Len(t, cachedEntries, 2)
	assert.Equal(t, "User A", cachedEntries[0].Name)
	assert.Equal(t, "User B", cachedEntries[1].Name)
}

func TestTeamLeadTagsCachedPerTeam(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	teamLeadUserID := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"

	// Cache TEAM_LEAD tags for the team (viewer-independent)
	teamLeadTags := map[string]bool{
		teamLeadUserID: true,
	}
	cacheKey := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamID)
	c.Set(cacheKey, teamLeadTags)

	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify TEAM_LEAD tags are cached per team
	cached, ok := c.Get(cacheKey)
	require.True(t, ok, "TEAM_LEAD tags should be in cache")
	retrievedTags, ok := cached.(map[string]bool)
	require.True(t, ok)
	assert.True(t, retrievedTags[teamLeadUserID], "team lead should be marked")
}

func TestGetOrComputeTags_TeamLeadCacheHit(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	currentUserID := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"
	teamLeadUserID := "US01K8XV6VK9ED2GBZSQ2VDTAT8B"
	entries := []cachedLeaderboardEntry{
		{ID: currentUserID, Name: "User A", Score: 100, Rank: 1},
		{ID: teamLeadUserID, Name: "User B", Score: 90, Rank: 2},
	}

	// Pre-populate TEAM_LEAD tag cache (per team, not per user)
	teamLeadTags := map[string]bool{
		teamLeadUserID: true,
	}
	cacheKey := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamID)
	c.Set(cacheKey, teamLeadTags)
	time.Sleep(10 * time.Millisecond)

	// Call getOrComputeTags - TEAM_LEAD from cache, ME computed on-the-fly
	tags := getOrComputeTags(context.Background(), nil, c, teamID, currentUserID, entries)

	// Current user should have ME tag
	assert.Contains(t, tags, currentUserID)
	assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe}, tags[currentUserID])

	// Team lead should have TEAM_LEAD tag
	assert.Contains(t, tags, teamLeadUserID)
	assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagTeamLead}, tags[teamLeadUserID])
}

func TestGetOrComputeTags_CacheMiss(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"
	otherUserID := "US01K8XV6VK9ED2GBZSQ2VDTAT8B"
	entries := []cachedLeaderboardEntry{
		{ID: userID, Name: "User A", Score: 100, Rank: 1},
		{ID: otherUserID, Name: "User B", Score: 90, Rank: 2},
	}

	// No pre-populated cache - should compute and cache TEAM_LEAD tags per team
	// Note: passing nil for db means TEAM_LEAD tags won't be computed (requires role lookup)
	// ME tag is computed on-the-fly (no caching needed)
	tags := getOrComputeTags(context.Background(), nil, c, teamID, userID, entries)

	// User should have ME tag on their entry (computed on-the-fly)
	assert.Contains(t, tags, userID)
	assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe}, tags[userID])

	// Other user's entry should not have any tags (no ME tag, and no TEAM_LEAD without db)
	assert.NotContains(t, tags, otherUserID)

	// Verify TEAM_LEAD tags were cached per team (even if empty)
	time.Sleep(10 * time.Millisecond)
	cacheKey := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamID)
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "TEAM_LEAD tags should be cached per team")
	cachedTags, ok := cached.(map[string]bool)
	require.True(t, ok)
	// No team leads without db access
	assert.Empty(t, cachedTags)
}

func TestInvalidateTeamMemberLeaderboardTags(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamA := "TM01K8XV6VK9ED2GBZSQ2VDTAT8A"
	teamB := "TM01K8XV6VK9ED2GBZSQ2VDTAT8B"

	// Cache TEAM_LEAD tags for both teams (viewer-independent)
	tagsA := map[string]bool{"US01K8XV6VK9ED2GBZSQ2VDTAT8X": true}
	tagsB := map[string]bool{"US01K8XV6VK9ED2GBZSQ2VDTAT8Y": true}

	cacheKeyA := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamA)
	cacheKeyB := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamB)

	c.Set(cacheKeyA, tagsA)
	c.Set(cacheKeyB, tagsB)
	time.Sleep(10 * time.Millisecond)

	// Verify both are cached
	_, okA := c.Get(cacheKeyA)
	_, okB := c.Get(cacheKeyB)
	assert.True(t, okA, "team A TEAM_LEAD tags should be in cache before invalidation")
	assert.True(t, okB, "team B TEAM_LEAD tags should be in cache before invalidation")

	// Invalidate all tag caches
	c.InvalidateTeamMemberLeaderboardTags()
	time.Sleep(10 * time.Millisecond)

	// Verify both are invalidated
	_, okA = c.Get(cacheKeyA)
	_, okB = c.Get(cacheKeyB)
	assert.False(t, okA, "team A TEAM_LEAD tags should be invalidated")
	assert.False(t, okB, "team B TEAM_LEAD tags should be invalidated")
}

func TestLeaderboardDataNotAffectedByTagInvalidation(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"

	// Cache leaderboard data
	entries := []cachedLeaderboardEntry{
		{ID: userID, Name: "User A", Score: 100, Rank: 1},
	}
	leaderboardKey := cache.TeamMemberLeaderboardKey(teamID)
	c.Set(leaderboardKey, entries)

	// Cache TEAM_LEAD tags (per team, viewer-independent)
	tags := map[string]bool{userID: true}
	tagsKey := cache.TeamMemberLeaderboardTeamLeadTagsKey(teamID)
	c.Set(tagsKey, tags)

	time.Sleep(10 * time.Millisecond)

	// Invalidate only tag cache
	c.InvalidateTeamMemberLeaderboardTags()
	time.Sleep(10 * time.Millisecond)

	// Tags should be invalidated
	_, okTags := c.Get(tagsKey)
	assert.False(t, okTags, "TEAM_LEAD tags should be invalidated")

	// Leaderboard data should still be cached
	_, okLeaderboard := c.Get(leaderboardKey)
	assert.True(t, okLeaderboard, "leaderboard data should still be in cache")
}
