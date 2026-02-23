package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallenges(t *testing.T) {
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
	eventID := data.EventIDs[projectID][0]

	t.Run("admin can create SIMPLE challenge", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge {
						id
						name
						description
						buttonText
						allowSelfCompletion
					}
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":                "SIMPLE",
				"name":                "Simple Test Challenge",
				"description":         "<p>Test description</p>",
				"buttonText":          "Complete",
				"allowSelfCompletion": true,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateChallenge struct {
				ID                  string `json:"id"`
				Name                string `json:"name"`
				Description         string `json:"description"`
				ButtonText          string `json:"buttonText"`
				AllowSelfCompletion bool   `json:"allowSelfCompletion"`
			} `json:"createChallenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateChallenge.ID)
		assert.Equal(t, "Simple Test Challenge", result.CreateChallenge.Name)
		assert.Equal(t, "Complete", result.CreateChallenge.ButtonText)
		assert.True(t, result.CreateChallenge.AllowSelfCompletion)
	})

	t.Run("admin can create challenge without event", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, input: $input) {
					... on SimpleChallenge {
						id
						name
						event {
							id
						}
					}
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"input": map[string]any{
				"type":                "SIMPLE",
				"name":                "Project-Level Challenge",
				"description":         "<p>No event</p>",
				"buttonText":          "Complete",
				"allowSelfCompletion": true,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateChallenge struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Event *struct {
					ID string `json:"id"`
				} `json:"event"`
			} `json:"createChallenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateChallenge.ID)
		assert.Equal(t, "Project-Level Challenge", result.CreateChallenge.Name)
		assert.Nil(t, result.CreateChallenge.Event, "event should be nil for project-level challenge")
	})

	t.Run("admin can create EXTERNAL challenge", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on ExternalChallenge {
						id
						name
						url
						buttonText
					}
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":        "EXTERNAL",
				"name":        "External Test Challenge",
				"description": "<p>External challenge</p>",
				"buttonText":  "Open Link",
				"url":         "https://example.com/challenge",
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			CreateChallenge struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				URL        string `json:"url"`
				ButtonText string `json:"buttonText"`
			} `json:"createChallenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateChallenge.ID)
		assert.Equal(t, "External Test Challenge", result.CreateChallenge.Name)
		assert.Equal(t, "https://example.com/challenge", result.CreateChallenge.URL)
	})

	t.Run("user cannot create challenge", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge {
						id
					}
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":       "SIMPLE",
				"name":       "Unauthorized Challenge",
				"buttonText": "Complete",
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("admin can publish challenge", func(t *testing.T) {
		// First create a challenge
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge {
						id
						publishedAt
					}
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":       "SIMPLE",
				"name":       "To Be Published Challenge",
				"buttonText": "Complete",
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateChallenge struct {
				ID          string  `json:"id"`
				PublishedAt *string `json:"publishedAt"`
			} `json:"createChallenge"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		challengeID := createResult.CreateChallenge.ID
		assert.NotNil(t, createResult.CreateChallenge.PublishedAt)

		// Now publish it
		publishTime := time.Now().Format(time.RFC3339)
		publishResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation PublishChallenge($id: ID!, $publishedAt: DateTime!) {
				publishChallenge(id: $id, publishedAt: $publishedAt) {
					... on SimpleChallenge {
						id
						publishedAt
					}
				}
			}
		`, map[string]any{
			"id":          challengeID,
			"publishedAt": publishTime,
		})

		require.False(t, publishResp.HasErrors(), "unexpected error: %s", publishResp.ErrorMessage())

		var publishResult struct {
			PublishChallenge struct {
				ID          string  `json:"id"`
				PublishedAt *string `json:"publishedAt"`
			} `json:"publishChallenge"`
		}
		require.NoError(t, publishResp.UnmarshalData(&publishResult))
		assert.NotNil(t, publishResult.PublishChallenge.PublishedAt)
	})

	t.Run("admin can set challenge visibility", func(t *testing.T) {
		// Create a challenge
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge {
						id
					}
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":       "SIMPLE",
				"name":       "Visibility Test Challenge",
				"buttonText": "Complete",
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		challengeID := createResult.CreateChallenge.ID

		// Set visibility
		visibleAt := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
		startedAt := time.Now().Add(2 * time.Hour).Format(time.RFC3339)

		visibilityResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation SetChallengeVisibility($id: ID!, $visibleAt: DateTime!, $startedAt: DateTime) {
				setChallengeVisibility(id: $id, visibleAt: $visibleAt, startedAt: $startedAt) {
					... on SimpleChallenge {
						id
						visibleAt
						startedAt
					}
				}
			}
		`, map[string]any{
			"id":        challengeID,
			"visibleAt": visibleAt,
			"startedAt": startedAt,
		})

		require.False(t, visibilityResp.HasErrors(), "unexpected error: %s", visibilityResp.ErrorMessage())

		var visibilityResult struct {
			SetChallengeVisibility struct {
				ID        string  `json:"id"`
				VisibleAt *string `json:"visibleAt"`
				StartedAt *string `json:"startedAt"`
			} `json:"setChallengeVisibility"`
		}
		require.NoError(t, visibilityResp.UnmarshalData(&visibilityResult))
		assert.NotNil(t, visibilityResult.SetChallengeVisibility.VisibleAt)
		assert.NotNil(t, visibilityResult.SetChallengeVisibility.StartedAt)
	})

	// Create a published challenge with self-completion for enrollment/completion tests
	var selfCompleteChallengeID string
	{
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge {
						id
					}
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":                "SIMPLE",
				"name":                "Self Complete Challenge",
				"buttonText":          "Complete",
				"allowSelfCompletion": true,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		selfCompleteChallengeID = createResult.CreateChallenge.ID

		// Publish it
		publishTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		publishResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation PublishChallenge($id: ID!, $publishedAt: DateTime!) {
				publishChallenge(id: $id, publishedAt: $publishedAt) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"id":          selfCompleteChallengeID,
			"publishedAt": publishTime,
		})
		require.False(t, publishResp.HasErrors())
	}

	t.Run("user can enroll in challenge", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation EnrollInChallenge($challengeId: ID!) {
				enrollInChallenge(challengeId: $challengeId) {
					... on SimpleChallenge {
						id
						userEnrolledAt
					}
				}
			}
		`, map[string]any{
			"challengeId": selfCompleteChallengeID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			EnrollInChallenge struct {
				ID             string  `json:"id"`
				UserEnrolledAt *string `json:"userEnrolledAt"`
			} `json:"enrollInChallenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.NotNil(t, result.EnrollInChallenge.UserEnrolledAt)
	})

	t.Run("user can self-complete challenge", func(t *testing.T) {
		t.Skip("SelfCompleteChallenge resolver not yet implemented")
	})

	t.Run("user cannot self-complete when not allowed", func(t *testing.T) {
		t.Skip("SelfCompleteChallenge resolver not yet implemented")
	})

	t.Run("user can unenroll from challenge", func(t *testing.T) {
		// Create and publish a new challenge
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":       "SIMPLE",
				"name":       "Unenroll Test Challenge",
				"buttonText": "Complete",
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		challengeID := createResult.CreateChallenge.ID

		publishTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		publishResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation PublishChallenge($id: ID!, $publishedAt: DateTime!) {
				publishChallenge(id: $id, publishedAt: $publishedAt) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"id":          challengeID,
			"publishedAt": publishTime,
		})
		require.False(t, publishResp.HasErrors())

		// Enroll
		enrollResp := client.WithAuth(userToken).MustExecute(t, `
			mutation EnrollInChallenge($challengeId: ID!) {
				enrollInChallenge(challengeId: $challengeId) {
					... on SimpleChallenge { id userEnrolledAt }
				}
			}
		`, map[string]any{"challengeId": challengeID})
		require.False(t, enrollResp.HasErrors())

		// Unenroll
		unenrollResp := client.WithAuth(userToken).MustExecute(t, `
			mutation UnenrollFromChallenge($challengeId: ID!) {
				unenrollFromChallenge(challengeId: $challengeId)
			}
		`, map[string]any{"challengeId": challengeID})

		require.False(t, unenrollResp.HasErrors(), "unexpected error: %s", unenrollResp.ErrorMessage())

		var unenrollResult struct {
			UnenrollFromChallenge bool `json:"unenrollFromChallenge"`
		}
		require.NoError(t, unenrollResp.UnmarshalData(&unenrollResult))
		assert.True(t, unenrollResult.UnenrollFromChallenge)
	})

	t.Run("admin can complete challenge for user", func(t *testing.T) {
		// Create and publish an EXTERNAL challenge (not self-completable)
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on ExternalChallenge { id }
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":        "EXTERNAL",
				"name":        "Admin Complete Challenge",
				"buttonText":  "Open",
				"url":         "https://example.com",
				"description": "<p>Test</p>",
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		challengeID := createResult.CreateChallenge.ID

		publishTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		publishResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation PublishChallenge($id: ID!, $publishedAt: DateTime!) {
				publishChallenge(id: $id, publishedAt: $publishedAt) {
					... on ExternalChallenge { id }
				}
			}
		`, map[string]any{
			"id":          challengeID,
			"publishedAt": publishTime,
		})
		require.False(t, publishResp.HasErrors())

		// Admin completes for user
		completeResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CompleteChallenge($userId: ID!, $challengeId: ID!) {
				completeChallenge(userId: $userId, challengeId: $challengeId) {
					... on ExternalChallenge {
						id
					}
				}
			}
		`, map[string]any{
			"userId":      userID,
			"challengeId": challengeID,
		})

		require.False(t, completeResp.HasErrors(), "unexpected error: %s", completeResp.ErrorMessage())
	})

	t.Run("admin can bulk enroll users", func(t *testing.T) {
		// Create and publish a challenge
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":       "SIMPLE",
				"name":       "Bulk Enroll Challenge",
				"buttonText": "Complete",
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		challengeID := createResult.CreateChallenge.ID

		publishTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		publishResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation PublishChallenge($id: ID!, $publishedAt: DateTime!) {
				publishChallenge(id: $id, publishedAt: $publishedAt) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"id":          challengeID,
			"publishedAt": publishTime,
		})
		require.False(t, publishResp.HasErrors())

		// Bulk enroll multiple users
		userIDs := []string{data.UserIDs[0], data.UserIDs[2], data.UserIDs[3]}
		bulkEnrollResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation BulkEnrollUsersInChallenge($target: EnrollmentTargetInput!, $challengeId: ID!) {
				bulkEnrollUsersInChallenge(target: $target, challengeId: $challengeId) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"target": map[string]any{
				"userIds": userIDs,
			},
			"challengeId": challengeID,
		})

		require.False(t, bulkEnrollResp.HasErrors(), "unexpected error: %s", bulkEnrollResp.ErrorMessage())

		var bulkEnrollResult struct {
			BulkEnrollUsersInChallenge []struct {
				ID string `json:"id"`
			} `json:"bulkEnrollUsersInChallenge"`
		}
		require.NoError(t, bulkEnrollResp.UnmarshalData(&bulkEnrollResult))
		assert.Len(t, bulkEnrollResult.BulkEnrollUsersInChallenge, len(userIDs))
	})

	t.Run("admin can assign challenge to event", func(t *testing.T) {
		// Create a challenge without event assignment initially
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":       "SIMPLE",
				"name":       "Event Assignment Challenge",
				"buttonText": "Complete",
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		challengeID := createResult.CreateChallenge.ID

		// Get another event ID if available
		secondEventID := eventID
		if len(data.EventIDs[projectID]) > 1 {
			secondEventID = data.EventIDs[projectID][1]
		}

		// Assign to a different event
		assignResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AssignChallengeToEvent($challengeId: ID!, $eventId: ID!) {
				assignChallengeToEvent(challengeId: $challengeId, eventId: $eventId) {
					... on SimpleChallenge {
						id
						event {
							id
						}
					}
				}
			}
		`, map[string]any{
			"challengeId": challengeID,
			"eventId":     secondEventID,
		})

		require.False(t, assignResp.HasErrors(), "unexpected error: %s", assignResp.ErrorMessage())

		var assignResult struct {
			AssignChallengeToEvent struct {
				ID    string `json:"id"`
				Event struct {
					ID string `json:"id"`
				} `json:"event"`
			} `json:"assignChallengeToEvent"`
		}
		require.NoError(t, assignResp.UnmarshalData(&assignResult))
		assert.Equal(t, secondEventID, assignResult.AssignChallengeToEvent.Event.ID)
	})

	t.Run("query challenge by id", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					... on SimpleChallenge {
						id
						name
						description
						buttonText
						publishedAt
					}
				}
			}
		`, map[string]any{
			"id": selfCompleteChallengeID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Challenge struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description string  `json:"description"`
				ButtonText  string  `json:"buttonText"`
				PublishedAt *string `json:"publishedAt"`
			} `json:"challenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, selfCompleteChallengeID, result.Challenge.ID)
		assert.NotEmpty(t, result.Challenge.Name)
	})

	t.Run("query challenges with filter", func(t *testing.T) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetChallenges($filter: ChallengeFilter) {
				challenges(filter: $filter, first: 100) {
					edges {
						node {
							... on SimpleChallenge {
								id
								name
							}
							... on ExternalChallenge {
								id
								name
							}
							... on QuizChallenge {
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
			Challenges struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"challenges"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Greater(t, result.Challenges.TotalCount, 0)
	})
}

func TestChallengeVisibility(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup user IDs
	userID := data.UserIDs[0]
	adminUserID := data.UserIDs[1]
	otherUserID := data.UserIDs[2]

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

	otherUserToken, err := testutil.GenerateUserToken(otherUserID)
	require.NoError(t, err)

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]

	// Helper to create a challenge with specific visibility settings
	createChallenge := func(t *testing.T, name string) string {
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":       "SIMPLE",
				"name":       name,
				"buttonText": "Complete",
			},
		})
		require.False(t, createResp.HasErrors())

		var result struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, createResp.UnmarshalData(&result))
		return result.CreateChallenge.ID
	}

	// Helper to publish a challenge
	publishChallenge := func(t *testing.T, challengeID string) {
		publishTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		publishResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation PublishChallenge($id: ID!, $publishedAt: DateTime!) {
				publishChallenge(id: $id, publishedAt: $publishedAt) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"id":          challengeID,
			"publishedAt": publishTime,
		})
		require.False(t, publishResp.HasErrors())
	}

	// Helper to set visibility
	setVisibility := func(t *testing.T, challengeID string, visibleAt *string) {
		if visibleAt == nil {
			return
		}
		visibilityResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation SetChallengeVisibility($id: ID!, $visibleAt: DateTime!) {
				setChallengeVisibility(id: $id, visibleAt: $visibleAt) {
					... on SimpleChallenge { id visibleAt }
				}
			}
		`, map[string]any{
			"id":        challengeID,
			"visibleAt": *visibleAt,
		})
		require.False(t, visibilityResp.HasErrors(), "unexpected error: %s", visibilityResp.ErrorMessage())
	}

	t.Run("unenrolled user cannot see challenge with future visible_at", func(t *testing.T) {
		// Create and publish a challenge with future visible_at
		challengeID := createChallenge(t, "Future Visibility Challenge")
		publishChallenge(t, challengeID)

		futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		setVisibility(t, challengeID, &futureTime)

		// Query challenge directly - should fail
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					... on SimpleChallenge { id name }
				}
			}
		`, map[string]any{"id": challengeID})

		require.True(t, resp.HasErrors(), "expected error for unenrolled user accessing future visible_at challenge")
		assert.Contains(t, resp.ErrorMessage(), "not found")
	})

	t.Run("unenrolled user cannot see challenge with NULL visible_at", func(t *testing.T) {
		// Create and publish a challenge (without setting visible_at, it stays NULL)
		challengeID := createChallenge(t, "NULL Visibility Challenge")
		publishChallenge(t, challengeID)
		// No setVisibility call - visible_at remains NULL

		// Query challenge directly - should fail
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					... on SimpleChallenge { id name }
				}
			}
		`, map[string]any{"id": challengeID})

		require.True(t, resp.HasErrors(), "expected error for unenrolled user accessing NULL visible_at challenge")
		assert.Contains(t, resp.ErrorMessage(), "not found")
	})

	t.Run("enrolled user CAN see challenge with future visible_at", func(t *testing.T) {
		// Create and publish a challenge with future visible_at
		challengeID := createChallenge(t, "Enrolled Future Visibility")
		publishChallenge(t, challengeID)

		futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		setVisibility(t, challengeID, &futureTime)

		// Enroll the user (admin can bulk enroll)
		enrollResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation BulkEnrollUsersInChallenge($target: EnrollmentTargetInput!, $challengeId: ID!) {
				bulkEnrollUsersInChallenge(target: $target, challengeId: $challengeId) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"target":      map[string]any{"userIds": []string{userID}},
			"challengeId": challengeID,
		})
		require.False(t, enrollResp.HasErrors())

		// Now user should be able to see it
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					... on SimpleChallenge { id name }
				}
			}
		`, map[string]any{"id": challengeID})

		require.False(t, resp.HasErrors(), "enrolled user should see future visible_at challenge: %s", resp.ErrorMessage())

		var result struct {
			Challenge struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"challenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, challengeID, result.Challenge.ID)
		assert.Equal(t, "Enrolled Future Visibility", result.Challenge.Name)
	})

	t.Run("enrolled user CAN see challenge with NULL visible_at", func(t *testing.T) {
		// Create and publish a challenge (without setting visible_at)
		challengeID := createChallenge(t, "Enrolled NULL Visibility")
		publishChallenge(t, challengeID)
		// No setVisibility call - visible_at remains NULL

		// Enroll the user
		enrollResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation BulkEnrollUsersInChallenge($target: EnrollmentTargetInput!, $challengeId: ID!) {
				bulkEnrollUsersInChallenge(target: $target, challengeId: $challengeId) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"target":      map[string]any{"userIds": []string{userID}},
			"challengeId": challengeID,
		})
		require.False(t, enrollResp.HasErrors())

		// Now user should be able to see it
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					... on SimpleChallenge { id name }
				}
			}
		`, map[string]any{"id": challengeID})

		require.False(t, resp.HasErrors(), "enrolled user should see NULL visible_at challenge: %s", resp.ErrorMessage())

		var result struct {
			Challenge struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"challenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, challengeID, result.Challenge.ID)
		assert.Equal(t, "Enrolled NULL Visibility", result.Challenge.Name)
	})

	t.Run("user CAN see challenge with past visible_at regardless of enrollment", func(t *testing.T) {
		// Create and publish a challenge with past visible_at
		challengeID := createChallenge(t, "Past Visibility Challenge")
		publishChallenge(t, challengeID)

		pastTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
		setVisibility(t, challengeID, &pastTime)

		// User (not enrolled) should be able to see it
		resp := client.WithAuth(otherUserToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					... on SimpleChallenge { id name }
				}
			}
		`, map[string]any{"id": challengeID})

		require.False(t, resp.HasErrors(), "user should see past visible_at challenge: %s", resp.ErrorMessage())

		var result struct {
			Challenge struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"challenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, challengeID, result.Challenge.ID)
		assert.Equal(t, "Past Visibility Challenge", result.Challenge.Name)
	})

	t.Run("admin can see ALL challenges regardless of visibility", func(t *testing.T) {
		// Create challenges with various visibility states
		futureChallengeID := createChallenge(t, "Admin Test Future")
		publishChallenge(t, futureChallengeID)
		futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		setVisibility(t, futureChallengeID, &futureTime)

		nullChallengeID := createChallenge(t, "Admin Test NULL")
		publishChallenge(t, nullChallengeID)
		// No visibility set - stays NULL

		// Admin should see both
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					... on SimpleChallenge { id name }
				}
			}
		`, map[string]any{"id": futureChallengeID})
		require.False(t, resp.HasErrors(), "admin should see future visible_at challenge: %s", resp.ErrorMessage())

		resp = client.WithAuth(adminToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					... on SimpleChallenge { id name }
				}
			}
		`, map[string]any{"id": nullChallengeID})
		require.False(t, resp.HasErrors(), "admin should see NULL visible_at challenge: %s", resp.ErrorMessage())
	})

	t.Run("challenge list includes enrolled-but-not-visible challenges", func(t *testing.T) {
		// Create a challenge with NULL visible_at (not visible to unenrolled)
		challengeID := createChallenge(t, "List Test Enrolled")
		publishChallenge(t, challengeID)
		// No visibility set - stays NULL

		// Enroll user
		enrollResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation BulkEnrollUsersInChallenge($target: EnrollmentTargetInput!, $challengeId: ID!) {
				bulkEnrollUsersInChallenge(target: $target, challengeId: $challengeId) {
					... on SimpleChallenge { id }
				}
			}
		`, map[string]any{
			"target":      map[string]any{"userIds": []string{userID}},
			"challengeId": challengeID,
		})
		require.False(t, enrollResp.HasErrors())

		// Query project challenges - enrolled user should see it
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetProject($id: ID!) {
				project(id: $id) {
					challenges {
						... on SimpleChallenge { id name }
					}
				}
			}
		`, map[string]any{"id": projectID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Project struct {
				Challenges []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"challenges"`
			} `json:"project"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Find our challenge in the list
		found := false
		for _, ch := range result.Project.Challenges {
			if ch.ID == challengeID {
				found = true
				assert.Equal(t, "List Test Enrolled", ch.Name)
				break
			}
		}
		assert.True(t, found, "enrolled user should see enrolled challenge in project challenges list")
	})

	t.Run("challenge list excludes non-visible challenges for unenrolled user", func(t *testing.T) {
		// Create a challenge with NULL visible_at
		challengeID := createChallenge(t, "List Test Not Visible")
		publishChallenge(t, challengeID)
		// No visibility set - stays NULL
		// No enrollment for otherUser

		// Query project challenges - other user should NOT see it
		resp := client.WithAuth(otherUserToken).MustExecute(t, `
			query GetProject($id: ID!) {
				project(id: $id) {
					challenges {
						... on SimpleChallenge { id name }
					}
				}
			}
		`, map[string]any{"id": projectID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			Project struct {
				Challenges []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"challenges"`
			} `json:"project"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		// Should NOT find the challenge (it has NULL visible_at and user is not enrolled)
		for _, ch := range result.Project.Challenges {
			if ch.ID == challengeID {
				t.Errorf("unenrolled user should NOT see challenge with NULL visible_at in list")
			}
		}
	})
}
