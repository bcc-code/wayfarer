package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeams(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup user IDs - use different users for different roles
	userID := data.UserIDs[0]
	adminUserID := data.UserIDs[1]

	// Assign database roles
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	userToken, err := testutil.GenerateUserToken(userID)
	require.NoError(t, err)

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)

	// Get first project's team
	projectID := data.ProjectIDs[0]
	teamIDs := data.TeamIDs[projectID]
	require.NotEmpty(t, teamIDs, "should have seeded teams")
	teamID := teamIDs[0]

	t.Run("admin can query team by id with members", func(t *testing.T) {
		// Use admin token since members access is restricted
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetTeam($id: ID!) {
				team(id: $id) {
					id
					name
					description
					parentProject { id name }
					members {
						user { id name }
						joinedAt
					}
				}
			}
		`, map[string]any{"id": teamID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Team struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				Description   string `json:"description"`
				ParentProject struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"parentProject"`
				Members []struct {
					User struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"user"`
					JoinedAt string `json:"joinedAt"`
				} `json:"members"`
			} `json:"team"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, teamID, result.Team.ID)
		assert.NotEmpty(t, result.Team.Name)
		assert.Equal(t, projectID, result.Team.ParentProject.ID)
	})

	t.Run("user can query team basic info", func(t *testing.T) {
		// Users can query basic team info without members
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetTeam($id: ID!) {
				team(id: $id) {
					id
					name
					description
					parentProject { id }
				}
			}
		`, map[string]any{"id": teamID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Team struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				Description   string `json:"description"`
				ParentProject struct {
					ID string `json:"id"`
				} `json:"parentProject"`
			} `json:"team"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, teamID, result.Team.ID)
		assert.NotEmpty(t, result.Team.Name)
	})

	t.Run("admin can list all teams via filter", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetTeams($projectId: ID!) {
				teams(filter: { projectId: $projectId }, first: 100) {
					edges {
						node {
							id
							name
							joinCode
						}
					}
					totalCount
				}
			}
		`, map[string]any{"projectId": projectID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Teams struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						Name     string `json:"name"`
						JoinCode string `json:"joinCode"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"teams"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Greater(t, result.Teams.TotalCount, 0)
	})

	t.Run("admin can create team", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateTeam($projectId: ID!, $input: CreateTeamInput!) {
				createTeam(projectId: $projectId, input: $input) {
					id
					name
					description
					joinCode
					parentProject { id }
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"input": map[string]any{
				"name":        "E2E Test Team",
				"description": "Created by E2E test",
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateTeam struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				Description   string `json:"description"`
				JoinCode      string `json:"joinCode"`
				ParentProject struct {
					ID string `json:"id"`
				} `json:"parentProject"`
			} `json:"createTeam"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateTeam.ID)
		assert.Equal(t, "E2E Test Team", result.CreateTeam.Name)
		assert.NotEmpty(t, result.CreateTeam.JoinCode)
		assert.Equal(t, projectID, result.CreateTeam.ParentProject.ID)
	})

	t.Run("user cannot create team", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation CreateTeam($projectId: ID!, $input: CreateTeamInput!) {
				createTeam(projectId: $projectId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"input": map[string]any{
				"name":        "Should Fail",
				"description": "This should not be created",
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("admin can query superteams", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetSuperTeams($projectId: ID!) {
				superteams(filter: { projectId: $projectId }, first: 100) {
					edges {
						node {
							id
							name
							teams { id name }
						}
					}
					totalCount
				}
			}
		`, map[string]any{"projectId": projectID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Superteams struct {
				Edges []struct {
					Node struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						Teams []struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"teams"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"superteams"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Should have seeded superteams
		assert.Greater(t, result.Superteams.TotalCount, 0)
	})

	t.Run("admin can add team members", func(t *testing.T) {
		// Create a fresh team for this test
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateTeam($projectId: ID!, $input: CreateTeamInput!) {
				createTeam(projectId: $projectId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"input": map[string]any{
				"name":        "Member Test Team",
				"description": "Team for testing member additions",
			},
		})
		require.False(t, createResp.HasErrors(), "unexpected error creating team: %s", createResp.ErrorMessage())

		var createResult struct {
			CreateTeam struct{ ID string } `json:"createTeam"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		newTeamID := createResult.CreateTeam.ID

		// Add members
		userToAdd := data.UserIDs[2] // Use a different user than the admin
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddMembers($teamId: ID!, $userIds: [ID!]!) {
				addTeamMembers(teamId: $teamId, userIds: $userIds, force: true) {
					id
					members {
						user { id }
					}
				}
			}
		`, map[string]any{
			"teamId":  newTeamID,
			"userIds": []string{userToAdd},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AddTeamMembers struct {
				ID      string `json:"id"`
				Members []struct {
					User struct {
						ID string `json:"id"`
					} `json:"user"`
				} `json:"members"`
			} `json:"addTeamMembers"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Verify the user was added
		found := false
		for _, m := range result.AddTeamMembers.Members {
			if m.User.ID == userToAdd {
				found = true
				break
			}
		}
		assert.True(t, found, "user should be in team members")
	})
}
