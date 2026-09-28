package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAchievements(t *testing.T) {
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

	// Store created achievement IDs for later tests
	var simpleAchievementID string
	var contentAchievementID string

	t.Run("admin can create simple achievement", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
					name
					descriptionPending
					descriptionCompleted
					notificationText
					imagePending
					imageCompleted
					points
					hidden
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Test Simple Achievement",
				"descriptionPending":   "Complete to earn this achievement",
				"descriptionCompleted": "You earned this achievement!",
				"notificationText":     "Congratulations!",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               100,
				"hidden":               false,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateSimpleAchievement struct {
				ID                   string `json:"id"`
				Name                 string `json:"name"`
				DescriptionPending   string `json:"descriptionPending"`
				DescriptionCompleted string `json:"descriptionCompleted"`
				NotificationText     string `json:"notificationText"`
				ImagePending         string `json:"imagePending"`
				ImageCompleted       string `json:"imageCompleted"`
				Points               int    `json:"points"`
				Hidden               bool   `json:"hidden"`
			} `json:"createSimpleAchievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateSimpleAchievement.ID)
		assert.Equal(t, "Test Simple Achievement", result.CreateSimpleAchievement.Name)
		assert.Equal(t, 100, result.CreateSimpleAchievement.Points)
		assert.False(t, result.CreateSimpleAchievement.Hidden)

		simpleAchievementID = result.CreateSimpleAchievement.ID
	})

	t.Run("admin can create content achievement", func(t *testing.T) {
		// First, get an external content ID from seeded data
		contentResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetExternalContents($filter: ExternalContentFilter!) {
				externalContents(filter: $filter, first: 1) {
					edges {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"filter": map[string]any{},
		})
		require.False(t, contentResp.HasErrors(), "unexpected error: %s", contentResp.ErrorMessage())

		var contentResult struct {
			ExternalContents struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"externalContents"`
		}
		require.NoError(t, contentResp.UnmarshalData(&contentResult))

		if len(contentResult.ExternalContents.Edges) == 0 {
			t.Skip("No external content available for content achievement test")
		}

		externalContentID := contentResult.ExternalContents.Edges[0].Node.ID

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateContentAchievement($input: CreateContentAchievementInput!) {
				createContentAchievement(input: $input) {
					id
					name
					descriptionPending
					descriptionCompleted
					points
					totalItems
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Test Content Achievement",
				"descriptionPending":   "Complete the content to earn",
				"descriptionCompleted": "Content completed!",
				"notificationText":     "You finished!",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               50,
				"hidden":               false,
				"items": []map[string]any{
					{"externalContentId": externalContentID},
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateContentAchievement struct {
				ID                   string `json:"id"`
				Name                 string `json:"name"`
				DescriptionPending   string `json:"descriptionPending"`
				DescriptionCompleted string `json:"descriptionCompleted"`
				Points               int    `json:"points"`
				TotalItems           int    `json:"totalItems"`
			} `json:"createContentAchievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateContentAchievement.ID)
		assert.Equal(t, "Test Content Achievement", result.CreateContentAchievement.Name)
		assert.Equal(t, 1, result.CreateContentAchievement.TotalItems)

		contentAchievementID = result.CreateContentAchievement.ID
	})

	t.Run("admin can create streak achievement", func(t *testing.T) {
		// Get an external content ID from seeded data
		contentResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetExternalContents($filter: ExternalContentFilter!) {
				externalContents(filter: $filter, first: 1) {
					edges {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"filter": map[string]any{},
		})
		require.False(t, contentResp.HasErrors(), "unexpected error: %s", contentResp.ErrorMessage())

		var contentResult struct {
			ExternalContents struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"externalContents"`
		}
		require.NoError(t, contentResp.UnmarshalData(&contentResult))

		if len(contentResult.ExternalContents.Edges) == 0 {
			t.Skip("No external content available for streak achievement test")
		}

		ecID := contentResult.ExternalContents.Edges[0].Node.ID

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateStreakAchievement($input: CreateStreakAchievementInput!) {
				createStreakAchievement(input: $input) {
					id
					name
					totalItems
					points
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Test Streak Achievement",
				"descriptionPending":   "Complete all content before deadlines",
				"descriptionCompleted": "You completed all content on time!",
				"notificationText":     "Streak completed!",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               75,
				"hidden":               false,
				"items": []map[string]any{
					{"externalContentId": ecID},
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateStreakAchievement struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				TotalItems int    `json:"totalItems"`
				Points     int    `json:"points"`
			} `json:"createStreakAchievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateStreakAchievement.ID)
		assert.Equal(t, "Test Streak Achievement", result.CreateStreakAchievement.Name)
		assert.Equal(t, 1, result.CreateStreakAchievement.TotalItems)
	})

	t.Run("user cannot create achievement", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Unauthorized Achievement",
				"descriptionPending":   "Should fail",
				"descriptionCompleted": "Should fail",
				"notificationText":     "Should fail",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               10,
				"hidden":               false,
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("admin can award achievement to user", func(t *testing.T) {
		if simpleAchievementID == "" {
			t.Skip("No simple achievement created")
		}

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AwardAchievement($userId: ID!, $achievementId: ID!) {
				awardAchievement(userId: $userId, achievementId: $achievementId) {
					... on SimpleAchievement {
						id
						name
					}
				}
			}
		`, map[string]any{
			"userId":        userID,
			"achievementId": simpleAchievementID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			AwardAchievement struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"awardAchievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, simpleAchievementID, result.AwardAchievement.ID)
	})

	t.Run("admin can revoke achievement from user", func(t *testing.T) {
		if simpleAchievementID == "" {
			t.Skip("No simple achievement created")
		}

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation RevokeAchievement($userId: ID!, $achievementId: ID!) {
				revokeAchievement(userId: $userId, achievementId: $achievementId)
			}
		`, map[string]any{
			"userId":        userID,
			"achievementId": simpleAchievementID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			RevokeAchievement bool `json:"revokeAchievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.True(t, result.RevokeAchievement)
	})

	t.Run("admin can bulk award achievements", func(t *testing.T) {
		t.Skip("BulkAwardAchievements resolver not yet implemented")
	})

	t.Run("m2m can mark content item completed", func(t *testing.T) {
		if contentAchievementID == "" {
			t.Skip("No content achievement created")
		}

		// Get an external content ID
		contentResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetExternalContents($filter: ExternalContentFilter!) {
				externalContents(filter: $filter, first: 1) {
					edges {
						node {
							id
						}
					}
				}
			}
		`, map[string]any{
			"filter": map[string]any{},
		})
		require.False(t, contentResp.HasErrors())

		var contentResult struct {
			ExternalContents struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"externalContents"`
		}
		require.NoError(t, contentResp.UnmarshalData(&contentResult))

		if len(contentResult.ExternalContents.Edges) == 0 {
			t.Skip("No external content available")
		}

		externalContentID := contentResult.ExternalContents.Edges[0].Node.ID

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation MarkContentItemCompleted($userId: ID!, $externalContentId: ID!) {
				markContentItemCompleted(userId: $userId, externalContentId: $externalContentId) {
					id
					name
				}
			}
		`, map[string]any{
			"userId":            userID,
			"externalContentId": externalContentID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			MarkContentItemCompleted []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"markContentItemCompleted"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		// May return empty if no achievements contain this content
	})

	t.Run("admin can update achievement", func(t *testing.T) {
		if simpleAchievementID == "" {
			t.Skip("No simple achievement created")
		}

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateAchievement($id: ID!, $input: UpdateAchievementInput!) {
				updateAchievement(id: $id, input: $input) {
					... on SimpleAchievement {
						id
						name
						points
					}
				}
			}
		`, map[string]any{
			"id": simpleAchievementID,
			"input": map[string]any{
				"name":   "Updated Achievement Name",
				"points": 150,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			UpdateAchievement struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Points int    `json:"points"`
			} `json:"updateAchievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, "Updated Achievement Name", result.UpdateAchievement.Name)
		assert.Equal(t, 150, result.UpdateAchievement.Points)
	})

	t.Run("admin can delete achievement", func(t *testing.T) {
		// Create a new achievement to delete
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "To Be Deleted",
				"descriptionPending":   "This will be deleted",
				"descriptionCompleted": "Never seen",
				"notificationText":     "Deleted",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               10,
				"hidden":               false,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateSimpleAchievement struct {
				ID string `json:"id"`
			} `json:"createSimpleAchievement"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		deleteID := createResult.CreateSimpleAchievement.ID

		// Delete it
		deleteResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation DeleteAchievement($id: ID!) {
				deleteAchievement(id: $id)
			}
		`, map[string]any{
			"id": deleteID,
		})

		require.False(t, deleteResp.HasErrors(), "unexpected error: %s", deleteResp.ErrorMessage())

		var deleteResult struct {
			DeleteAchievement bool `json:"deleteAchievement"`
		}
		require.NoError(t, deleteResp.UnmarshalData(&deleteResult))
		assert.True(t, deleteResult.DeleteAchievement)
	})

	t.Run("query achievement by id", func(t *testing.T) {
		if simpleAchievementID == "" {
			t.Skip("No simple achievement created")
		}

		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetAchievement($id: ID!) {
				achievement(id: $id) {
					... on SimpleAchievement {
						id
						name
						descriptionPending
						points
					}
				}
			}
		`, map[string]any{
			"id": simpleAchievementID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Achievement struct {
				ID                 string `json:"id"`
				Name               string `json:"name"`
				DescriptionPending string `json:"descriptionPending"`
				Points             int    `json:"points"`
			} `json:"achievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, simpleAchievementID, result.Achievement.ID)
	})

	t.Run("query achievements with filter", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetAchievements($filter: AchievementFilter!) {
				achievements(filter: $filter, first: 100) {
					edges {
						node {
							... on SimpleAchievement {
								id
								name
							}
							... on ContentAchievement {
								id
								name
							}
							... on StreakAchievement {
								id
								name
							}
							... on QuizAchievement {
								id
								name
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
			Achievements struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"achievements"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Greater(t, result.Achievements.TotalCount, 0)
	})

	t.Run("admin can reorder achievements", func(t *testing.T) {
		// Get existing achievements for the project
		listResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetAchievements($filter: AchievementFilter!) {
				achievements(filter: $filter, first: 10) {
					edges {
						node {
							... on SimpleAchievement { id }
							... on ContentAchievement { id }
							... on StreakAchievement { id }
							... on QuizAchievement { id }
						}
					}
				}
			}
		`, map[string]any{
			"filter": map[string]any{
				"projectId": projectID,
			},
		})
		require.False(t, listResp.HasErrors())

		var listResult struct {
			Achievements struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"achievements"`
		}
		require.NoError(t, listResp.UnmarshalData(&listResult))

		if len(listResult.Achievements.Edges) < 2 {
			t.Skip("Need at least 2 achievements to test reordering")
		}

		// Reverse the order
		achievementIDs := make([]string, len(listResult.Achievements.Edges))
		for i, edge := range listResult.Achievements.Edges {
			achievementIDs[len(listResult.Achievements.Edges)-1-i] = edge.Node.ID
		}

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation ReorderAchievements($projectId: ID!, $achievementIds: [ID!]!) {
				reorderAchievements(projectId: $projectId, achievementIds: $achievementIds) {
					... on SimpleAchievement { id }
					... on ContentAchievement { id }
					... on StreakAchievement { id }
					... on QuizAchievement { id }
				}
			}
		`, map[string]any{
			"projectId":      projectID,
			"achievementIds": achievementIDs,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			ReorderAchievements []struct {
				ID string `json:"id"`
			} `json:"reorderAchievements"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Len(t, result.ReorderAchievements, len(achievementIDs))
	})

	// Tests for awardableFrom functionality
	t.Run("admin can create achievement with future awardableFrom", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
					name
					awardableFrom
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Future Awardable Achievement",
				"descriptionPending":   "Not yet awardable",
				"descriptionCompleted": "Awarded!",
				"notificationText":     "You got it!",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               100,
				"hidden":               false,
				"awardableFrom":        futureTime,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateSimpleAchievement struct {
				ID            string  `json:"id"`
				Name          string  `json:"name"`
				AwardableFrom *string `json:"awardableFrom"`
			} `json:"createSimpleAchievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateSimpleAchievement.ID)
		assert.Equal(t, "Future Awardable Achievement", result.CreateSimpleAchievement.Name)
		assert.NotNil(t, result.CreateSimpleAchievement.AwardableFrom)
	})

	t.Run("awarding achievement with future awardableFrom fails", func(t *testing.T) {
		// Create an achievement with awardableFrom in the future
		futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Cannot Award Yet",
				"descriptionPending":   "Not yet awardable",
				"descriptionCompleted": "Awarded!",
				"notificationText":     "You got it!",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               50,
				"hidden":               false,
				"awardableFrom":        futureTime,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateSimpleAchievement struct {
				ID string `json:"id"`
			} `json:"createSimpleAchievement"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		futureAchievementID := createResult.CreateSimpleAchievement.ID

		// Try to award - should fail
		awardResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AwardAchievement($userId: ID!, $achievementId: ID!) {
				awardAchievement(userId: $userId, achievementId: $achievementId) {
					... on SimpleAchievement {
						id
					}
				}
			}
		`, map[string]any{
			"userId":        userID,
			"achievementId": futureAchievementID,
		})

		require.True(t, awardResp.HasErrors(), "expected error when awarding future achievement")
		assert.Contains(t, awardResp.ErrorMessage(), "not yet available")
	})

	t.Run("awarding achievement with past awardableFrom succeeds", func(t *testing.T) {
		// Create an achievement with awardableFrom in the past
		pastTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSimpleAchievement($input: CreateSimpleAchievementInput!) {
				createSimpleAchievement(input: $input) {
					id
					awardableFrom
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name":                 "Past Awardable Achievement",
				"descriptionPending":   "Already awardable",
				"descriptionCompleted": "Awarded!",
				"notificationText":     "You got it!",
				"imagePending":         "https://example.com/pending.png",
				"imageCompleted":       "https://example.com/completed.png",
				"projectId":            projectID,
				"points":               75,
				"hidden":               false,
				"awardableFrom":        pastTime,
			},
		})
		require.False(t, createResp.HasErrors(), "unexpected error: %s", createResp.ErrorMessage())

		var createResult struct {
			CreateSimpleAchievement struct {
				ID            string  `json:"id"`
				AwardableFrom *string `json:"awardableFrom"`
			} `json:"createSimpleAchievement"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		pastAchievementID := createResult.CreateSimpleAchievement.ID
		assert.NotNil(t, createResult.CreateSimpleAchievement.AwardableFrom)

		// Award should succeed
		awardResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AwardAchievement($userId: ID!, $achievementId: ID!) {
				awardAchievement(userId: $userId, achievementId: $achievementId) {
					... on SimpleAchievement {
						id
						name
					}
				}
			}
		`, map[string]any{
			"userId":        userID,
			"achievementId": pastAchievementID,
		})

		require.False(t, awardResp.HasErrors(), "unexpected error: %s", awardResp.ErrorMessage())

		var awardResult struct {
			AwardAchievement struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"awardAchievement"`
		}
		require.NoError(t, awardResp.UnmarshalData(&awardResult))
		assert.Equal(t, pastAchievementID, awardResult.AwardAchievement.ID)
	})

	t.Run("admin can update achievement awardableFrom", func(t *testing.T) {
		if simpleAchievementID == "" {
			t.Skip("No simple achievement created")
		}

		futureTime := time.Now().Add(48 * time.Hour).Format(time.RFC3339)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateAchievement($id: ID!, $input: UpdateAchievementInput!) {
				updateAchievement(id: $id, input: $input) {
					... on SimpleAchievement {
						id
						awardableFrom
					}
				}
			}
		`, map[string]any{
			"id": simpleAchievementID,
			"input": map[string]any{
				"awardableFrom": futureTime,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			UpdateAchievement struct {
				ID            string  `json:"id"`
				AwardableFrom *string `json:"awardableFrom"`
			} `json:"updateAchievement"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.NotNil(t, result.UpdateAchievement.AwardableFrom)
	})
}
