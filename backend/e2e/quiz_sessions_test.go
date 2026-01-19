package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuizSessions(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup user IDs
	userID := data.UserIDs[0]
	user2ID := data.UserIDs[1]
	adminUserID := data.UserIDs[2]

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
	teamID := data.TeamIDs[projectID][0]
	churchID := data.ChurchIDs[0]

	// Helper to create a challenge
	createChallenge := func(t *testing.T, name string) string {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID!, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input": map[string]any{
				"type":        "QUIZ",
				"name":        name,
				"buttonText":  "Take Quiz",
				"description": "<p>Test challenge</p>",
			},
		})
		require.False(t, resp.HasErrors(), "failed to create challenge: %s", resp.ErrorMessage())

		var result struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.CreateChallenge.ID
	}

	// Helper to create a quiz
	createQuiz := func(t *testing.T, name string, challengeID string) string {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateQuiz($input: CreateQuizInput!) {
				createQuiz(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":            projectID,
				"challengeId":          challengeID,
				"name":                 name,
				"description":          "Test quiz for sessions",
				"randomizeQuestions":   false,
				"revealCorrectAnswers": true,
				"allowRetakes":         false,
				"completionPoints":     10,
			},
		})
		require.False(t, resp.HasErrors(), "failed to create quiz: %s", resp.ErrorMessage())

		var result struct {
			CreateQuiz struct{ ID string } `json:"createQuiz"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.CreateQuiz.ID
	}

	// Helper to publish a challenge
	publishChallenge := func(t *testing.T, challengeID string) {
		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation PublishChallenge($id: ID!, $publishedAt: DateTime!) {
				publishChallenge(id: $id, publishedAt: $publishedAt) {
					id
				}
			}
		`, map[string]any{
			"id":          challengeID,
			"publishedAt": publishedTime,
		})
		require.False(t, resp.HasErrors(), "failed to publish challenge: %s", resp.ErrorMessage())
	}

	// Helper to add a question to quiz
	addQuestion := func(t *testing.T, quizID string, questionText string) string {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on PredefinedQuestion {
						id
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":  "PREDEFINED",
				"questionText":  questionText,
				"questionOrder": 0,
				"points":        5,
				"predefinedAnswers": []map[string]any{
					{"answerText": "Correct", "isCorrect": true, "answerOrder": 0},
					{"answerText": "Wrong", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion struct{ ID string } `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AddQuizQuestion.ID
	}

	// ==================== SESSION CRUD TESTS ====================

	t.Run("admin can create quiz session", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Create Challenge")
		quizID := createQuiz(t, "Session Create Quiz", challengeID)
		addQuestion(t, quizID, "Question 1?")

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
					name
					state
					openAt
					lockAt
					finishAt
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
				"name":   "Test Session",
			},
		})

		require.False(t, resp.HasErrors(), "failed to create session: %s", resp.ErrorMessage())

		var result struct {
			CreateQuizSession struct {
				ID       string  `json:"id"`
				Name     *string `json:"name"`
				State    string  `json:"state"`
				OpenAt   *string `json:"openAt"`
				LockAt   *string `json:"lockAt"`
				FinishAt *string `json:"finishAt"`
			} `json:"createQuizSession"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.CreateQuizSession.ID)
		assert.Equal(t, "Test Session", *result.CreateQuizSession.Name)
		assert.Equal(t, "DRAFT", result.CreateQuizSession.State)
	})

	t.Run("user cannot create quiz session", func(t *testing.T) {
		challengeID := createChallenge(t, "Session User Create Challenge")
		quizID := createQuiz(t, "Session User Create Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
				"name":   "User Session",
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "unauthorized")
	})

	t.Run("can update session in DRAFT state", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Update Challenge")
		quizID := createQuiz(t, "Session Update Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
					name
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
				"name":   "Original Name",
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Update session
		updateResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateSession($id: ID!, $input: UpdateQuizSessionInput!) {
				updateQuizSession(id: $id, input: $input) {
					id
					name
				}
			}
		`, map[string]any{
			"id": sessionID,
			"input": map[string]any{
				"name": "Updated Name",
			},
		})
		require.False(t, updateResp.HasErrors(), "failed to update: %s", updateResp.ErrorMessage())

		var updateResult struct {
			UpdateQuizSession struct {
				ID   string  `json:"id"`
				Name *string `json:"name"`
			} `json:"updateQuizSession"`
		}
		require.NoError(t, updateResp.UnmarshalData(&updateResult))

		assert.Equal(t, "Updated Name", *updateResult.UpdateQuizSession.Name)
	})

	t.Run("can delete session in DRAFT state", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Delete Challenge")
		quizID := createQuiz(t, "Session Delete Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Delete session
		deleteResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation DeleteSession($id: ID!) {
				deleteQuizSession(id: $id)
			}
		`, map[string]any{"id": sessionID})
		require.False(t, deleteResp.HasErrors(), "failed to delete: %s", deleteResp.ErrorMessage())

		var deleteResult struct {
			DeleteQuizSession bool `json:"deleteQuizSession"`
		}
		require.NoError(t, deleteResp.UnmarshalData(&deleteResult))
		assert.True(t, deleteResult.DeleteQuizSession)
	})

	// ==================== STATE TRANSITION TESTS ====================

	t.Run("session state transitions", func(t *testing.T) {
		challengeID := createChallenge(t, "Session State Challenge")
		quizID := createQuiz(t, "Session State Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
					state
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID
		assert.Equal(t, "DRAFT", createResult.CreateQuizSession.State)

		// DRAFT -> OPEN
		openResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) {
					id
					state
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, openResp.HasErrors(), "failed to open: %s", openResp.ErrorMessage())

		var openResult struct {
			OpenQuizSession struct{ State string } `json:"openQuizSession"`
		}
		require.NoError(t, openResp.UnmarshalData(&openResult))
		assert.Equal(t, "OPEN", openResult.OpenQuizSession.State)

		// OPEN -> LOCKED
		lockResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation LockSession($id: ID!) {
				lockQuizSession(id: $id) {
					id
					state
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, lockResp.HasErrors(), "failed to lock: %s", lockResp.ErrorMessage())

		var lockResult struct {
			LockQuizSession struct{ State string } `json:"lockQuizSession"`
		}
		require.NoError(t, lockResp.UnmarshalData(&lockResult))
		assert.Equal(t, "LOCKED", lockResult.LockQuizSession.State)

		// LOCKED -> FINISHED
		finishResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation FinishSession($id: ID!) {
				finishQuizSession(id: $id) {
					id
					state
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, finishResp.HasErrors(), "failed to finish: %s", finishResp.ErrorMessage())

		var finishResult struct {
			FinishQuizSession struct{ State string } `json:"finishQuizSession"`
		}
		require.NoError(t, finishResp.UnmarshalData(&finishResult))
		assert.Equal(t, "FINISHED", finishResult.FinishQuizSession.State)
	})

	t.Run("cannot update session after DRAFT state", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Update After Open Challenge")
		quizID := createQuiz(t, "Session Update After Open Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create and open session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open session
		openResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) {
					id
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, openResp.HasErrors())

		// Try to update - should fail
		updateResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateSession($id: ID!, $input: UpdateQuizSessionInput!) {
				updateQuizSession(id: $id, input: $input) {
					id
				}
			}
		`, map[string]any{
			"id": sessionID,
			"input": map[string]any{
				"name": "Should Fail",
			},
		})

		require.True(t, updateResp.HasErrors())
		assert.Contains(t, updateResp.ErrorMessage(), "DRAFT")
	})

	t.Run("cannot delete session after DRAFT state", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Delete After Open Challenge")
		quizID := createQuiz(t, "Session Delete After Open Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create and open session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open session
		openResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) {
					id
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, openResp.HasErrors())

		// Try to delete - should fail
		deleteResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation DeleteSession($id: ID!) {
				deleteQuizSession(id: $id)
			}
		`, map[string]any{"id": sessionID})

		require.True(t, deleteResp.HasErrors())
		assert.Contains(t, deleteResp.ErrorMessage(), "DRAFT")
	})

	t.Run("reopen session from LOCKED state", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Reopen Challenge")
		quizID := createQuiz(t, "Session Reopen Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open and lock
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		client.WithAuth(adminToken).MustExecute(t, `
			mutation LockSession($id: ID!) {
				lockQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Reopen
		reopenResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation ReopenSession($id: ID!) {
				reopenQuizSession(id: $id) {
					id
					state
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, reopenResp.HasErrors(), "failed to reopen: %s", reopenResp.ErrorMessage())

		var reopenResult struct {
			ReopenQuizSession struct{ State string } `json:"reopenQuizSession"`
		}
		require.NoError(t, reopenResp.UnmarshalData(&reopenResult))
		assert.Equal(t, "OPEN", reopenResult.ReopenQuizSession.State)
	})

	t.Run("invalid state transitions fail", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Invalid Transition Challenge")
		quizID := createQuiz(t, "Session Invalid Transition Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Try to lock from DRAFT (should fail)
		lockResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation LockSession($id: ID!) {
				lockQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})
		require.True(t, lockResp.HasErrors())
		assert.Contains(t, lockResp.ErrorMessage(), "OPEN")

		// Try to finish from DRAFT (should fail)
		finishResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation FinishSession($id: ID!) {
				finishQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})
		require.True(t, finishResp.HasErrors())
		assert.Contains(t, finishResp.ErrorMessage(), "LOCKED")
	})

	// ==================== ACCESS MANAGEMENT TESTS ====================

	t.Run("grant access to individual users", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Grant User Access Challenge")
		quizID := createQuiz(t, "Session Grant User Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
					accessCount
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct {
				ID          string `json:"id"`
				AccessCount int    `json:"accessCount"`
			} `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID
		assert.Equal(t, 0, createResult.CreateQuizSession.AccessCount)

		// Grant access to 2 users
		grantResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID, user2ID},
			},
		})
		require.False(t, grantResp.HasErrors(), "failed to grant: %s", grantResp.ErrorMessage())

		var grantResult struct {
			GrantQuizSessionAccess int `json:"grantQuizSessionAccess"`
		}
		require.NoError(t, grantResp.UnmarshalData(&grantResult))
		assert.Equal(t, 2, grantResult.GrantQuizSessionAccess)

		// Query session to verify access count
		queryResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetSession($id: ID!) {
				quizSession(id: $id) {
					id
					accessCount
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			QuizSession struct {
				AccessCount int `json:"accessCount"`
			} `json:"quizSession"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))
		assert.Equal(t, 2, queryResult.QuizSession.AccessCount)
	})

	t.Run("grant access to team", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Grant Team Access Challenge")
		quizID := createQuiz(t, "Session Grant Team Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access to team
		grantResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"teamIds":   []string{teamID},
			},
		})
		require.False(t, grantResp.HasErrors(), "failed to grant: %s", grantResp.ErrorMessage())

		var grantResult struct {
			GrantQuizSessionAccess int `json:"grantQuizSessionAccess"`
		}
		require.NoError(t, grantResp.UnmarshalData(&grantResult))
		assert.Greater(t, grantResult.GrantQuizSessionAccess, 0)
	})

	t.Run("grant access to church", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Grant Church Access Challenge")
		quizID := createQuiz(t, "Session Grant Church Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access to church
		grantResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"churchIds": []string{churchID},
			},
		})
		require.False(t, grantResp.HasErrors(), "failed to grant: %s", grantResp.ErrorMessage())

		var grantResult struct {
			GrantQuizSessionAccess int `json:"grantQuizSessionAccess"`
		}
		require.NoError(t, grantResp.UnmarshalData(&grantResult))
		assert.Greater(t, grantResult.GrantQuizSessionAccess, 0)
	})

	t.Run("revoke user access", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Revoke Access Challenge")
		quizID := createQuiz(t, "Session Revoke Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session and grant access
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID, user2ID},
			},
		})

		// Revoke access for one user
		revokeResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation RevokeAccess($sessionId: ID!, $userIds: [ID!]!) {
				revokeQuizSessionAccess(sessionId: $sessionId, userIds: $userIds)
			}
		`, map[string]any{
			"sessionId": sessionID,
			"userIds":   []string{userID},
		})
		require.False(t, revokeResp.HasErrors(), "failed to revoke: %s", revokeResp.ErrorMessage())

		var revokeResult struct {
			RevokeQuizSessionAccess bool `json:"revokeQuizSessionAccess"`
		}
		require.NoError(t, revokeResp.UnmarshalData(&revokeResult))
		assert.True(t, revokeResult.RevokeQuizSessionAccess)

		// Verify access count
		queryResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetSession($id: ID!) {
				quizSession(id: $id) {
					accessCount
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			QuizSession struct {
				AccessCount int `json:"accessCount"`
			} `json:"quizSession"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))
		assert.Equal(t, 1, queryResult.QuizSession.AccessCount)
	})

	// ==================== USER ACTIONS TESTS ====================

	t.Run("user can start session with access", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Start With Access Challenge")
		quizID := createQuiz(t, "Session Start With Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access to user
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// User starts session
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) {
					id
					startedAt
				}
			}
		`, map[string]any{"sessionId": sessionID})
		require.False(t, startResp.HasErrors(), "failed to start: %s", startResp.ErrorMessage())

		var startResult struct {
			StartQuizSession struct {
				ID        string `json:"id"`
				StartedAt string `json:"startedAt"`
			} `json:"startQuizSession"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		assert.NotEmpty(t, startResult.StartQuizSession.ID)
		assert.NotEmpty(t, startResult.StartQuizSession.StartedAt)
	})

	t.Run("user cannot start session without access", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Start No Access Challenge")
		quizID := createQuiz(t, "Session Start No Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create and open session without granting access
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// User tries to start without access
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) {
					id
				}
			}
		`, map[string]any{"sessionId": sessionID})

		require.True(t, startResp.HasErrors())
		assert.Contains(t, startResp.ErrorMessage(), "access")
	})

	t.Run("user cannot start session when not OPEN", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Start Not Open Challenge")
		quizID := createQuiz(t, "Session Start Not Open Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session and grant access but don't open
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		// User tries to start in DRAFT state
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) {
					id
				}
			}
		`, map[string]any{"sessionId": sessionID})

		require.True(t, startResp.HasErrors())
		assert.Contains(t, startResp.ErrorMessage(), "not open")
	})

	t.Run("starting session again returns existing submission", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Resume Challenge")
		quizID := createQuiz(t, "Session Resume Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session, grant access, and open
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Start first time
		start1Resp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) {
					id
					startedAt
				}
			}
		`, map[string]any{"sessionId": sessionID})
		require.False(t, start1Resp.HasErrors())

		var start1Result struct {
			StartQuizSession struct {
				ID        string `json:"id"`
				StartedAt string `json:"startedAt"`
			} `json:"startQuizSession"`
		}
		require.NoError(t, start1Resp.UnmarshalData(&start1Result))
		firstSubmissionID := start1Result.StartQuizSession.ID
		firstStartedAt := start1Result.StartQuizSession.StartedAt

		// Start second time
		start2Resp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) {
					id
					startedAt
				}
			}
		`, map[string]any{"sessionId": sessionID})
		require.False(t, start2Resp.HasErrors())

		var start2Result struct {
			StartQuizSession struct {
				ID        string `json:"id"`
				StartedAt string `json:"startedAt"`
			} `json:"startQuizSession"`
		}
		require.NoError(t, start2Resp.UnmarshalData(&start2Result))

		assert.Equal(t, firstSubmissionID, start2Result.StartQuizSession.ID, "should return same submission")
		assert.Equal(t, firstStartedAt, start2Result.StartQuizSession.StartedAt)
	})

	t.Run("reset submission while session is open", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Reset Challenge")
		quizID := createQuiz(t, "Session Reset Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session, grant access, and open
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Start session
		start1Resp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) {
					id
				}
			}
		`, map[string]any{"sessionId": sessionID})
		require.False(t, start1Resp.HasErrors())

		var start1Result struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		require.NoError(t, start1Resp.UnmarshalData(&start1Result))
		firstSubmissionID := start1Result.StartQuizSession.ID

		// Reset submission
		resetResp := client.WithAuth(userToken).MustExecute(t, `
			mutation ResetSubmission($sessionId: ID!) {
				resetQuizSessionSubmission(sessionId: $sessionId)
			}
		`, map[string]any{"sessionId": sessionID})
		require.False(t, resetResp.HasErrors(), "failed to reset: %s", resetResp.ErrorMessage())

		var resetResult struct {
			ResetQuizSessionSubmission bool `json:"resetQuizSessionSubmission"`
		}
		require.NoError(t, resetResp.UnmarshalData(&resetResult))
		assert.True(t, resetResult.ResetQuizSessionSubmission)

		// Start again - should get new submission
		start2Resp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) {
					id
				}
			}
		`, map[string]any{"sessionId": sessionID})
		require.False(t, start2Resp.HasErrors())

		var start2Result struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		require.NoError(t, start2Resp.UnmarshalData(&start2Result))

		assert.NotEqual(t, firstSubmissionID, start2Result.StartQuizSession.ID, "should get new submission after reset")
	})

	t.Run("cannot reset submission when session is locked", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Reset Locked Challenge")
		quizID := createQuiz(t, "Session Reset Locked Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session, grant access, open and lock
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Start session
		client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) { id }
			}
		`, map[string]any{"sessionId": sessionID})

		// Lock session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation LockSession($id: ID!) {
				lockQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Try to reset - should fail
		resetResp := client.WithAuth(userToken).MustExecute(t, `
			mutation ResetSubmission($sessionId: ID!) {
				resetQuizSessionSubmission(sessionId: $sessionId)
			}
		`, map[string]any{"sessionId": sessionID})

		require.True(t, resetResp.HasErrors())
		assert.Contains(t, resetResp.ErrorMessage(), "open")
	})

	// ==================== QUERY TESTS ====================

	t.Run("query sessions by quiz", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Query Challenge")
		quizID := createQuiz(t, "Session Query Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create multiple sessions
		for range 3 {
			client.WithAuth(adminToken).MustExecute(t, `
				mutation CreateSession($input: CreateQuizSessionInput!) {
					createQuizSession(input: $input) {
						id
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"quizId": quizID,
				},
			})
		}

		// Query sessions
		queryResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetSessions($quizId: ID!) {
				quizSessions(quizId: $quizId) {
					id
					state
				}
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			QuizSessions []struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"quizSessions"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))
		assert.Len(t, queryResult.QuizSessions, 3)
	})

	t.Run("query sessions filtered by state", func(t *testing.T) {
		challengeID := createChallenge(t, "Session Filter State Challenge")
		quizID := createQuiz(t, "Session Filter State Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create sessions in different states
		var session1ID, session2ID string

		resp1 := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, resp1.HasErrors())
		var result1 struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, resp1.UnmarshalData(&result1))
		session1ID = result1.CreateQuizSession.ID

		resp2 := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, resp2.HasErrors())
		var result2 struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, resp2.UnmarshalData(&result2))
		session2ID = result2.CreateQuizSession.ID

		// Open first session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": session1ID})

		// Query only OPEN sessions
		queryOpenResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetSessions($quizId: ID!, $state: QuizSessionState) {
				quizSessions(quizId: $quizId, state: $state) {
					id
					state
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"state":  "OPEN",
		})
		require.False(t, queryOpenResp.HasErrors())

		var queryOpenResult struct {
			QuizSessions []struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"quizSessions"`
		}
		require.NoError(t, queryOpenResp.UnmarshalData(&queryOpenResult))
		assert.Len(t, queryOpenResult.QuizSessions, 1)
		assert.Equal(t, session1ID, queryOpenResult.QuizSessions[0].ID)

		// Query only DRAFT sessions
		queryDraftResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetSessions($quizId: ID!, $state: QuizSessionState) {
				quizSessions(quizId: $quizId, state: $state) {
					id
					state
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"state":  "DRAFT",
		})
		require.False(t, queryDraftResp.HasErrors())

		var queryDraftResult struct {
			QuizSessions []struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"quizSessions"`
		}
		require.NoError(t, queryDraftResp.UnmarshalData(&queryDraftResult))
		assert.Len(t, queryDraftResult.QuizSessions, 1)
		assert.Equal(t, session2ID, queryDraftResult.QuizSessions[0].ID)
	})

	t.Run("user can see userHasAccess and userSubmission fields", func(t *testing.T) {
		challengeID := createChallenge(t, "Session User Fields Challenge")
		quizID := createQuiz(t, "Session User Fields Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// User without access queries session
		queryNoAccessResp := client.WithAuth(userToken).MustExecute(t, `
			query GetSession($id: ID!) {
				quizSession(id: $id) {
					id
					userHasAccess
					userSubmission {
						id
					}
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, queryNoAccessResp.HasErrors())

		var queryNoAccessResult struct {
			QuizSession struct {
				ID             string `json:"id"`
				UserHasAccess  bool   `json:"userHasAccess"`
				UserSubmission *struct {
					ID string `json:"id"`
				} `json:"userSubmission"`
			} `json:"quizSession"`
		}
		require.NoError(t, queryNoAccessResp.UnmarshalData(&queryNoAccessResult))
		assert.False(t, queryNoAccessResult.QuizSession.UserHasAccess)
		assert.Nil(t, queryNoAccessResult.QuizSession.UserSubmission)

		// Grant access
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// User with access queries session
		queryWithAccessResp := client.WithAuth(userToken).MustExecute(t, `
			query GetSession($id: ID!) {
				quizSession(id: $id) {
					id
					userHasAccess
					userSubmission {
						id
					}
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, queryWithAccessResp.HasErrors())

		var queryWithAccessResult struct {
			QuizSession struct {
				ID             string `json:"id"`
				UserHasAccess  bool   `json:"userHasAccess"`
				UserSubmission *struct {
					ID string `json:"id"`
				} `json:"userSubmission"`
			} `json:"quizSession"`
		}
		require.NoError(t, queryWithAccessResp.UnmarshalData(&queryWithAccessResult))
		assert.True(t, queryWithAccessResult.QuizSession.UserHasAccess)
		assert.Nil(t, queryWithAccessResult.QuizSession.UserSubmission) // Not started yet

		// Start session
		client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) { id }
			}
		`, map[string]any{"sessionId": sessionID})

		// Query again - should see submission
		queryWithSubmissionResp := client.WithAuth(userToken).MustExecute(t, `
			query GetSession($id: ID!) {
				quizSession(id: $id) {
					id
					userHasAccess
					userSubmission {
						id
					}
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, queryWithSubmissionResp.HasErrors())

		var queryWithSubmissionResult struct {
			QuizSession struct {
				ID             string `json:"id"`
				UserHasAccess  bool   `json:"userHasAccess"`
				UserSubmission *struct {
					ID string `json:"id"`
				} `json:"userSubmission"`
			} `json:"quizSession"`
		}
		require.NoError(t, queryWithSubmissionResp.UnmarshalData(&queryWithSubmissionResult))
		assert.True(t, queryWithSubmissionResult.QuizSession.UserHasAccess)
		assert.NotNil(t, queryWithSubmissionResult.QuizSession.UserSubmission)
		assert.NotEmpty(t, queryWithSubmissionResult.QuizSession.UserSubmission.ID)
	})

	// ==================== USER ACTIVE SESSION TESTS ====================

	t.Run("userActiveSession returns nil when user has no access", func(t *testing.T) {
		challengeID := createChallenge(t, "UserActiveSession No Access Challenge")
		quizID := createQuiz(t, "UserActiveSession No Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session but don't grant access to user
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Query quiz.userActiveSession as user (no access)
		queryResp := client.WithAuth(userToken).MustExecute(t, `
			query GetQuiz($id: ID!) {
				quiz(id: $id) {
					id
					userActiveSession {
						id
						state
					}
				}
			}
		`, map[string]any{"id": quizID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			Quiz struct {
				ID                string `json:"id"`
				UserActiveSession *struct {
					ID    string `json:"id"`
					State string `json:"state"`
				} `json:"userActiveSession"`
			} `json:"quiz"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))
		assert.Nil(t, queryResult.Quiz.UserActiveSession)
	})

	t.Run("userActiveSession returns nil when session is not OPEN", func(t *testing.T) {
		challengeID := createChallenge(t, "UserActiveSession Not Open Challenge")
		quizID := createQuiz(t, "UserActiveSession Not Open Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session and grant access but keep in DRAFT
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access but don't open
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		// Query quiz.userActiveSession as user
		queryResp := client.WithAuth(userToken).MustExecute(t, `
			query GetQuiz($id: ID!) {
				quiz(id: $id) {
					id
					userActiveSession {
						id
						state
					}
				}
			}
		`, map[string]any{"id": quizID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			Quiz struct {
				ID                string `json:"id"`
				UserActiveSession *struct {
					ID    string `json:"id"`
					State string `json:"state"`
				} `json:"userActiveSession"`
			} `json:"quiz"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))
		assert.Nil(t, queryResult.Quiz.UserActiveSession)
	})

	t.Run("userActiveSession returns session when user has access to OPEN session", func(t *testing.T) {
		challengeID := createChallenge(t, "UserActiveSession Open Challenge")
		quizID := createQuiz(t, "UserActiveSession Open Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Query quiz.userActiveSession as user
		queryResp := client.WithAuth(userToken).MustExecute(t, `
			query GetQuiz($id: ID!) {
				quiz(id: $id) {
					id
					userActiveSession {
						id
						state
					}
				}
			}
		`, map[string]any{"id": quizID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			Quiz struct {
				ID                string `json:"id"`
				UserActiveSession *struct {
					ID    string `json:"id"`
					State string `json:"state"`
				} `json:"userActiveSession"`
			} `json:"quiz"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))
		assert.NotNil(t, queryResult.Quiz.UserActiveSession)
		assert.Equal(t, sessionID, queryResult.Quiz.UserActiveSession.ID)
		assert.Equal(t, "OPEN", queryResult.Quiz.UserActiveSession.State)
	})

	// ==================== CHALLENGE VISIBILITY TESTS ====================

	t.Run("non-admin user cannot see quiz challenge without session access", func(t *testing.T) {
		challengeID := createChallenge(t, "Visibility No Access Challenge")
		publishChallenge(t, challengeID)
		quizID := createQuiz(t, "Visibility No Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session but don't grant access
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Query event.challenges as non-admin user
		queryResp := client.WithAuth(userToken).MustExecute(t, `
			query GetEventChallenges($id: ID!) {
				event(id: $id) {
					challenges {
						id
						name
					}
				}
			}
		`, map[string]any{"id": eventID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			Event struct {
				Challenges []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"challenges"`
			} `json:"event"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))

		// Challenge should NOT be in the list
		for _, ch := range queryResult.Event.Challenges {
			assert.NotEqual(t, challengeID, ch.ID, "quiz challenge should not be visible without session access")
		}
	})

	t.Run("non-admin user can see quiz challenge with session access", func(t *testing.T) {
		challengeID := createChallenge(t, "Visibility With Access Challenge")
		publishChallenge(t, challengeID)
		quizID := createQuiz(t, "Visibility With Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Query event.challenges as non-admin user
		queryResp := client.WithAuth(userToken).MustExecute(t, `
			query GetEventChallenges($id: ID!) {
				event(id: $id) {
					challenges {
						id
						name
					}
				}
			}
		`, map[string]any{"id": eventID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			Event struct {
				Challenges []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"challenges"`
			} `json:"event"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))

		// Challenge SHOULD be in the list
		found := false
		for _, ch := range queryResult.Event.Challenges {
			if ch.ID == challengeID {
				found = true
				break
			}
		}
		assert.True(t, found, "quiz challenge should be visible with session access")
	})

	t.Run("admin user can see quiz challenge without session access", func(t *testing.T) {
		challengeID := createChallenge(t, "Visibility Admin Challenge")
		publishChallenge(t, challengeID)
		quizID := createQuiz(t, "Visibility Admin Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session but don't grant access to anyone
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Query event.challenges as admin user
		queryResp := client.WithAuth(adminToken).MustExecute(t, `
			query GetEventChallenges($id: ID!) {
				event(id: $id) {
					challenges {
						id
						name
					}
				}
			}
		`, map[string]any{"id": eventID})
		require.False(t, queryResp.HasErrors())

		var queryResult struct {
			Event struct {
				Challenges []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"challenges"`
			} `json:"event"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))

		// Challenge SHOULD be in the list for admin
		found := false
		for _, ch := range queryResult.Event.Challenges {
			if ch.ID == challengeID {
				found = true
				break
			}
		}
		assert.True(t, found, "admin should see quiz challenge without session access")
	})

	t.Run("project challenges filters quiz challenges by session access", func(t *testing.T) {
		challengeID := createChallenge(t, "Project Visibility Challenge")
		publishChallenge(t, challengeID)
		quizID := createQuiz(t, "Project Visibility Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session but don't grant access
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Query project.challenges as non-admin user - should NOT see challenge
		queryResp1 := client.WithAuth(userToken).MustExecute(t, `
			query GetProjectChallenges($id: ID!) {
				project(id: $id) {
					challenges {
						id
						name
					}
				}
			}
		`, map[string]any{"id": projectID})
		require.False(t, queryResp1.HasErrors())

		var queryResult1 struct {
			Project struct {
				Challenges []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"challenges"`
			} `json:"project"`
		}
		require.NoError(t, queryResp1.UnmarshalData(&queryResult1))

		found := false
		for _, ch := range queryResult1.Project.Challenges {
			if ch.ID == challengeID {
				found = true
				break
			}
		}
		assert.False(t, found, "quiz challenge should not be visible without session access")

		// Grant access
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		// Query again - should now see challenge
		queryResp2 := client.WithAuth(userToken).MustExecute(t, `
			query GetProjectChallenges($id: ID!) {
				project(id: $id) {
					challenges {
						id
						name
					}
				}
			}
		`, map[string]any{"id": projectID})
		require.False(t, queryResp2.HasErrors())

		var queryResult2 struct {
			Project struct {
				Challenges []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"challenges"`
			} `json:"project"`
		}
		require.NoError(t, queryResp2.UnmarshalData(&queryResult2))

		found = false
		for _, ch := range queryResult2.Project.Challenges {
			if ch.ID == challengeID {
				found = true
				break
			}
		}
		assert.True(t, found, "quiz challenge should be visible after granting session access")
	})

	// ==================== DIRECT CHALLENGE QUERY TESTS ====================

	t.Run("cannot load quiz challenge directly without session access", func(t *testing.T) {
		challengeID := createChallenge(t, "Direct Query No Access Challenge")
		publishChallenge(t, challengeID)
		quizID := createQuiz(t, "Direct Query No Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session but don't grant access
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Try to query challenge directly as non-admin user
		queryResp := client.WithAuth(userToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					id
					name
				}
			}
		`, map[string]any{"id": challengeID})

		require.True(t, queryResp.HasErrors())
		assert.Contains(t, queryResp.ErrorMessage(), "not found")
	})

	t.Run("can load quiz challenge directly with session access", func(t *testing.T) {
		challengeID := createChallenge(t, "Direct Query With Access Challenge")
		publishChallenge(t, challengeID)
		quizID := createQuiz(t, "Direct Query With Access Quiz", challengeID)
		addQuestion(t, quizID, "Question?")

		// Create session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"quizId": quizID,
			},
		})
		require.False(t, createResp.HasErrors())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// Grant access
		client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{userID},
			},
		})

		// Open session
		client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})

		// Query challenge directly as non-admin user
		queryResp := client.WithAuth(userToken).MustExecute(t, `
			query GetChallenge($id: ID!) {
				challenge(id: $id) {
					id
					name
				}
			}
		`, map[string]any{"id": challengeID})
		require.False(t, queryResp.HasErrors(), "failed to get challenge: %s", queryResp.ErrorMessage())

		var queryResult struct {
			Challenge struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"challenge"`
		}
		require.NoError(t, queryResp.UnmarshalData(&queryResult))
		assert.Equal(t, challengeID, queryResult.Challenge.ID)
	})
}
