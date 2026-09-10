package loaders

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionLeaderboardConfigsByEventCache_DeduplicatesInput(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	cached, missing := partitionLeaderboardConfigsByEventCache(
		[]string{"EV001", "EV002", "EV001", "EV002", "EV001"}, c,
	)

	assert.Empty(t, cached)
	assert.ElementsMatch(t, []string{"EV001", "EV002"}, missing)
	assert.Len(t, missing, 2, "each event ID should appear at most once in missing")
}

func TestPartitionLeaderboardConfigsByEventCache_MixOfCachedAndUncached(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	cachedConfig := &model.LeaderboardConfig{ID: "LC001", Name: "Cached Config"}
	c.Set(cache.LeaderboardConfigsByEventKey("EV001"), []*model.LeaderboardConfig{cachedConfig})
	c.Wait()

	cached, missing := partitionLeaderboardConfigsByEventCache(
		[]string{"EV001", "EV002", "EV003"}, c,
	)

	require.Contains(t, cached, "EV001")
	assert.Equal(t, []*model.LeaderboardConfig{cachedConfig}, cached["EV001"])
	assert.NotContains(t, cached, "EV002")
	assert.NotContains(t, cached, "EV003")
	assert.ElementsMatch(t, []string{"EV002", "EV003"}, missing)
}

func TestPartitionLeaderboardConfigsByEventCache_DuplicateCachedIDOnlyLookedUpOnce(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	c.Set(cache.LeaderboardConfigsByEventKey("EV001"), []*model.LeaderboardConfig{})
	c.Wait()

	hitsBefore := c.Hits()

	cached, missing := partitionLeaderboardConfigsByEventCache(
		[]string{"EV001", "EV001", "EV001"}, c,
	)

	hitsAfter := c.Hits()

	assert.Contains(t, cached, "EV001")
	assert.Empty(t, missing)
	assert.Equal(t, uint64(1), hitsAfter-hitsBefore, "a duplicate already-cached ID should only be looked up once")
}

func TestStoreLeaderboardConfigsByEventInCache_EmptyResultIsCached(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	configsByEvent := make(map[string][]*model.LeaderboardConfig)
	storeLeaderboardConfigsByEventInCache([]string{"EV001"}, []*sqlc.LeaderboardConfig{}, configsByEvent, c)
	c.Wait()

	require.Contains(t, configsByEvent, "EV001")
	assert.Empty(t, configsByEvent["EV001"])
	assert.NotNil(t, configsByEvent["EV001"], "empty result should be an empty slice, not nil")

	cached, ok := c.Get(cache.LeaderboardConfigsByEventKey("EV001"))
	require.True(t, ok, "an empty result should still be cached to avoid re-querying")
	assert.Equal(t, []*model.LeaderboardConfig{}, cached)
}

func TestStoreLeaderboardConfigsByEventInCache_GroupsRowsByEvent(t *testing.T) {
	c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	require.NoError(t, err)

	eventID1 := "EV001"
	eventID2 := "EV002"
	rows := []*sqlc.LeaderboardConfig{
		{ID: "LC001", EventID: &eventID1, Name: "A", EntityType: "PERSONS"},
		{ID: "LC002", EventID: &eventID1, Name: "B", EntityType: "TEAMS"},
		{ID: "LC003", EventID: &eventID2, Name: "C", EntityType: "CHURCHES"},
	}

	configsByEvent := make(map[string][]*model.LeaderboardConfig)
	storeLeaderboardConfigsByEventInCache([]string{eventID1, eventID2}, rows, configsByEvent, c)
	c.Wait()

	require.Len(t, configsByEvent[eventID1], 2)
	require.Len(t, configsByEvent[eventID2], 1)

	cachedEV001, ok := c.Get(cache.LeaderboardConfigsByEventKey(eventID1))
	require.True(t, ok)
	assert.Len(t, cachedEV001, 2)
}
