package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTeamChurchVisibility tests that teams are visible via myChurchTeams based on
// church membership of team members and the team creator.
//
// Expected behavior:
// 1. Empty team: Visible to church admins from the creator's church
// 2. Team with members from church B only: Visible only to church B admins (not creator's church A)
// 3. Team with members from multiple churches: Visible to admins from all churches with members
//
// Note: The current SQL query uses OR logic (member OR creator), so tests for scenario 2
// will fail. This is acceptable - the tests document the expected behavior.
func TestTeamChurchVisibility(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for test isolation
	require.NoError(t, dbMgr.Clean(ctx))

	// Create test churches
	church1ID := ulid.NewChurchID()
	church2ID := ulid.NewChurchID()
	require.NoError(t, dbMgr.CreateTestChurch(ctx, church1ID, "Church One", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestChurch(ctx, church2ID, "Church Two", "NO", "S"))

	// Create test project
	projectID := ulid.NewProjectID()
	require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Team Visibility Test Project"))

	// Update settings to point to this project
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, projectID)
	require.NoError(t, err)

	// Create users
	// Church Admin 1 - from Church 1, will create the team
	churchAdmin1ID := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, churchAdmin1ID, "Church Admin 1", "UNKNOWN", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), church1ID))

	// Church Admin 2 - from Church 2
	churchAdmin2ID := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, churchAdmin2ID, "Church Admin 2", "UNKNOWN", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), church2ID))

	// User from Church 1
	userChurch1ID := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, userChurch1ID, "User Church 1", "UNKNOWN", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), church1ID))

	// User from Church 2
	userChurch2ID := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, userChurch2ID, "User Church 2", "UNKNOWN", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), church2ID))

	// Superadmin (for adding members)
	superadminID := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, superadminID, "Superadmin", "UNKNOWN", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), church1ID))

	// Assign roles
	// Church Admin 1 needs CHURCH_ADMIN for church1 and PROJECT_ADMIN for the project to create teams
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdmin1ID, testutil.RoleChurchAdmin, &church1ID, nil, nil))
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdmin1ID, testutil.RoleProjectAdmin, nil, &projectID, nil))

	// Church Admin 2 needs CHURCH_ADMIN for church2
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdmin2ID, testutil.RoleChurchAdmin, &church2ID, nil, nil))

	// Superadmin role
	require.NoError(t, dbMgr.AssignRole(ctx, superadminID, testutil.RoleSuperAdmin))

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// Generate tokens
	// Church admins need admin JWT role to access myChurchTeams and create teams
	churchAdmin1Token, err := testutil.GenerateAdminToken(churchAdmin1ID)
	require.NoError(t, err)

	churchAdmin2Token, err := testutil.GenerateAdminToken(churchAdmin2ID)
	require.NoError(t, err)

	superadminToken, err := testutil.GenerateSuperAdminToken(superadminID)
	require.NoError(t, err)

	// GraphQL queries and mutations
	const getMyChurchTeamsQuery = `
		query GetMyChurchTeams($projectId: ID!) {
			project(id: $projectId) {
				myChurchTeams {
					id
					name
				}
			}
		}
	`

	const createTeamMutation = `
		mutation CreateTeam($projectId: ID!, $input: CreateTeamInput!) {
			createTeam(projectId: $projectId, input: $input) {
				id
				name
				joinCode
			}
		}
	`

	const addTeamMembersMutation = `
		mutation AddTeamMembers($teamId: ID!, $userIds: [ID!]!) {
			addTeamMembers(teamId: $teamId, userIds: $userIds, force: true) {
				id
				members {
					user { id }
				}
			}
		}
	`

	// Helper to check if a team is in the myChurchTeams list
	teamInList := func(teams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}, teamID string) bool {
		for _, team := range teams {
			if team.ID == teamID {
				return true
			}
		}
		return false
	}

	var teamID string

	t.Run("Scenario 1: Empty team visible to creator's church", func(t *testing.T) {
		// Church Admin 1 creates team "Team Alpha"
		createResp := client.WithAuth(churchAdmin1Token).MustExecute(t, createTeamMutation, map[string]any{
			"projectId": projectID,
			"input": map[string]any{
				"name":        "Team Alpha",
				"description": "Test team for visibility",
			},
		})
		require.False(t, createResp.HasErrors(), "unexpected error creating team: %s", createResp.ErrorMessage())

		var createResult struct {
			CreateTeam struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				JoinCode string `json:"joinCode"`
			} `json:"createTeam"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		teamID = createResult.CreateTeam.ID
		assert.NotEmpty(t, teamID)

		// Church Admin 1 queries myChurchTeams - should see the team
		resp1 := client.WithAuth(churchAdmin1Token).MustExecute(t, getMyChurchTeamsQuery, map[string]any{
			"projectId": projectID,
		})
		require.False(t, resp1.HasErrors(), "unexpected error: %s", resp1.ErrorMessage())

		var result1 struct {
			Project struct {
				MyChurchTeams []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"myChurchTeams"`
			} `json:"project"`
		}
		require.NoError(t, resp1.UnmarshalData(&result1))
		assert.True(t, teamInList(result1.Project.MyChurchTeams, teamID), "Church Admin 1 should see the team they created")

		// Church Admin 2 queries myChurchTeams - should NOT see the team
		resp2 := client.WithAuth(churchAdmin2Token).MustExecute(t, getMyChurchTeamsQuery, map[string]any{
			"projectId": projectID,
		})
		require.False(t, resp2.HasErrors(), "unexpected error: %s", resp2.ErrorMessage())

		var result2 struct {
			Project struct {
				MyChurchTeams []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"myChurchTeams"`
			} `json:"project"`
		}
		require.NoError(t, resp2.UnmarshalData(&result2))
		assert.False(t, teamInList(result2.Project.MyChurchTeams, teamID), "Church Admin 2 should NOT see the team (no members from their church)")
	})

	t.Run("Scenario 2: Adding member from different church changes visibility", func(t *testing.T) {
		// Superadmin adds userChurch2 to the team
		addResp := client.WithAuth(superadminToken).MustExecute(t, addTeamMembersMutation, map[string]any{
			"teamId":  teamID,
			"userIds": []string{userChurch2ID},
		})
		require.False(t, addResp.HasErrors(), "unexpected error adding member: %s", addResp.ErrorMessage())

		// Church Admin 1 queries myChurchTeams - should NOT see the team
		// (team has no members from church 1, only the creator is from church 1)
		// NOTE: This test will FAIL with current SQL (OR logic) - documenting expected behavior
		resp1 := client.WithAuth(churchAdmin1Token).MustExecute(t, getMyChurchTeamsQuery, map[string]any{
			"projectId": projectID,
		})
		require.False(t, resp1.HasErrors(), "unexpected error: %s", resp1.ErrorMessage())

		var result1 struct {
			Project struct {
				MyChurchTeams []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"myChurchTeams"`
			} `json:"project"`
		}
		require.NoError(t, resp1.UnmarshalData(&result1))
		assert.False(t, teamInList(result1.Project.MyChurchTeams, teamID), "Church Admin 1 should NOT see the team (members only from church 2)")

		// Church Admin 2 queries myChurchTeams - should see the team
		resp2 := client.WithAuth(churchAdmin2Token).MustExecute(t, getMyChurchTeamsQuery, map[string]any{
			"projectId": projectID,
		})
		require.False(t, resp2.HasErrors(), "unexpected error: %s", resp2.ErrorMessage())

		var result2 struct {
			Project struct {
				MyChurchTeams []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"myChurchTeams"`
			} `json:"project"`
		}
		require.NoError(t, resp2.UnmarshalData(&result2))
		assert.True(t, teamInList(result2.Project.MyChurchTeams, teamID), "Church Admin 2 should see the team (has member from their church)")
	})

	t.Run("Scenario 3: Adding member from church 1 makes both see team", func(t *testing.T) {
		// Superadmin adds userChurch1 to the team
		addResp := client.WithAuth(superadminToken).MustExecute(t, addTeamMembersMutation, map[string]any{
			"teamId":  teamID,
			"userIds": []string{userChurch1ID},
		})
		require.False(t, addResp.HasErrors(), "unexpected error adding member: %s", addResp.ErrorMessage())

		// Church Admin 1 queries myChurchTeams - should see the team
		resp1 := client.WithAuth(churchAdmin1Token).MustExecute(t, getMyChurchTeamsQuery, map[string]any{
			"projectId": projectID,
		})
		require.False(t, resp1.HasErrors(), "unexpected error: %s", resp1.ErrorMessage())

		var result1 struct {
			Project struct {
				MyChurchTeams []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"myChurchTeams"`
			} `json:"project"`
		}
		require.NoError(t, resp1.UnmarshalData(&result1))
		assert.True(t, teamInList(result1.Project.MyChurchTeams, teamID), "Church Admin 1 should see the team (has member from their church)")

		// Church Admin 2 queries myChurchTeams - should still see the team
		resp2 := client.WithAuth(churchAdmin2Token).MustExecute(t, getMyChurchTeamsQuery, map[string]any{
			"projectId": projectID,
		})
		require.False(t, resp2.HasErrors(), "unexpected error: %s", resp2.ErrorMessage())

		var result2 struct {
			Project struct {
				MyChurchTeams []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"myChurchTeams"`
			} `json:"project"`
		}
		require.NoError(t, resp2.UnmarshalData(&result2))
		assert.True(t, teamInList(result2.Project.MyChurchTeams, teamID), "Church Admin 2 should still see the team (has member from their church)")
	})
}
