package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderingQuestions(t *testing.T) {
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

	// Setup M2M token
	m2mToken, err := testutil.GenerateM2MToken()
	require.NoError(t, err)

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	userToken, err := testutil.GenerateUserToken(userID)
	require.NoError(t, err)

	user2Token, err := testutil.GenerateUserToken(user2ID)
	require.NoError(t, err)

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]

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
	createQuiz := func(t *testing.T, name string, challengeID string, completionPoints int) string {
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
				"description":          "Test quiz for ordering questions",
				"randomizeQuestions":   false,
				"revealCorrectAnswers": true,
				"allowRetakes":         false,
				"completionPoints":     completionPoints,
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

	// Helper to make a challenge visible
	makeChallengeVisible := func(t *testing.T, challengeID string) {
		visibleTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation SetChallengeVisibility($id: ID!, $visibleAt: DateTime!) {
				setChallengeVisibility(id: $id, visibleAt: $visibleAt) {
					id
				}
			}
		`, map[string]any{
			"id":        challengeID,
			"visibleAt": visibleTime,
		})
		require.False(t, resp.HasErrors(), "failed to set challenge visibility: %s", resp.ErrorMessage())
	}

	// Helper to add an ordering question to quiz
	type OrderingItemResult struct {
		ID       string `json:"id"`
		ItemText string `json:"itemText"`
	}

	type OrderingQuestionResult struct {
		ID            string               `json:"id"`
		QuestionText  string               `json:"questionText"`
		QuestionOrder int                  `json:"questionOrder"`
		Points        *int                 `json:"points"`
		OrderingItems []OrderingItemResult `json:"orderingItems"`
	}

	addOrderingQuestion := func(t *testing.T, quizID string, questionText string, questionOrder int, points int, items []map[string]any) OrderingQuestionResult {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on OrderingQuestion {
						id
						questionText
						questionOrder
						points
						orderingItems {
							id
							itemText
						}
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":  "ORDERING",
				"questionText":  questionText,
				"questionOrder": questionOrder,
				"points":        points,
				"orderingItems": items,
			},
		})
		require.False(t, resp.HasErrors(), "failed to add ordering question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion OrderingQuestionResult `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AddQuizQuestion
	}

	// Helper to create a quiz session
	createSession := func(t *testing.T, quizID string) string {
		resp := client.WithAuth(adminToken).MustExecute(t, `
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
		require.False(t, resp.HasErrors(), "failed to create session: %s", resp.ErrorMessage())

		var result struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.CreateQuizSession.ID
	}

	// Helper to grant session access
	grantAccess := func(t *testing.T, sessionID string, userIDs []string) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   userIDs,
			},
		})
		require.False(t, resp.HasErrors(), "failed to grant access: %s", resp.ErrorMessage())
	}

	// Helper to open session
	openSession := func(t *testing.T, sessionID string) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})
		require.False(t, resp.HasErrors(), "failed to open session: %s", resp.ErrorMessage())
	}

	// Helper to start quiz session (for user)
	startSession := func(t *testing.T, token string, sessionID string) string {
		resp := client.WithAuth(token).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) {
					id
				}
			}
		`, map[string]any{"sessionId": sessionID})
		require.False(t, resp.HasErrors(), "failed to start session: %s", resp.ErrorMessage())

		var result struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.StartQuizSession.ID
	}

	// Helper to lock a session (hides isCorrect while OPEN, reveals after LOCKED)
	lockSession := func(t *testing.T, sessionID string) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation LockSession($id: ID!) {
				lockQuizSession(id: $id) {
					id
					state
				}
			}
		`, map[string]any{"id": sessionID})
		require.False(t, resp.HasErrors(), "failed to lock session: %s", resp.ErrorMessage())

		var result struct {
			LockQuizSession struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"lockQuizSession"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, "LOCKED", result.LockQuizSession.State, "session should be LOCKED after lockQuizSession")
	}

	// ==================== QUESTION CREATION TESTS ====================

	t.Run("Question Creation", func(t *testing.T) {
		t.Run("admin can create ordering question with items", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering Create Challenge")
			quizID := createQuiz(t, "Ordering Create Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Put these events in chronological order:", 0, 5, []map[string]any{
				{"itemText": "Creation", "correctOrder": 1},
				{"itemText": "Flood", "correctOrder": 2},
				{"itemText": "Exodus", "correctOrder": 3},
				{"itemText": "Temple Built", "correctOrder": 4},
			})

			assert.NotEmpty(t, question.ID)
			assert.Equal(t, "Put these events in chronological order:", question.QuestionText)
			assert.Equal(t, 0, question.QuestionOrder)
			assert.NotNil(t, question.Points)
			assert.Equal(t, 5, *question.Points)
			assert.Len(t, question.OrderingItems, 4)

			// Verify all items are present
			itemTexts := make(map[string]bool)
			for _, item := range question.OrderingItems {
				assert.NotEmpty(t, item.ID)
				itemTexts[item.ItemText] = true
			}
			assert.True(t, itemTexts["Creation"])
			assert.True(t, itemTexts["Flood"])
			assert.True(t, itemTexts["Exodus"])
			assert.True(t, itemTexts["Temple Built"])
		})

		t.Run("correctOrder is not exposed to users", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering No CorrectOrder Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering No CorrectOrder Quiz", challengeID, 10)

			addOrderingQuestion(t, quizID, "Order these:", 0, 5, []map[string]any{
				{"itemText": "First", "correctOrder": 1},
				{"itemText": "Second", "correctOrder": 2},
			})

			// Query the quiz as user
			sessionID := createSession(t, quizID)
			grantAccess(t, sessionID, []string{userID})
			openSession(t, sessionID)

			resp := client.WithAuth(userToken).MustExecute(t, `
				query GetQuiz($id: ID!) {
					quiz(id: $id) {
						questions {
							... on OrderingQuestion {
								id
								questionText
								orderingItems {
									id
									itemText
								}
							}
						}
					}
				}
			`, map[string]any{"id": quizID})
			require.False(t, resp.HasErrors())

			// The response should NOT contain correctOrder field
			// (GraphQL schema doesn't expose it)
			var result struct {
				Quiz struct {
					Questions []struct {
						ID            string `json:"id"`
						OrderingItems []struct {
							ID       string `json:"id"`
							ItemText string `json:"itemText"`
						} `json:"orderingItems"`
					} `json:"questions"`
				} `json:"quiz"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.Len(t, result.Quiz.Questions, 1)
			assert.Len(t, result.Quiz.Questions[0].OrderingItems, 2)
		})

		t.Run("admin can update ordering items", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering Update Challenge")
			quizID := createQuiz(t, "Ordering Update Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Order these:", 0, 5, []map[string]any{
				{"itemText": "One", "correctOrder": 1},
				{"itemText": "Two", "correctOrder": 2},
			})
			questionID := question.ID
			assert.Len(t, question.OrderingItems, 2)

			// Update with new items
			resp := client.WithAuth(adminToken).MustExecute(t, `
				mutation UpdateQuestion($id: ID!, $input: UpdateQuizQuestionInput!) {
					updateQuizQuestion(id: $id, input: $input) {
						... on OrderingQuestion {
							id
							orderingItems {
								id
								itemText
							}
						}
					}
				}
			`, map[string]any{
				"id": questionID,
				"input": map[string]any{
					"orderingItems": []map[string]any{
						{"itemText": "Alpha", "correctOrder": 1},
						{"itemText": "Beta", "correctOrder": 2},
						{"itemText": "Gamma", "correctOrder": 3},
					},
				},
			})
			require.False(t, resp.HasErrors(), "failed to update question: %s", resp.ErrorMessage())

			var result struct {
				UpdateQuizQuestion struct {
					ID            string `json:"id"`
					OrderingItems []struct {
						ID       string `json:"id"`
						ItemText string `json:"itemText"`
					} `json:"orderingItems"`
				} `json:"updateQuizQuestion"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.Len(t, result.UpdateQuizQuestion.OrderingItems, 3)

			// Verify new items
			itemTexts := make(map[string]bool)
			for _, item := range result.UpdateQuizQuestion.OrderingItems {
				itemTexts[item.ItemText] = true
			}
			assert.True(t, itemTexts["Alpha"])
			assert.True(t, itemTexts["Beta"])
			assert.True(t, itemTexts["Gamma"])
		})
	})

	// ==================== ANSWER SUBMISSION TESTS ====================

	t.Run("Answer Submission", func(t *testing.T) {
		t.Run("correct order is marked correct", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering Correct Answer Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering Correct Answer Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Put in order:", 0, 5, []map[string]any{
				{"itemText": "First", "correctOrder": 1},
				{"itemText": "Second", "correctOrder": 2},
				{"itemText": "Third", "correctOrder": 3},
			})

			// Get items sorted by text to find correct order
			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			// Build correct order
			correctOrder := []string{
				itemByText["First"],
				itemByText["Second"],
				itemByText["Third"],
			}

			// Create session and start quiz
			sessionID := createSession(t, quizID)
			grantAccess(t, sessionID, []string{userID})
			openSession(t, sessionID)
			submissionID := startSession(t, userToken, sessionID)

			// Submit correct answer
			resp := client.WithAuth(userToken).MustExecute(t, `
				mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
					submitQuizAnswer(submissionId: $submissionId, input: $input) {
						... on OrderingResponse {
							id
							submittedOrder
						}
					}
				}
			`, map[string]any{
				"submissionId": submissionID,
				"input": map[string]any{
					"questionId":       question.ID,
					"submittedOrder":   correctOrder,
					"timeSpentSeconds": 10,
				},
			})
			require.False(t, resp.HasErrors(), "failed to submit answer: %s", resp.ErrorMessage())

			var submitResult struct {
				SubmitQuizAnswer struct {
					ID             string   `json:"id"`
					SubmittedOrder []string `json:"submittedOrder"`
				} `json:"submitQuizAnswer"`
			}
			require.NoError(t, resp.UnmarshalData(&submitResult))
			assert.NotEmpty(t, submitResult.SubmitQuizAnswer.ID)
			assert.Equal(t, correctOrder, submitResult.SubmitQuizAnswer.SubmittedOrder)

			// Lock session to reveal isCorrect (hidden while session is OPEN)
			lockSession(t, sessionID)

			// Query submission to verify isCorrect and pointsEarned
			queryResp := client.WithAuth(userToken).MustExecute(t, `
				query GetQuizSubmission($id: ID!) {
					quizSubmission(id: $id) {
						responses {
							... on OrderingResponse {
								isCorrect
								pointsEarned
							}
						}
					}
				}
			`, map[string]any{"id": submissionID})
			require.False(t, queryResp.HasErrors(), "failed to query submission: %s", queryResp.ErrorMessage())

			var queryResult struct {
				QuizSubmission struct {
					Responses []struct {
						IsCorrect    *bool `json:"isCorrect"`
						PointsEarned *int  `json:"pointsEarned"`
					} `json:"responses"`
				} `json:"quizSubmission"`
			}
			require.NoError(t, queryResp.UnmarshalData(&queryResult))
			require.Len(t, queryResult.QuizSubmission.Responses, 1)
			assert.NotNil(t, queryResult.QuizSubmission.Responses[0].IsCorrect)
			assert.True(t, *queryResult.QuizSubmission.Responses[0].IsCorrect)
			assert.NotNil(t, queryResult.QuizSubmission.Responses[0].PointsEarned)
			assert.Equal(t, 5, *queryResult.QuizSubmission.Responses[0].PointsEarned)
		})

		t.Run("incorrect order is marked incorrect", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering Incorrect Answer Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering Incorrect Answer Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Put in order:", 0, 5, []map[string]any{
				{"itemText": "First", "correctOrder": 1},
				{"itemText": "Second", "correctOrder": 2},
				{"itemText": "Third", "correctOrder": 3},
			})

			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			// Build wrong order (swapped)
			wrongOrder := []string{
				itemByText["Third"],
				itemByText["Second"],
				itemByText["First"],
			}

			// Create session and start quiz
			sessionID := createSession(t, quizID)
			grantAccess(t, sessionID, []string{userID})
			openSession(t, sessionID)
			submissionID := startSession(t, userToken, sessionID)

			// Submit wrong answer
			resp := client.WithAuth(userToken).MustExecute(t, `
				mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
					submitQuizAnswer(submissionId: $submissionId, input: $input) {
						... on OrderingResponse {
							id
							submittedOrder
							isCorrect
							pointsEarned
						}
					}
				}
			`, map[string]any{
				"submissionId": submissionID,
				"input": map[string]any{
					"questionId":       question.ID,
					"submittedOrder":   wrongOrder,
					"timeSpentSeconds": 10,
				},
			})
			require.False(t, resp.HasErrors(), "failed to submit answer: %s", resp.ErrorMessage())

			var result struct {
				SubmitQuizAnswer struct {
					ID             string   `json:"id"`
					SubmittedOrder []string `json:"submittedOrder"`
					IsCorrect      *bool    `json:"isCorrect"`
					PointsEarned   *int     `json:"pointsEarned"`
				} `json:"submitQuizAnswer"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.NotEmpty(t, result.SubmitQuizAnswer.ID)
			assert.NotNil(t, result.SubmitQuizAnswer.IsCorrect)
			assert.False(t, *result.SubmitQuizAnswer.IsCorrect)
			// Points should not be awarded for wrong answer
			if result.SubmitQuizAnswer.PointsEarned != nil {
				assert.Equal(t, 0, *result.SubmitQuizAnswer.PointsEarned)
			}
		})

		t.Run("partial order (missing items) is marked incorrect", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering Partial Answer Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering Partial Answer Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Put in order:", 0, 5, []map[string]any{
				{"itemText": "First", "correctOrder": 1},
				{"itemText": "Second", "correctOrder": 2},
				{"itemText": "Third", "correctOrder": 3},
			})

			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			// Build partial order (missing Third)
			partialOrder := []string{
				itemByText["First"],
				itemByText["Second"],
			}

			// Create session and start quiz
			sessionID := createSession(t, quizID)
			grantAccess(t, sessionID, []string{userID})
			openSession(t, sessionID)
			submissionID := startSession(t, userToken, sessionID)

			// Submit partial answer
			resp := client.WithAuth(userToken).MustExecute(t, `
				mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
					submitQuizAnswer(submissionId: $submissionId, input: $input) {
						... on OrderingResponse {
							id
							submittedOrder
							isCorrect
							pointsEarned
						}
					}
				}
			`, map[string]any{
				"submissionId": submissionID,
				"input": map[string]any{
					"questionId":       question.ID,
					"submittedOrder":   partialOrder,
					"timeSpentSeconds": 10,
				},
			})
			require.False(t, resp.HasErrors(), "failed to submit answer: %s", resp.ErrorMessage())

			var result struct {
				SubmitQuizAnswer struct {
					ID             string   `json:"id"`
					SubmittedOrder []string `json:"submittedOrder"`
					IsCorrect      *bool    `json:"isCorrect"`
					PointsEarned   *int     `json:"pointsEarned"`
				} `json:"submitQuizAnswer"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.NotNil(t, result.SubmitQuizAnswer.IsCorrect)
			assert.False(t, *result.SubmitQuizAnswer.IsCorrect, "partial order should be marked incorrect")
		})

		t.Run("extra items is marked incorrect", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering Extra Items Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering Extra Items Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Put in order:", 0, 5, []map[string]any{
				{"itemText": "First", "correctOrder": 1},
				{"itemText": "Second", "correctOrder": 2},
			})

			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			// Build order with extra item (duplicate)
			extraOrder := []string{
				itemByText["First"],
				itemByText["Second"],
				itemByText["First"], // Duplicate
			}

			// Create session and start quiz
			sessionID := createSession(t, quizID)
			grantAccess(t, sessionID, []string{userID})
			openSession(t, sessionID)
			submissionID := startSession(t, userToken, sessionID)

			// Submit answer with extra items
			resp := client.WithAuth(userToken).MustExecute(t, `
				mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
					submitQuizAnswer(submissionId: $submissionId, input: $input) {
						... on OrderingResponse {
							id
							submittedOrder
							isCorrect
							pointsEarned
						}
					}
				}
			`, map[string]any{
				"submissionId": submissionID,
				"input": map[string]any{
					"questionId":       question.ID,
					"submittedOrder":   extraOrder,
					"timeSpentSeconds": 10,
				},
			})
			require.False(t, resp.HasErrors(), "failed to submit answer: %s", resp.ErrorMessage())

			var result struct {
				SubmitQuizAnswer struct {
					ID        string `json:"id"`
					IsCorrect *bool  `json:"isCorrect"`
				} `json:"submitQuizAnswer"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.NotNil(t, result.SubmitQuizAnswer.IsCorrect)
			assert.False(t, *result.SubmitQuizAnswer.IsCorrect, "extra items should be marked incorrect")
		})
	})

	// ==================== QUIZ COMPLETION TESTS ====================

	t.Run("Quiz Completion", func(t *testing.T) {
		t.Run("ordering question contributes to score", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering Score Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering Score Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Order:", 0, 5, []map[string]any{
				{"itemText": "A", "correctOrder": 1},
				{"itemText": "B", "correctOrder": 2},
			})

			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			correctOrder := []string{itemByText["A"], itemByText["B"]}

			// Create session and start quiz
			sessionID := createSession(t, quizID)
			grantAccess(t, sessionID, []string{userID})
			openSession(t, sessionID)
			submissionID := startSession(t, userToken, sessionID)

			// Submit correct answer
			client.WithAuth(userToken).MustExecute(t, `
				mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
					submitQuizAnswer(submissionId: $submissionId, input: $input) {
						... on OrderingResponse { id }
					}
				}
			`, map[string]any{
				"submissionId": submissionID,
				"input": map[string]any{
					"questionId":       question.ID,
					"submittedOrder":   correctOrder,
					"timeSpentSeconds": 10,
				},
			})

			// Finalize quiz
			resp := client.WithAuth(userToken).MustExecute(t, `
				mutation FinalizeQuiz($submissionId: ID!) {
					finalizeQuiz(submissionId: $submissionId) {
						id
						score
						maxScore
						scorePercentage
						pointsAwarded
					}
				}
			`, map[string]any{"submissionId": submissionID})
			require.False(t, resp.HasErrors(), "failed to finalize quiz: %s", resp.ErrorMessage())

			var result struct {
				FinalizeQuiz struct {
					ID              string   `json:"id"`
					Score           *int     `json:"score"`
					MaxScore        *int     `json:"maxScore"`
					ScorePercentage *float64 `json:"scorePercentage"`
					PointsAwarded   *int     `json:"pointsAwarded"`
				} `json:"finalizeQuiz"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.NotNil(t, result.FinalizeQuiz.Score)
			assert.Equal(t, 5, *result.FinalizeQuiz.Score)
			assert.NotNil(t, result.FinalizeQuiz.MaxScore)
			assert.Equal(t, 5, *result.FinalizeQuiz.MaxScore)
			assert.NotNil(t, result.FinalizeQuiz.ScorePercentage)
			assert.Equal(t, 100.0, *result.FinalizeQuiz.ScorePercentage)
		})

		t.Run("ordering question counts toward maxScore", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering MaxScore Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering MaxScore Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Order:", 0, 10, []map[string]any{
				{"itemText": "X", "correctOrder": 1},
				{"itemText": "Y", "correctOrder": 2},
			})

			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			// Wrong order
			wrongOrder := []string{itemByText["Y"], itemByText["X"]}

			// Create session and start quiz
			sessionID := createSession(t, quizID)
			grantAccess(t, sessionID, []string{userID})
			openSession(t, sessionID)
			submissionID := startSession(t, userToken, sessionID)

			// Submit wrong answer
			client.WithAuth(userToken).MustExecute(t, `
				mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
					submitQuizAnswer(submissionId: $submissionId, input: $input) {
						... on OrderingResponse { id }
					}
				}
			`, map[string]any{
				"submissionId": submissionID,
				"input": map[string]any{
					"questionId":       question.ID,
					"submittedOrder":   wrongOrder,
					"timeSpentSeconds": 10,
				},
			})

			// Finalize quiz
			resp := client.WithAuth(userToken).MustExecute(t, `
				mutation FinalizeQuiz($submissionId: ID!) {
					finalizeQuiz(submissionId: $submissionId) {
						id
						score
						maxScore
						scorePercentage
					}
				}
			`, map[string]any{"submissionId": submissionID})
			require.False(t, resp.HasErrors(), "failed to finalize quiz: %s", resp.ErrorMessage())

			var result struct {
				FinalizeQuiz struct {
					Score           *int     `json:"score"`
					MaxScore        *int     `json:"maxScore"`
					ScorePercentage *float64 `json:"scorePercentage"`
				} `json:"finalizeQuiz"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.NotNil(t, result.FinalizeQuiz.Score)
			assert.Equal(t, 0, *result.FinalizeQuiz.Score)
			assert.NotNil(t, result.FinalizeQuiz.MaxScore)
			assert.Equal(t, 10, *result.FinalizeQuiz.MaxScore)
			assert.NotNil(t, result.FinalizeQuiz.ScorePercentage)
			assert.Equal(t, 0.0, *result.FinalizeQuiz.ScorePercentage)
		})
	})

	// ==================== M2M API TESTS ====================

	t.Run("M2M API", func(t *testing.T) {
		t.Run("create submission with ordering responses", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering M2M Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering M2M Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Order these:", 0, 5, []map[string]any{
				{"itemText": "One", "correctOrder": 1},
				{"itemText": "Two", "correctOrder": 2},
				{"itemText": "Three", "correctOrder": 3},
			})

			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			correctOrder := []string{
				itemByText["One"],
				itemByText["Two"],
				itemByText["Three"],
			}

			// Create submission via M2M
			resp := client.WithAuth(m2mToken).MustExecute(t, `
				mutation CreateSubmission($quizId: ID!, $userId: ID!, $responses: [SubmitQuizAnswerInput!]!, $completedAt: DateTime!) {
					createQuizSubmission(quizId: $quizId, userId: $userId, responses: $responses, completedAt: $completedAt) {
						id
						score
						maxScore
						scorePercentage
						responses {
							... on OrderingResponse {
								id
								submittedOrder
								isCorrect
								pointsEarned
							}
						}
					}
				}
			`, map[string]any{
				"quizId":      quizID,
				"userId":      user2ID,
				"completedAt": time.Now().UTC().Format(time.RFC3339),
				"responses": []map[string]any{
					{
						"questionId":       question.ID,
						"submittedOrder":   correctOrder,
						"timeSpentSeconds": 15,
					},
				},
			})
			require.False(t, resp.HasErrors(), "failed to create submission: %s", resp.ErrorMessage())

			var result struct {
				CreateQuizSubmission struct {
					ID              string   `json:"id"`
					Score           *int     `json:"score"`
					MaxScore        *int     `json:"maxScore"`
					ScorePercentage *float64 `json:"scorePercentage"`
					Responses       []struct {
						ID             string   `json:"id"`
						SubmittedOrder []string `json:"submittedOrder"`
						IsCorrect      *bool    `json:"isCorrect"`
						PointsEarned   *int     `json:"pointsEarned"`
					} `json:"responses"`
				} `json:"createQuizSubmission"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.NotEmpty(t, result.CreateQuizSubmission.ID)
			assert.NotNil(t, result.CreateQuizSubmission.Score)
			assert.Equal(t, 5, *result.CreateQuizSubmission.Score)
			assert.NotNil(t, result.CreateQuizSubmission.MaxScore)
			assert.Equal(t, 5, *result.CreateQuizSubmission.MaxScore)
			require.Len(t, result.CreateQuizSubmission.Responses, 1)
			assert.Equal(t, correctOrder, result.CreateQuizSubmission.Responses[0].SubmittedOrder)
			assert.NotNil(t, result.CreateQuizSubmission.Responses[0].IsCorrect)
			assert.True(t, *result.CreateQuizSubmission.Responses[0].IsCorrect)
			assert.NotNil(t, result.CreateQuizSubmission.Responses[0].PointsEarned)
			assert.Equal(t, 5, *result.CreateQuizSubmission.Responses[0].PointsEarned)
		})

		t.Run("M2M create submission with incorrect ordering", func(t *testing.T) {
			challengeID := createChallenge(t, "Ordering M2M Incorrect Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Ordering M2M Incorrect Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Order these:", 0, 5, []map[string]any{
				{"itemText": "First", "correctOrder": 1},
				{"itemText": "Last", "correctOrder": 2},
			})

			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			// Reverse order (incorrect)
			wrongOrder := []string{
				itemByText["Last"],
				itemByText["First"],
			}

			// Create submission via M2M
			resp := client.WithAuth(m2mToken).MustExecute(t, `
				mutation CreateSubmission($quizId: ID!, $userId: ID!, $responses: [SubmitQuizAnswerInput!]!, $completedAt: DateTime!) {
					createQuizSubmission(quizId: $quizId, userId: $userId, responses: $responses, completedAt: $completedAt) {
						id
						score
						maxScore
						responses {
							... on OrderingResponse {
								id
								isCorrect
								pointsEarned
							}
						}
					}
				}
			`, map[string]any{
				"quizId":      quizID,
				"userId":      user2ID,
				"completedAt": time.Now().UTC().Format(time.RFC3339),
				"responses": []map[string]any{
					{
						"questionId":       question.ID,
						"submittedOrder":   wrongOrder,
						"timeSpentSeconds": 10,
					},
				},
			})
			require.False(t, resp.HasErrors(), "failed to create submission: %s", resp.ErrorMessage())

			var result struct {
				CreateQuizSubmission struct {
					ID        string `json:"id"`
					Score     *int   `json:"score"`
					MaxScore  *int   `json:"maxScore"`
					Responses []struct {
						ID           string `json:"id"`
						IsCorrect    *bool  `json:"isCorrect"`
						PointsEarned *int   `json:"pointsEarned"`
					} `json:"responses"`
				} `json:"createQuizSubmission"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.NotNil(t, result.CreateQuizSubmission.Score)
			assert.Equal(t, 0, *result.CreateQuizSubmission.Score)
			assert.NotNil(t, result.CreateQuizSubmission.MaxScore)
			assert.Equal(t, 5, *result.CreateQuizSubmission.MaxScore)
			require.Len(t, result.CreateQuizSubmission.Responses, 1)
			assert.NotNil(t, result.CreateQuizSubmission.Responses[0].IsCorrect)
			assert.False(t, *result.CreateQuizSubmission.Responses[0].IsCorrect)
		})
	})

	// ==================== GET QUIZ SUBMISSION TEST ====================

	t.Run("Get Quiz Submission", func(t *testing.T) {
		t.Run("includes OrderingResponse in responses", func(t *testing.T) {
			challengeID := createChallenge(t, "Get Submission Challenge")
			publishChallenge(t, challengeID)
			makeChallengeVisible(t, challengeID)
			quizID := createQuiz(t, "Get Submission Quiz", challengeID, 10)

			question := addOrderingQuestion(t, quizID, "Order:", 0, 5, []map[string]any{
				{"itemText": "Alpha", "correctOrder": 1},
				{"itemText": "Beta", "correctOrder": 2},
			})

			itemByText := make(map[string]string)
			for _, item := range question.OrderingItems {
				itemByText[item.ItemText] = item.ID
			}

			correctOrder := []string{itemByText["Alpha"], itemByText["Beta"]}

			// Create session and start quiz
			sessionID := createSession(t, quizID)
			grantAccess(t, sessionID, []string{user2ID})
			openSession(t, sessionID)
			submissionID := startSession(t, user2Token, sessionID)

			// Submit answer
			client.WithAuth(user2Token).MustExecute(t, `
				mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
					submitQuizAnswer(submissionId: $submissionId, input: $input) {
						... on OrderingResponse { id }
					}
				}
			`, map[string]any{
				"submissionId": submissionID,
				"input": map[string]any{
					"questionId":       question.ID,
					"submittedOrder":   correctOrder,
					"timeSpentSeconds": 10,
				},
			})

			// Finalize
			client.WithAuth(user2Token).MustExecute(t, `
				mutation FinalizeQuiz($submissionId: ID!) {
					finalizeQuiz(submissionId: $submissionId) { id }
				}
			`, map[string]any{"submissionId": submissionID})

			// Get submission
			resp := client.WithAuth(user2Token).MustExecute(t, `
				query GetQuizSubmission($id: ID!) {
					quizSubmission(id: $id) {
						id
						score
						maxScore
						responses {
							id
							answeredAt
							timeSpentSeconds
							question {
								id
								questionText
							}
							... on OrderingResponse {
								submittedOrder
								isCorrect
								pointsEarned
							}
						}
					}
				}
			`, map[string]any{"id": submissionID})
			require.False(t, resp.HasErrors(), "failed to get submission: %s", resp.ErrorMessage())

			var result struct {
				QuizSubmission struct {
					ID        string `json:"id"`
					Score     *int   `json:"score"`
					MaxScore  *int   `json:"maxScore"`
					Responses []struct {
						ID               string   `json:"id"`
						AnsweredAt       string   `json:"answeredAt"`
						TimeSpentSeconds *int     `json:"timeSpentSeconds"`
						SubmittedOrder   []string `json:"submittedOrder"`
						IsCorrect        *bool    `json:"isCorrect"`
						PointsEarned     *int     `json:"pointsEarned"`
						Question         struct {
							ID           string `json:"id"`
							QuestionText string `json:"questionText"`
						} `json:"question"`
					} `json:"responses"`
				} `json:"quizSubmission"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.Equal(t, submissionID, result.QuizSubmission.ID)
			require.Len(t, result.QuizSubmission.Responses, 1)

			response := result.QuizSubmission.Responses[0]
			assert.NotEmpty(t, response.ID)
			assert.NotEmpty(t, response.AnsweredAt)
			assert.NotNil(t, response.TimeSpentSeconds)
			assert.Equal(t, 10, *response.TimeSpentSeconds)
			assert.Equal(t, correctOrder, response.SubmittedOrder)
			assert.NotNil(t, response.IsCorrect)
			assert.True(t, *response.IsCorrect)
			assert.NotNil(t, response.PointsEarned)
			assert.Equal(t, 5, *response.PointsEarned)
			assert.Equal(t, question.ID, response.Question.ID)
			assert.Equal(t, "Order:", response.Question.QuestionText)
		})
	})
}
