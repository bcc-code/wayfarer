package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeaderboardConfigEntryLimits(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))
	setupLeaderboardTestData(t, ctx, dbMgr)
	// The shared fixture ties every score; use distinct ranks for cursor checks.
	for i, userID := range []string{userYoung16ID, userYoung17ID, userAdult20ID, userAdult22ID, userAdult25ID, userSenior30ID, userSenior40ID, userElder50ID} {
		require.NoError(t, dbMgr.AddScoreForUser(ctx, userID, testProjectID, i+1))
	}
	require.NoError(t, dbMgr.AssignRole(ctx, userAdult20ID, testutil.RoleAdmin))
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	t.Cleanup(cleanup)
	client := testutil.NewGraphQLClient(router)
	t.Cleanup(client.Close)
	token, err := testutil.GenerateAdminToken(userAdult20ID)
	require.NoError(t, err)

	const create = `mutation($input: CreateLeaderboardConfigInput!) {
		createLeaderboardConfig(input: $input) { id maxEntries }
	}`
	const update = `mutation($id: ID!, $input: UpdateLeaderboardConfigInput!) {
		updateLeaderboardConfig(id: $id, input: $input) { id maxEntries }
	}`
	const read = `query($id: ID!, $first: Int, $after: String, $last: Int, $before: String) {
		leaderboardConfig(id: $id) {
			maxEntries
			leaderboard(first: $first, after: $after, last: $last, before: $before) {
				totalCount
				edges { cursor node { id rank } }
				me { id }
				pageInfo { hasNextPage hasPreviousPage }
			}
		}
	}`
	input := map[string]any{"projectId": testProjectID, "name": "Top three", "entityType": "PERSONS", "maxEntries": 3}
	resp := client.WithAuth(token).MustExecute(t, create, map[string]any{"input": input})
	require.False(t, resp.HasErrors(), resp.ErrorMessage())
	var created struct {
		CreateLeaderboardConfig struct {
			ID         string
			MaxEntries *int
		}
	}
	require.NoError(t, resp.UnmarshalData(&created))
	id := created.CreateLeaderboardConfig.ID
	require.NotNil(t, created.CreateLeaderboardConfig.MaxEntries)
	assert.Equal(t, 3, *created.CreateLeaderboardConfig.MaxEntries)

	for _, tt := range []struct {
		name           string
		args           map[string]any
		count          int
		previous, next bool
	}{
		{name: "configured default", count: 3},
		{name: "oversized request", args: map[string]any{"first": 100}, count: 3},
		{name: "first page", args: map[string]any{"first": 2}, count: 2, next: true},
		{name: "last page", args: map[string]any{"first": 2, "after": "2"}, count: 1, previous: true},
		{name: "cannot page past cap", args: map[string]any{"first": 100, "after": "3"}, previous: true},
		{name: "backward", args: map[string]any{"last": 2}, count: 2, previous: true},
		{name: "before beyond cap", args: map[string]any{"last": 100, "before": "99"}, count: 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]any{"id": id}
			for key, value := range tt.args {
				args[key] = value
			}
			resp := client.WithAuth(token).MustExecute(t, read, args)
			require.False(t, resp.HasErrors(), resp.ErrorMessage())
			var result struct {
				LeaderboardConfig struct {
					MaxEntries  *int
					Leaderboard struct {
						TotalCount int
						Edges      []struct {
							Cursor string
							Node   struct {
								ID   string
								Rank int
							}
						}
						Me       struct{ ID string }
						PageInfo struct{ HasNextPage, HasPreviousPage bool }
					}
				}
			}
			require.NoError(t, resp.UnmarshalData(&result))
			board := result.LeaderboardConfig.Leaderboard
			require.Len(t, board.Edges, tt.count)
			assert.Equal(t, 8, board.TotalCount)
			assert.Equal(t, userAdult20ID, board.Me.ID)
			assert.Equal(t, tt.next, board.PageInfo.HasNextPage)
			assert.Equal(t, tt.previous, board.PageInfo.HasPreviousPage)
		})
	}

	// Updating a warmed config must expose the new limit through every loader.
	for _, limit := range []any{5, nil} {
		input := map[string]any{"name": "Updated", "entityType": "PERSONS", "maxEntries": limit, "sortOrder": 0, "isActive": true}
		resp := client.WithAuth(token).MustExecute(t, update, map[string]any{"id": id, "input": input})
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
		resp = client.WithAuth(token).MustExecute(t, `query($projectId: ID!) {
			project(id: $projectId) { leaderboards { maxEntries leaderboard { edges { node { id } } } } }
		}`, map[string]any{"projectId": testProjectID})
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
		var result struct {
			Project struct {
				Leaderboards []struct {
					MaxEntries  *int
					Leaderboard struct {
						Edges []struct{ Node struct{ ID string } }
					}
				}
			}
		}
		require.NoError(t, resp.UnmarshalData(&result))
		require.Len(t, result.Project.Leaderboards, 1)
		board := result.Project.Leaderboards[0]
		if limit == nil {
			assert.Nil(t, board.MaxEntries)
			assert.Len(t, board.Leaderboard.Edges, 8)
		} else {
			require.NotNil(t, board.MaxEntries)
			assert.Equal(t, 5, *board.MaxEntries)
			assert.Len(t, board.Leaderboard.Edges, 5)
		}
	}

	for _, invalid := range []int{0, -1} {
		input["maxEntries"] = invalid
		resp := client.WithAuth(token).MustExecute(t, create, map[string]any{"input": input})
		assert.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "maxEntries must be")
		delete(input, "projectId")
		input["sortOrder"], input["isActive"] = 0, true
		resp = client.WithAuth(token).MustExecute(t, update, map[string]any{"id": id, "input": input})
		assert.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "maxEntries must be")
		input["projectId"] = testProjectID
	}
}
