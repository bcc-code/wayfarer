package e2e

import (
	"context"
	"maps"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuizBetting(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean and seed with deterministic data
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	// Setup user IDs
	userID := data.UserIDs[0]
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
				"description":          "Test quiz for betting",
				"randomizeQuestions":   false,
				"revealCorrectAnswers": true,
				"allowRetakes":         true,
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

	// Helper to add a question with betting configuration
	addQuestionWithBetting := func(t *testing.T, quizID string, questionText string, bettingConfig map[string]any) string {
		input := map[string]any{
			"questionType":  "PREDEFINED",
			"questionText":  questionText,
			"questionOrder": 0,
			"points":        10,
			"predefinedAnswers": []map[string]any{
				{"answerText": "Correct Answer", "isCorrect": true, "answerOrder": 0},
				{"answerText": "Wrong Answer", "isCorrect": false, "answerOrder": 1},
			},
		}

		// Merge betting config
		maps.Copy(input, bettingConfig)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on PredefinedQuestion {
						id
						bettingEnabled
						bettingMinPercentage
						bettingMaxPercentage
						bettingMinAbsolute
						bettingMaxAbsolute
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input":  input,
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion struct {
				ID                   string   `json:"id"`
				BettingEnabled       bool     `json:"bettingEnabled"`
				BettingMinPercentage *float64 `json:"bettingMinPercentage"`
				BettingMaxPercentage *float64 `json:"bettingMaxPercentage"`
				BettingMinAbsolute   *int     `json:"bettingMinAbsolute"`
				BettingMaxAbsolute   *int     `json:"bettingMaxAbsolute"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AddQuizQuestion.ID
	}

	// Helper to give user points via score adjustment
	giveUserPoints := func(t *testing.T, targetUserID string, points int) {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateScoreAdjustment($input: CreateScoreAdjustmentInput!) {
				createScoreAdjustment(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId": projectID,
				"userId":    targetUserID,
				"points":    points,
				"reason":    "Test points for betting",
			},
		})
		require.False(t, resp.HasErrors(), "failed to create score adjustment: %s", resp.ErrorMessage())
	}

	// Helper to start quiz and get submission ID (uses quiz session flow)
	startQuiz := func(t *testing.T, quizID string, targetUserID string, token string) string {
		// 1. Create quiz session
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) { id }
			}
		`, map[string]any{
			"input": map[string]any{"quizId": quizID},
		})
		require.False(t, createResp.HasErrors(), "failed to create quiz session: %s", createResp.ErrorMessage())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		// 2. Grant access to user
		grantResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"sessionId": sessionID,
				"userIds":   []string{targetUserID},
			},
		})
		require.False(t, grantResp.HasErrors(), "failed to grant session access: %s", grantResp.ErrorMessage())

		// 3. Open session
		openResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) {
				openQuizSession(id: $id) { id }
			}
		`, map[string]any{"id": sessionID})
		require.False(t, openResp.HasErrors(), "failed to open session: %s", openResp.ErrorMessage())

		// 4. User starts session
		startResp := client.WithAuth(token).MustExecute(t, `
			mutation StartSession($sessionId: ID!) {
				startQuizSession(sessionId: $sessionId) { id }
			}
		`, map[string]any{"sessionId": sessionID})
		require.False(t, startResp.HasErrors(), "failed to start quiz: %s", startResp.ErrorMessage())

		var startResult struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		return startResult.StartQuizSession.ID
	}

	// ==================== BETTING CONFIGURATION TESTS ====================

	t.Run("admin can create question with betting enabled", func(t *testing.T) {
		challengeID := createChallenge(t, "Betting Config Challenge 1")
		quizID := createQuiz(t, "Betting Config Quiz 1", challengeID)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on PredefinedQuestion {
						id
						bettingEnabled
						bettingMinPercentage
						bettingMaxPercentage
						bettingMinAbsolute
						bettingMaxAbsolute
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":         "PREDEFINED",
				"questionText":         "Question with betting?",
				"questionOrder":        0,
				"points":               10,
				"bettingEnabled":       true,
				"bettingMinPercentage": 10.0,
				"bettingMaxPercentage": 50.0,
				"bettingMinAbsolute":   5,
				"bettingMaxAbsolute":   100,
				"predefinedAnswers": []map[string]any{
					{"answerText": "Yes", "isCorrect": true, "answerOrder": 0},
					{"answerText": "No", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion struct {
				ID                   string   `json:"id"`
				BettingEnabled       bool     `json:"bettingEnabled"`
				BettingMinPercentage *float64 `json:"bettingMinPercentage"`
				BettingMaxPercentage *float64 `json:"bettingMaxPercentage"`
				BettingMinAbsolute   *int     `json:"bettingMinAbsolute"`
				BettingMaxAbsolute   *int     `json:"bettingMaxAbsolute"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.True(t, result.AddQuizQuestion.BettingEnabled)
		assert.NotNil(t, result.AddQuizQuestion.BettingMinPercentage)
		assert.Equal(t, 10.0, *result.AddQuizQuestion.BettingMinPercentage)
		assert.NotNil(t, result.AddQuizQuestion.BettingMaxPercentage)
		assert.Equal(t, 50.0, *result.AddQuizQuestion.BettingMaxPercentage)
		assert.NotNil(t, result.AddQuizQuestion.BettingMinAbsolute)
		assert.Equal(t, 5, *result.AddQuizQuestion.BettingMinAbsolute)
		assert.NotNil(t, result.AddQuizQuestion.BettingMaxAbsolute)
		assert.Equal(t, 100, *result.AddQuizQuestion.BettingMaxAbsolute)
	})

	t.Run("question betting disabled by default", func(t *testing.T) {
		challengeID := createChallenge(t, "Betting Default Challenge")
		quizID := createQuiz(t, "Betting Default Quiz", challengeID)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on PredefinedQuestion {
						id
						bettingEnabled
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":  "PREDEFINED",
				"questionText":  "Question without betting config?",
				"questionOrder": 0,
				"points":        10,
				"predefinedAnswers": []map[string]any{
					{"answerText": "Yes", "isCorrect": true, "answerOrder": 0},
					{"answerText": "No", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors())

		var result struct {
			AddQuizQuestion struct {
				BettingEnabled bool `json:"bettingEnabled"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.False(t, result.AddQuizQuestion.BettingEnabled)
	})

	t.Run("admin can update question betting configuration", func(t *testing.T) {
		challengeID := createChallenge(t, "Betting Update Challenge")
		quizID := createQuiz(t, "Betting Update Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Update betting?", map[string]any{
			"bettingEnabled": false,
		})

		// Update to enable betting
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateQuestion($id: ID!, $input: UpdateQuizQuestionInput!) {
				updateQuizQuestion(id: $id, input: $input) {
					... on PredefinedQuestion {
						id
						bettingEnabled
						bettingMinPercentage
						bettingMaxPercentage
					}
				}
			}
		`, map[string]any{
			"id": questionID,
			"input": map[string]any{
				"bettingEnabled":       true,
				"bettingMinPercentage": 5.0,
				"bettingMaxPercentage": 75.0,
			},
		})
		require.False(t, resp.HasErrors(), "failed to update question: %s", resp.ErrorMessage())

		var result struct {
			UpdateQuizQuestion struct {
				ID                   string   `json:"id"`
				BettingEnabled       bool     `json:"bettingEnabled"`
				BettingMinPercentage *float64 `json:"bettingMinPercentage"`
				BettingMaxPercentage *float64 `json:"bettingMaxPercentage"`
			} `json:"updateQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.True(t, result.UpdateQuizQuestion.BettingEnabled)
		assert.Equal(t, 5.0, *result.UpdateQuizQuestion.BettingMinPercentage)
		assert.Equal(t, 75.0, *result.UpdateQuizQuestion.BettingMaxPercentage)
	})

	// ==================== BET SUBMISSION TESTS ====================

	t.Run("user can submit answer with bet", func(t *testing.T) {
		challengeID := createChallenge(t, "Bet Submit Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "Bet Submit Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Submit bet?", map[string]any{
			"bettingEnabled": true,
		})

		// Give user some points to bet
		giveUserPoints(t, userID, 100)

		// Start quiz
		submissionID := startQuiz(t, quizID, userID, userToken)

		// Submit answer with bet
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
					betAmount
					... on PredefinedResponse {
						selectedAnswerIds
					}
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
				"betAmount":         25,
			},
		})
		require.False(t, resp.HasErrors(), "failed to submit answer: %s", resp.ErrorMessage())

		var result struct {
			SubmitQuizAnswer struct {
				ID        string `json:"id"`
				BetAmount *int   `json:"betAmount"`
			} `json:"submitQuizAnswer"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotNil(t, result.SubmitQuizAnswer.BetAmount)
		assert.Equal(t, 25, *result.SubmitQuizAnswer.BetAmount)
	})

	t.Run("bet required when betting enabled", func(t *testing.T) {
		challengeID := createChallenge(t, "No Bet Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "No Bet Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "No bet?", map[string]any{
			"bettingEnabled": true,
		})

		submissionID := startQuiz(t, quizID, userID, userToken)

		// Submit answer without bet - should fail when betting is enabled
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
					betAmount
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
			},
		})
		require.True(t, resp.HasErrors(), "expected error when betting is enabled but no bet provided")
		assert.Contains(t, resp.ErrorMessage(), "bet is required when betting is enabled")
	})

	t.Run("bet rejected when betting is disabled", func(t *testing.T) {
		challengeID := createChallenge(t, "Betting Disabled Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "Betting Disabled Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Betting disabled?", map[string]any{
			"bettingEnabled": false,
		})

		submissionID := startQuiz(t, quizID, userID, userToken)

		// Try to submit answer with bet
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
				"betAmount":         10,
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "betting is not enabled")
	})

	t.Run("bet rejected when exceeds current score", func(t *testing.T) {
		challengeID := createChallenge(t, "Bet Exceeds Score Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "Bet Exceeds Score Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Bet too high?", map[string]any{
			"bettingEnabled": true,
		})

		// User has 100 points from earlier test, try to bet more
		submissionID := startQuiz(t, quizID, userID, userToken)

		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
				"betAmount":         99999,
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "exceeds current score")
	})

	t.Run("bet rejected when below minimum absolute", func(t *testing.T) {
		challengeID := createChallenge(t, "Min Absolute Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "Min Absolute Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Min absolute test?", map[string]any{
			"bettingEnabled":     true,
			"bettingMinAbsolute": 20,
		})

		submissionID := startQuiz(t, quizID, userID, userToken)

		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
				"betAmount":         5, // Below minimum of 20
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "below minimum")
	})

	t.Run("bet rejected when exceeds maximum absolute", func(t *testing.T) {
		challengeID := createChallenge(t, "Max Absolute Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "Max Absolute Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Max absolute test?", map[string]any{
			"bettingEnabled":     true,
			"bettingMaxAbsolute": 30,
		})

		submissionID := startQuiz(t, quizID, userID, userToken)

		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
				"betAmount":         50, // Above maximum of 30
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "exceeds maximum")
	})

	t.Run("bet rejected when below minimum percentage", func(t *testing.T) {
		challengeID := createChallenge(t, "Min Percentage Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "Min Percentage Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Min percentage test?", map[string]any{
			"bettingEnabled":       true,
			"bettingMinPercentage": 20.0, // 20% of score
		})

		// Give user 100 points, minimum bet would be 20
		giveUserPoints(t, userID, 100)

		submissionID := startQuiz(t, quizID, userID, userToken)

		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
				"betAmount":         10, // 10 is less than 20% of score
			},
		})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "below minimum percentage")
	})

	t.Run("bet rejected when exceeds maximum percentage", func(t *testing.T) {
		challengeID := createChallenge(t, "Max Percentage Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "Max Percentage Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Max percentage test?", map[string]any{
			"bettingEnabled":       true,
			"bettingMaxPercentage": 1.0, // 1% of score - very restrictive
		})

		submissionID := startQuiz(t, quizID, userID, userToken)

		// Bet 10000 points - this should exceed 1% of any reasonable score
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
				"betAmount":         10000, // Should exceed 1% of score (and also exceeds total score)
			},
		})

		require.True(t, resp.HasErrors())
		// Could fail for either exceeds current score or exceeds max percentage
		assert.True(t, resp.HasErrors())
	})

	t.Run("zero bet rejected when betting enabled", func(t *testing.T) {
		challengeID := createChallenge(t, "Zero Bet Challenge")
		publishChallenge(t, challengeID)
		makeChallengeVisible(t, challengeID)
		quizID := createQuiz(t, "Zero Bet Quiz", challengeID)
		questionID := addQuestionWithBetting(t, quizID, "Zero bet?", map[string]any{
			"bettingEnabled":       true,
			"bettingMinPercentage": 50.0,
			"bettingMinAbsolute":   100,
		})

		submissionID := startQuiz(t, quizID, userID, userToken)

		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					id
					betAmount
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{},
				"timeSpentSeconds":  10,
				"betAmount":         0, // Zero bet should be rejected when betting is enabled
			},
		})
		require.True(t, resp.HasErrors(), "zero bet should be rejected when betting is enabled")
		assert.Contains(t, resp.ErrorMessage(), "bet is required when betting is enabled")
	})

	// ==================== BETTING FIELDS ON ALL QUESTION TYPES ====================

	t.Run("betting fields available on FreeText question", func(t *testing.T) {
		challengeID := createChallenge(t, "FreeText Betting Challenge")
		quizID := createQuiz(t, "FreeText Betting Quiz", challengeID)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on FreeTextQuestion {
						id
						bettingEnabled
						bettingMinPercentage
						bettingMaxPercentage
						bettingMinAbsolute
						bettingMaxAbsolute
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":         "FREE_TEXT",
				"questionText":         "Free text with betting?",
				"questionOrder":        0,
				"points":               10,
				"bettingEnabled":       true,
				"bettingMaxPercentage": 30.0,
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion struct {
				ID                   string   `json:"id"`
				BettingEnabled       bool     `json:"bettingEnabled"`
				BettingMaxPercentage *float64 `json:"bettingMaxPercentage"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.True(t, result.AddQuizQuestion.BettingEnabled)
		assert.Equal(t, 30.0, *result.AddQuizQuestion.BettingMaxPercentage)
	})

	t.Run("betting fields available on Number question", func(t *testing.T) {
		challengeID := createChallenge(t, "Number Betting Challenge")
		quizID := createQuiz(t, "Number Betting Quiz", challengeID)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on NumberQuestion {
						id
						bettingEnabled
						bettingMinAbsolute
						bettingMaxAbsolute
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":       "NUMBER",
				"questionText":       "Number with betting?",
				"questionOrder":      0,
				"points":             10,
				"minValue":           0,
				"maxValue":           100,
				"bettingEnabled":     true,
				"bettingMinAbsolute": 10,
				"bettingMaxAbsolute": 50,
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion struct {
				ID                 string `json:"id"`
				BettingEnabled     bool   `json:"bettingEnabled"`
				BettingMinAbsolute *int   `json:"bettingMinAbsolute"`
				BettingMaxAbsolute *int   `json:"bettingMaxAbsolute"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.True(t, result.AddQuizQuestion.BettingEnabled)
		assert.Equal(t, 10, *result.AddQuizQuestion.BettingMinAbsolute)
		assert.Equal(t, 50, *result.AddQuizQuestion.BettingMaxAbsolute)
	})

	t.Run("betting fields available on Ordering question", func(t *testing.T) {
		challengeID := createChallenge(t, "Ordering Betting Challenge")
		quizID := createQuiz(t, "Ordering Betting Quiz", challengeID)

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on OrderingQuestion {
						id
						bettingEnabled
						bettingMinPercentage
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":         "ORDERING",
				"questionText":         "Order with betting?",
				"questionOrder":        0,
				"points":               10,
				"bettingEnabled":       true,
				"bettingMinPercentage": 15.0,
				"orderingItems": []map[string]any{
					{"itemText": "First", "correctOrder": 1},
					{"itemText": "Second", "correctOrder": 2},
					{"itemText": "Third", "correctOrder": 3},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion struct {
				ID                   string   `json:"id"`
				BettingEnabled       bool     `json:"bettingEnabled"`
				BettingMinPercentage *float64 `json:"bettingMinPercentage"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.True(t, result.AddQuizQuestion.BettingEnabled)
		assert.Equal(t, 15.0, *result.AddQuizQuestion.BettingMinPercentage)
	})
}
