package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/require"
)

func TestLeaderboardConfigRelativeFilters(t *testing.T) {
	ctx := context.Background()
	db, _ := GetTestEnv()
	require.NoError(t, db.Clean(ctx))
	setupLeaderboardTestData(t, ctx, db)
	require.NoError(t, db.AssignRole(ctx, userAdult20ID, testutil.RoleAdmin))
	require.NoError(t, db.AssignRole(ctx, userSenior30ID, testutil.RoleAdmin))
	router, cleanup, err := testutil.SetupTestServer(ctx, db)
	require.NoError(t, err)
	t.Cleanup(cleanup)
	client := testutil.NewGraphQLClient(router)
	t.Cleanup(client.Close)
	admin, err := testutil.GenerateAdminToken(userAdult20ID)
	require.NoError(t, err)
	response := client.WithAuth(admin).MustExecute(t, `mutation($input:CreateLeaderboardConfigInput!){createLeaderboardConfig(input:$input){id filter{myChurch myTeam ageRange{min max}}}}`, map[string]any{"input": map[string]any{
		"projectId": testProjectID, "name": "U18 in my team", "entityType": "PERSONS", "filter": map[string]any{"myChurch": true, "myTeam": true, "ageRange": map[string]any{"min": 12, "max": 17}},
	}})
	require.False(t, response.HasErrors(), response.ErrorMessage())
	var created struct {
		CreateLeaderboardConfig struct {
			ID     string
			Filter struct {
				MyChurch, MyTeam bool
				AgeRange         struct{ Min, Max int }
			}
		}
	}
	require.NoError(t, response.UnmarshalData(&created))
	require.True(t, created.CreateLeaderboardConfig.Filter.MyChurch)
	require.True(t, created.CreateLeaderboardConfig.Filter.MyTeam)
	require.Equal(t, 17, created.CreateLeaderboardConfig.Filter.AgeRange.Max)
	const query = `query($id:ID!){leaderboardConfig(id:$id){filter{myTeam ageRange{min max}} leaderboard{totalCount edges{node{id}} me{id} nearestChurchRivals{id}}}}`
	// Same persisted config, different viewers, then the original viewer again:
	// the age group is fixed, so every viewer in the church and team sees the
	// same board, whatever their own age.
	for _, tc := range []struct {
		id   string
		want []string
	}{
		{userAdult20ID, []string{userYoung16ID, userYoung17ID}},
		{userSenior30ID, []string{userYoung16ID, userYoung17ID}},
		{userAdult20ID, []string{userYoung16ID, userYoung17ID}},
	} {
		token, err := testutil.GenerateAdminToken(tc.id)
		require.NoError(t, err)
		response = client.WithAuth(token).MustExecute(t, query, map[string]any{"id": created.CreateLeaderboardConfig.ID})
		require.False(t, response.HasErrors(), response.ErrorMessage())
		var result struct {
			LeaderboardConfig struct {
				Filter      struct{ MyTeam bool }
				Leaderboard struct {
					TotalCount int
					Edges      []struct{ Node struct{ ID string } }
				}
			}
		}
		require.NoError(t, response.UnmarshalData(&result))
		require.True(t, result.LeaderboardConfig.Filter.MyTeam)
		require.Equal(t, len(tc.want), result.LeaderboardConfig.Leaderboard.TotalCount)
		ids := []string{}
		for _, edge := range result.LeaderboardConfig.Leaderboard.Edges {
			ids = append(ids, edge.Node.ID)
		}
		require.ElementsMatch(t, tc.want, ids)
	}
}
