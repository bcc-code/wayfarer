package loaders

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionLeaderboardConfigsByProjectCache_DeduplicatesInput(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	cached, missing := partitionLeaderboardConfigsByProjectCache(
		[]string{"PR001", "PR002", "PR001", "PR002", "PR001"}, c,
	)

	assert.Empty(t, cached)
	assert.ElementsMatch(t, []string{"PR001", "PR002"}, missing)
	assert.Len(t, missing, 2, "each project ID should appear at most once in missing")
}

func TestPartitionLeaderboardConfigsByProjectCache_MixOfCachedAndUncached(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	cachedConfig := &model.LeaderboardConfig{ID: "LC001", Name: "Cached Config"}
	c.Set(cache.LeaderboardConfigsByProjectKey("PR001"), []*model.LeaderboardConfig{cachedConfig})
	c.Wait() // deterministically flush ristretto's async write buffer

	cached, missing := partitionLeaderboardConfigsByProjectCache(
		[]string{"PR001", "PR002", "PR003"}, c,
	)

	require.Contains(t, cached, "PR001")
	assert.Equal(t, []*model.LeaderboardConfig{cachedConfig}, cached["PR001"])
	assert.NotContains(t, cached, "PR002")
	assert.NotContains(t, cached, "PR003")
	assert.ElementsMatch(t, []string{"PR002", "PR003"}, missing)
}

func TestPartitionLeaderboardConfigsByProjectCache_DuplicateCachedIDOnlyLookedUpOnce(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	c.Set(cache.LeaderboardConfigsByProjectKey("PR001"), []*model.LeaderboardConfig{})
	c.Wait()

	hitsBefore := c.Hits()

	cached, missing := partitionLeaderboardConfigsByProjectCache(
		[]string{"PR001", "PR001", "PR001"}, c,
	)

	hitsAfter := c.Hits()

	assert.Contains(t, cached, "PR001")
	assert.Empty(t, missing)
	assert.Equal(t, uint64(1), hitsAfter-hitsBefore, "a duplicate already-cached ID should only be looked up once")
}

func TestPartitionLeaderboardConfigsByProjectCache_AllCached(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	c.Set(cache.LeaderboardConfigsByProjectKey("PR001"), []*model.LeaderboardConfig{})
	c.Wait()

	cached, missing := partitionLeaderboardConfigsByProjectCache([]string{"PR001"}, c)

	require.Contains(t, cached, "PR001")
	assert.Empty(t, missing)
}

func TestStoreLeaderboardConfigsByProjectInCache_EmptyResultIsCached(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	configsByProject := make(map[string][]*model.LeaderboardConfig)
	storeLeaderboardConfigsByProjectInCache([]string{"PR001"}, []*sqlc.LeaderboardConfig{}, configsByProject, c)
	c.Wait()

	require.Contains(t, configsByProject, "PR001")
	assert.Empty(t, configsByProject["PR001"])
	assert.NotNil(t, configsByProject["PR001"], "empty result should be an empty slice, not nil")

	cached, ok := c.Get(cache.LeaderboardConfigsByProjectKey("PR001"))
	require.True(t, ok, "an empty result should still be cached to avoid re-querying")
	assert.Equal(t, []*model.LeaderboardConfig{}, cached)
}

func TestStoreLeaderboardConfigsByProjectInCache_GroupsRowsByProject(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	rows := []*sqlc.LeaderboardConfig{
		{ID: "LC001", ProjectID: "PR001", Name: "A", EntityType: "PERSONS"},
		{ID: "LC002", ProjectID: "PR001", Name: "B", EntityType: "TEAMS"},
		{ID: "LC003", ProjectID: "PR002", Name: "C", EntityType: "CHURCHES"},
	}

	configsByProject := make(map[string][]*model.LeaderboardConfig)
	storeLeaderboardConfigsByProjectInCache([]string{"PR001", "PR002"}, rows, configsByProject, c)
	c.Wait()

	require.Len(t, configsByProject["PR001"], 2)
	require.Len(t, configsByProject["PR002"], 1)

	cachedPR001, ok := c.Get(cache.LeaderboardConfigsByProjectKey("PR001"))
	require.True(t, ok)
	assert.Len(t, cachedPR001, 2)
}
