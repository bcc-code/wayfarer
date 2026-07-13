package loaders

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserProjectKey(t *testing.T) {
	key := UserProjectKey{
		UserID:    "US01K8XV6VK9ED2GBZSQ2VDTAT8T",
		ProjectID: "PR01K8XV6VK9ED2GBZSQ2VDTAT8T",
	}

	assert.Equal(t, "US01K8XV6VK9ED2GBZSQ2VDTAT8T:PR01K8XV6VK9ED2GBZSQ2VDTAT8T", key.String())
	assert.Equal(t, key, key.Raw())
}

func TestMapUserProjectScores(t *testing.T) {
	keys := []UserProjectKey{
		{UserID: "US01K8XV6VK9ED2GBZSQ2VDTAT8T", ProjectID: "PR01K8XV6VK9ED2GBZSQ2VDTAT8T"},
		{UserID: "US01K8XV6VK9ED2GBZSQ2VDTAT9T", ProjectID: "PR01K8XV6VK9ED2GBZSQ2VDTAT8T"},
		{UserID: "US01K8XV6VK9ED2GBZSQ2VDTAT8T", ProjectID: "PR01K8XV6VK9ED2GBZSQ2VDTAT9T"},
	}
	// Second key has no journal entries; the row order differs from key order
	rows := []*sqlc.GetBulkUserProjectScoresRow{
		{UserID: "US01K8XV6VK9ED2GBZSQ2VDTAT8T", ProjectID: "PR01K8XV6VK9ED2GBZSQ2VDTAT9T", TotalScore: 75},
		{UserID: "US01K8XV6VK9ED2GBZSQ2VDTAT8T", ProjectID: "PR01K8XV6VK9ED2GBZSQ2VDTAT8T", TotalScore: 120},
	}

	results := mapUserProjectScores(keys, rows)

	require.Len(t, results, len(keys))
	for _, r := range results {
		require.NoError(t, r.Error)
	}
	assert.Equal(t, int64(120), results[0].Data)
	assert.Equal(t, int64(0), results[1].Data, "missing pair should default to 0")
	assert.Equal(t, int64(75), results[2].Data)
}

func TestMapUserProjectScoresEmptyRows(t *testing.T) {
	keys := []UserProjectKey{
		{UserID: "US01K8XV6VK9ED2GBZSQ2VDTAT8T", ProjectID: "PR01K8XV6VK9ED2GBZSQ2VDTAT8T"},
	}

	results := mapUserProjectScores(keys, nil)

	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	assert.Equal(t, int64(0), results[0].Data)
}
