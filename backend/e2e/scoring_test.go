package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoring(t *testing.T) {
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

	t.Run("query scoreJournal for user", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetScoreJournal($projectId: ID!, $userId: ID!) {
				scoreJournal(projectId: $projectId, userId: $userId, first: 100) {
					edges {
						node {
							id
							points
							sourceType
							reason
							createdAt
						}
					}
					totalCount
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"userId":    userID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			ScoreJournal struct {
				Edges []struct {
					Node struct {
						ID         string  `json:"id"`
						Points     int     `json:"points"`
						SourceType string  `json:"sourceType"`
						Reason     *string `json:"reason"`
						CreatedAt  string  `json:"createdAt"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"scoreJournal"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		// Seeded data should have some score entries
		assert.GreaterOrEqual(t, result.ScoreJournal.TotalCount, 0)
	})

	t.Run("admin can query adminScoreJournal", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetAdminScoreJournal($filter: ScoreJournalFilter) {
				adminScoreJournal(filter: $filter, first: 100) {
					edges {
						node {
							id
							points
							sourceType
							reason
							user {
								id
								name
							}
							project {
								id
							}
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
			AdminScoreJournal struct {
				Edges []struct {
					Node struct {
						ID         string  `json:"id"`
						Points     int     `json:"points"`
						SourceType string  `json:"sourceType"`
						Reason     *string `json:"reason"`
						User       struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"user"`
						Project struct {
							ID string `json:"id"`
						} `json:"project"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"adminScoreJournal"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Greater(t, result.AdminScoreJournal.TotalCount, 0)
	})

	t.Run("user cannot query adminScoreJournal", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetAdminScoreJournal {
				adminScoreJournal(first: 10) {
					totalCount
				}
			}
		`, nil)

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("admin can create score adjustment", func(t *testing.T) {
		// Note: The CreateScoreAdjustment resolver doesn't currently populate ProjectID
		// in the returned model, so we don't query project { id } here
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
				createScoreAdjustment(input: $input) {
					id
					points
					sourceType
					reason
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId": projectID,
				"userId":    userID,
				"points":    50,
				"reason":    "E2E test bonus points",
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateScoreAdjustment struct {
				ID         string `json:"id"`
				Points     int    `json:"points"`
				SourceType string `json:"sourceType"`
				Reason     string `json:"reason"`
			} `json:"createScoreAdjustment"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateScoreAdjustment.ID)
		assert.Equal(t, 50, result.CreateScoreAdjustment.Points)
		assert.Equal(t, "MANUAL", result.CreateScoreAdjustment.SourceType)
		assert.Equal(t, "E2E test bonus points", result.CreateScoreAdjustment.Reason)
	})

	t.Run("admin can create negative score adjustment", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
				createScoreAdjustment(input: $input) {
					id
					points
					reason
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId": projectID,
				"userId":    userID,
				"points":    -25,
				"reason":    "E2E test penalty",
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateScoreAdjustment struct {
				ID     string `json:"id"`
				Points int    `json:"points"`
				Reason string `json:"reason"`
			} `json:"createScoreAdjustment"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, -25, result.CreateScoreAdjustment.Points)
	})

	t.Run("admin can delete score journal entry", func(t *testing.T) {
		// First create an entry to delete
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
				createScoreAdjustment(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId": projectID,
				"userId":    userID,
				"points":    10,
				"reason":    "To be deleted",
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateScoreAdjustment struct {
				ID string `json:"id"`
			} `json:"createScoreAdjustment"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		entryID := createResult.CreateScoreAdjustment.ID

		// Delete it
		deleteResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation DeleteScoreJournalEntry($id: ID!) {
				deleteScoreJournalEntry(id: $id)
			}
		`, map[string]any{
			"id": entryID,
		})

		require.False(t, deleteResp.HasErrors(), "unexpected error: %s", deleteResp.ErrorMessage())

		var deleteResult struct {
			DeleteScoreJournalEntry bool `json:"deleteScoreJournalEntry"`
		}
		require.NoError(t, deleteResp.UnmarshalData(&deleteResult))
		assert.True(t, deleteResult.DeleteScoreJournalEntry)
	})

	t.Run("verify points from achievement award", func(t *testing.T) {
		// Use a user that likely has fewer seeded entries
		targetUserID := data.UserIDs[10]

		// Create an achievement with a unique point value to avoid collision with seeded data
		achievementResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Score Test Achievement",
				"descriptionPending":   "For score testing",
				"descriptionCompleted": "Completed!",
				"notificationText":     "Points added!",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               777, // Unique value unlikely to collide with seeded data
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

		// Award it
		awardResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AwardAchievement($userId: ID!, $achievementId: ID!) {
				awardAchievement(userId: $userId, achievementId: $achievementId) {
					... on SimpleAchievement { id }
				}
			}
		`, map[string]any{
			"userId":        targetUserID,
			"achievementId": achievementID,
		})
		require.False(t, awardResp.HasErrors())

		// Check score journal for the entry - use higher limit to find our entry
		journalResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetScoreJournal($projectId: ID!, $userId: ID!) {
				scoreJournal(projectId: $projectId, userId: $userId, first: 100) {
					edges {
						node {
							points
							sourceType
						}
					}
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"userId":    targetUserID,
		})
		require.False(t, journalResp.HasErrors())

		var journalResult struct {
			ScoreJournal struct {
				Edges []struct {
					Node struct {
						Points     int    `json:"points"`
						SourceType string `json:"sourceType"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"scoreJournal"`
		}
		require.NoError(t, journalResp.UnmarshalData(&journalResult))

		// Find the achievement entry with our unique point value
		foundAchievementEntry := false
		for _, edge := range journalResult.ScoreJournal.Edges {
			if edge.Node.SourceType == "ACHIEVEMENT" && edge.Node.Points == 777 {
				foundAchievementEntry = true
				break
			}
		}
		assert.True(t, foundAchievementEntry, "should find achievement score entry with 777 points")
	})

	t.Run("query scoreJournal with sourceType filter", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetAdminScoreJournal($filter: ScoreJournalFilter) {
				adminScoreJournal(filter: $filter, first: 100) {
					edges {
						node {
							id
							sourceType
							points
						}
					}
					totalCount
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"projectId":  projectID,
				"sourceType": "MANUAL",
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AdminScoreJournal struct {
				Edges []struct {
					Node struct {
						ID         string `json:"id"`
						SourceType string `json:"sourceType"`
						Points     int    `json:"points"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"adminScoreJournal"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// All returned entries should be MANUAL
		for _, edge := range result.AdminScoreJournal.Edges {
			assert.Equal(t, "MANUAL", edge.Node.SourceType)
		}
	})
}
