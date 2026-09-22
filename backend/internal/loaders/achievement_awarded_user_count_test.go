package loaders

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapAchievementAwardedUserCounts(t *testing.T) {
	keys := []string{
		"AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
		"AC01K8XV6VK9ED2GBZSQ2VDTAT9T",
		"AC01K8XV6VK9ED2GBZSQ2VDTATAT",
	}
	// Second key has no awards; the row order differs from key order.
	rows := []*sqlc.GetBulkAchievementAwardedUserCountsRow{
		{AchievementID: "AC01K8XV6VK9ED2GBZSQ2VDTATAT", AwardedUserCount: 3},
		{AchievementID: "AC01K8XV6VK9ED2GBZSQ2VDTAT8T", AwardedUserCount: 88},
	}

	results := mapAchievementAwardedUserCounts(keys, rows)

	require.Len(t, results, len(keys))
	for _, r := range results {
		require.NoError(t, r.Error)
	}
	assert.Equal(t, int64(88), results[0].Data)
	assert.Equal(t, int64(0), results[1].Data, "an unearned achievement should be 0, not missing")
	assert.Equal(t, int64(3), results[2].Data)
}

func TestMapAchievementAwardedUserCountsEmptyRows(t *testing.T) {
	results := mapAchievementAwardedUserCounts([]string{"AC01K8XV6VK9ED2GBZSQ2VDTAT8T"}, nil)

	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	assert.Equal(t, int64(0), results[0].Data)
}
