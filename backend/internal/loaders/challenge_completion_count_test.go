package loaders

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapChallengeCompletionCounts(t *testing.T) {
	keys := []string{
		"CL01K8XV6VK9ED2GBZSQ2VDTAT8T",
		"CL01K8XV6VK9ED2GBZSQ2VDTAT9T",
		"CL01K8XV6VK9ED2GBZSQ2VDTATAT",
	}
	// Second key has no completions; the row order differs from key order.
	rows := []*sqlc.GetBulkChallengeCompletionCountsRow{
		{ChallengeID: "CL01K8XV6VK9ED2GBZSQ2VDTATAT", CompletionCount: 7},
		{ChallengeID: "CL01K8XV6VK9ED2GBZSQ2VDTAT8T", CompletionCount: 42},
	}

	results := mapChallengeCompletionCounts(keys, rows)

	require.Len(t, results, len(keys))
	for _, r := range results {
		require.NoError(t, r.Error)
	}
	assert.Equal(t, int64(42), results[0].Data)
	assert.Equal(t, int64(0), results[1].Data, "a challenge nobody completed should be 0, not missing")
	assert.Equal(t, int64(7), results[2].Data)
}

func TestMapChallengeCompletionCountsEmptyRows(t *testing.T) {
	results := mapChallengeCompletionCounts([]string{"CL01K8XV6VK9ED2GBZSQ2VDTAT8T"}, nil)

	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	assert.Equal(t, int64(0), results[0].Data)
}
