package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuperTeams(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup user IDs
	userID := data.UserIDs[0]
	adminUserID := data.UserIDs[1]

	// Assign admin role
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

	projectID := data.ProjectIDs[0]

	// Get superteam ID from seeded data for testing queries
	superTeamIDs := data.SuperTeamIDs[projectID]
	var seededSuperTeamID string
	if len(superTeamIDs) > 0 {
		seededSuperTeamID = superTeamIDs[0]
	}

	t.Run("admin can create superteam", func(t *testing.T) {
		t.Skip("CreateSuperTeam resolver not yet implemented")
	})

	t.Run("admin can create superteam with initial teams", func(t *testing.T) {
		t.Skip("CreateSuperTeam resolver not yet implemented")
	})

	t.Run("admin can update superteam", func(t *testing.T) {
		t.Skip("UpdateSuperTeam resolver not yet implemented")
	})

	t.Run("user cannot create superteam", func(t *testing.T) {
		t.Skip("CreateSuperTeam resolver not yet implemented")
	})

	t.Run("admin can assign teams to superteam", func(t *testing.T) {
		t.Skip("AssignTeamsToSuperTeam resolver not yet implemented")
	})

	t.Run("query superteam by id", func(t *testing.T) {
		if seededSuperTeamID == "" {
			t.Skip("No superteam in seeded data")
		}

		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetSuperTeam($id: ID!) {
				superteam(id: $id) {
					id
					name
					description
					teams {
						id
						name
					}
					parentProject {
						id
					}
				}
			}
		`, map[string]any{
			"id": seededSuperTeamID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Superteam struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				Description   string `json:"description"`
				Teams         []struct{ ID, Name string } `json:"teams"`
				ParentProject struct{ ID string } `json:"parentProject"`
			} `json:"superteam"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, seededSuperTeamID, result.Superteam.ID)
		assert.Equal(t, projectID, result.Superteam.ParentProject.ID)
	})

	t.Run("query superteams with filter", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetSuperTeams($filter: SuperTeamFilter) {
				superteams(filter: $filter, first: 100) {
					edges {
						node {
							id
							name
						}
					}
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"projectId": projectID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Superteams struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"superteams"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Greater(t, result.Superteams.TotalCount, 0)
	})

	t.Run("admin can award superteam achievement", func(t *testing.T) {
		if seededSuperTeamID == "" {
			t.Skip("No superteam in seeded data")
		}

		// First create an achievement
		achievementResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "SuperTeam Achievement",
				"descriptionPending":   "For super teams",
				"descriptionCompleted": "Super team did it!",
				"notificationText":     "Great job!",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               200,
				"hidden":               false,
			},
		})
		require.False(t, achievementResp.HasErrors())

		var achievementResult struct {
			CreateSimpleAchievement struct {
				ID string `json:"id"`
			} `json:"createSimpleAchievement"`
		}
		require.NoError(t, achievementResp.UnmarshalData(&achievementResult))
		achievementID := achievementResult.CreateSimpleAchievement.ID

		// Award to superteam
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AwardSuperTeamAchievement($superTeamId: ID!, $achievementId: ID!) {
				awardSuperTeamAchievement(superTeamId: $superTeamId, achievementId: $achievementId) {
					... on SimpleAchievement {
						id
						name
					}
				}
			}
		`, map[string]any{
			"superTeamId":   seededSuperTeamID,
			"achievementId": achievementID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
	})

	t.Run("admin can revoke superteam achievement", func(t *testing.T) {
		if seededSuperTeamID == "" {
			t.Skip("No superteam in seeded data")
		}

		// Create another achievement to revoke
		achievementResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Revocable Achievement",
				"descriptionPending":   "To be revoked",
				"descriptionCompleted": "Completed",
				"notificationText":     "Revoked",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               50,
				"hidden":               false,
			},
		})
		require.False(t, achievementResp.HasErrors())

		var achievementResult struct {
			CreateSimpleAchievement struct {
				ID string `json:"id"`
			} `json:"createSimpleAchievement"`
		}
		require.NoError(t, achievementResp.UnmarshalData(&achievementResult))
		achievementID := achievementResult.CreateSimpleAchievement.ID

		// Award first
		awardResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AwardSuperTeamAchievement($superTeamId: ID!, $achievementId: ID!) {
				awardSuperTeamAchievement(superTeamId: $superTeamId, achievementId: $achievementId) {
					... on SimpleAchievement { id }
				}
			}
		`, map[string]any{
			"superTeamId":   seededSuperTeamID,
			"achievementId": achievementID,
		})
		require.False(t, awardResp.HasErrors())

		// Now revoke
		revokeResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation RevokeSuperTeamAchievement($superTeamId: ID!, $achievementId: ID!) {
				revokeSuperTeamAchievement(superTeamId: $superTeamId, achievementId: $achievementId)
			}
		`, map[string]any{
			"superTeamId":   seededSuperTeamID,
			"achievementId": achievementID,
		})

		require.False(t, revokeResp.HasErrors(), "unexpected error: %s", revokeResp.ErrorMessage())

		var revokeResult struct {
			RevokeSuperTeamAchievement bool `json:"revokeSuperTeamAchievement"`
		}
		require.NoError(t, revokeResp.UnmarshalData(&revokeResult))
		assert.True(t, revokeResult.RevokeSuperTeamAchievement)
	})

	t.Run("admin can delete superteam", func(t *testing.T) {
		t.Skip("DeleteSuperTeam resolver not yet implemented")
	})
}
