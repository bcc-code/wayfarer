package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The leaderboard apply worker calls InvalidateLeaderboardScores after it
// applied queued score deltas: every score-derived ranking must go (they may
// have been cached between the score write and the apply), while caches that
// don't depend on leaderboard tables stay.
func TestInvalidateLeaderboardScores(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()

	const (
		projectID = "PR00000000000000000000000001"
		eventID   = "EV00000000000000000000000001"
		userID    = "US00000000000000000000000001"
		teamID    = "TM00000000000000000000000001"
	)

	stale := []string{
		LeaderboardKey("project", projectID, "PERSONS", nil, "first"),
		LeaderboardKey("event", eventID, "TEAMS", map[string]string{"f": "x"}, "first"),
		LeaderboardPositionKey("project", projectID, "PERSONS", nil, userID),
		"leaderboard:count:project:" + projectID + ":PERSONS",
		"leaderboard:full:event:" + eventID + ":PERSONS",
		TeamMemberLeaderboardKey(teamID),
		GQLResponseUserPrefix(userID) + "somehash",
		PrefixGQLResponseShared + "somehash",
	}
	kept := []string{
		LeaderboardConfigKey("LC00000000000000000000000001"),
		LeaderboardConfigsByProjectKey(projectID),
		UserKey(userID),
		ProjectKey(projectID),
		UserProjectPointsKey(userID, projectID),
	}
	for _, k := range append(append([]string{}, stale...), kept...) {
		c.Set(k, "cached")
	}
	c.Wait()
	for _, k := range append(append([]string{}, stale...), kept...) {
		_, ok := c.Get(k)
		require.True(t, ok, "precondition: %s should be cached", k)
	}

	c.InvalidateLeaderboardScores()
	c.Wait()

	for _, k := range stale {
		_, ok := c.Get(k)
		assert.False(t, ok, "expected %s to be invalidated", k)
	}
	for _, k := range kept {
		_, ok := c.Get(k)
		assert.True(t, ok, "expected %s to survive", k)
	}
}
