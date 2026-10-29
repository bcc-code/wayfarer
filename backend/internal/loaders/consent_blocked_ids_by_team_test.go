package loaders

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapConsentBlockedIDsByTeam(t *testing.T) {
	teamA := "TM01K8XV6VK9ED2GBZSQ2VDTAT8A"
	teamB := "TM01K8XV6VK9ED2GBZSQ2VDTAT8B"
	teamC := "TM01K8XV6VK9ED2GBZSQ2VDTAT8C"
	// teamB has no blocked members; the rows are not in key order
	rows := []*sqlc.GetConsentBlockedUserIDsInTeamsRow{
		{TeamID: teamC, UserID: "US3"},
		{TeamID: teamA, UserID: "US1"},
		{TeamID: teamA, UserID: "US2"},
	}

	results := mapConsentBlockedIDsByTeam([]string{teamA, teamB, teamC}, rows)

	require.Len(t, results, 3)
	for _, r := range results {
		require.NoError(t, r.Error)
	}
	assert.Equal(t, []string{"US1", "US2"}, results[0].Data)
	assert.Equal(t, []string{}, results[1].Data, "a team without blocked members should get an empty slice, not nil")
	assert.Equal(t, []string{"US3"}, results[2].Data)
}

func TestMapConsentBlockedIDsByTeamEmptyRows(t *testing.T) {
	results := mapConsentBlockedIDsByTeam([]string{"TM01K8XV6VK9ED2GBZSQ2VDTAT8A"}, nil)

	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	assert.Equal(t, []string{}, results[0].Data)
}
