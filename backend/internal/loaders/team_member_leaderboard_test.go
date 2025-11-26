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

func TestTeamMemberLeaderboardTagsCacheKey(t *testing.T) {
	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8T"
	cacheKey := cache.TeamMemberLeaderboardTagsKey(teamID, userID)

	assert.NotEmpty(t, cacheKey)
	assert.Contains(t, cacheKey, teamID)
	assert.Contains(t, cacheKey, userID)
	assert.Contains(t, cacheKey, "tags")
}

func TestTeamMemberLeaderboardTagsCacheKeyUniqueness(t *testing.T) {
	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	userA := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"
	userB := "US01K8XV6VK9ED2GBZSQ2VDTAT8B"

	cacheKeyA := cache.TeamMemberLeaderboardTagsKey(teamID, userA)
	cacheKeyB := cache.TeamMemberLeaderboardTagsKey(teamID, userB)

	assert.NotEqual(t, cacheKeyA, cacheKeyB, "different users should have different cache keys")
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

func TestTagsCachedSeparatelyPerUser(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	userA := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"
	userB := "US01K8XV6VK9ED2GBZSQ2VDTAT8B"

	// Cache tags for user A
	tagsA := cachedLeaderboardTags{
		userA: []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe},
	}
	cacheKeyA := cache.TeamMemberLeaderboardTagsKey(teamID, userA)
	c.Set(cacheKeyA, tagsA)

	// Cache tags for user B
	tagsB := cachedLeaderboardTags{
		userB: []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe},
	}
	cacheKeyB := cache.TeamMemberLeaderboardTagsKey(teamID, userB)
	c.Set(cacheKeyB, tagsB)

	time.Sleep(10 * time.Millisecond) // Allow ristretto to process async writes

	// Verify user A's tags
	cachedA, ok := c.Get(cacheKeyA)
	require.True(t, ok, "user A tags should be in cache")
	retrievedTagsA, ok := cachedA.(cachedLeaderboardTags)
	require.True(t, ok)
	assert.Contains(t, retrievedTagsA, userA)
	assert.NotContains(t, retrievedTagsA, userB)

	// Verify user B's tags
	cachedB, ok := c.Get(cacheKeyB)
	require.True(t, ok, "user B tags should be in cache")
	retrievedTagsB, ok := cachedB.(cachedLeaderboardTags)
	require.True(t, ok)
	assert.Contains(t, retrievedTagsB, userB)
	assert.NotContains(t, retrievedTagsB, userA)
}

func TestGetOrComputeTags_CacheHit(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	userID := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"
	entries := []cachedLeaderboardEntry{
		{ID: userID, Name: "User A", Score: 100, Rank: 1},
	}

	// Pre-populate tag cache
	expectedTags := cachedLeaderboardTags{
		userID: []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe},
	}
	cacheKey := cache.TeamMemberLeaderboardTagsKey(teamID, userID)
	c.Set(cacheKey, expectedTags)
	time.Sleep(10 * time.Millisecond)

	// Call getOrComputeTags - should return cached tags (db not needed for cache hit)
	tags := getOrComputeTags(context.Background(), nil, c, teamID, userID, entries)

	assert.Contains(t, tags, userID)
	assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe}, tags[userID])
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

	// No pre-populated cache - should compute and cache
	// Note: passing nil for db means TEAM_LEAD tags won't be computed (requires role lookup)
	// This test verifies ME tag computation works correctly
	tags := getOrComputeTags(context.Background(), nil, c, teamID, userID, entries)

	// User should have ME tag on their entry
	assert.Contains(t, tags, userID)
	assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe}, tags[userID])

	// Other user's entry should not have any tags (no ME tag, and no TEAM_LEAD without db)
	assert.NotContains(t, tags, otherUserID)

	// Verify tags were cached
	time.Sleep(10 * time.Millisecond)
	cacheKey := cache.TeamMemberLeaderboardTagsKey(teamID, userID)
	cached, ok := c.Get(cacheKey)
	assert.True(t, ok, "computed tags should be cached")
	cachedTags, ok := cached.(cachedLeaderboardTags)
	require.True(t, ok)
	assert.Contains(t, cachedTags, userID)
}

func TestInvalidateTeamMemberLeaderboardTags(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, c)

	teamID := "TM01K8XV6VK9ED2GBZSQ2VDTAT8T"
	userA := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"
	userB := "US01K8XV6VK9ED2GBZSQ2VDTAT8B"

	// Cache tags for both users
	tagsA := cachedLeaderboardTags{userA: []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe}}
	tagsB := cachedLeaderboardTags{userB: []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe}}

	cacheKeyA := cache.TeamMemberLeaderboardTagsKey(teamID, userA)
	cacheKeyB := cache.TeamMemberLeaderboardTagsKey(teamID, userB)

	c.Set(cacheKeyA, tagsA)
	c.Set(cacheKeyB, tagsB)
	time.Sleep(10 * time.Millisecond)

	// Verify both are cached
	_, okA := c.Get(cacheKeyA)
	_, okB := c.Get(cacheKeyB)
	assert.True(t, okA, "user A tags should be in cache before invalidation")
	assert.True(t, okB, "user B tags should be in cache before invalidation")

	// Invalidate all tag caches
	c.InvalidateTeamMemberLeaderboardTags()
	time.Sleep(10 * time.Millisecond)

	// Verify both are invalidated
	_, okA = c.Get(cacheKeyA)
	_, okB = c.Get(cacheKeyB)
	assert.False(t, okA, "user A tags should be invalidated")
	assert.False(t, okB, "user B tags should be invalidated")
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

	// Cache tags
	tags := cachedLeaderboardTags{userID: []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe}}
	tagsKey := cache.TeamMemberLeaderboardTagsKey(teamID, userID)
	c.Set(tagsKey, tags)

	time.Sleep(10 * time.Millisecond)

	// Invalidate only tag cache
	c.InvalidateTeamMemberLeaderboardTags()
	time.Sleep(10 * time.Millisecond)

	// Tags should be invalidated
	_, okTags := c.Get(tagsKey)
	assert.False(t, okTags, "tags should be invalidated")

	// Leaderboard data should still be cached
	_, okLeaderboard := c.Get(leaderboardKey)
	assert.True(t, okLeaderboard, "leaderboard data should still be in cache")
}
