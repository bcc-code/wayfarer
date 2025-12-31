package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuizLifecycle(t *testing.T) {
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
	eventID := data.EventIDs[projectID][0] // Use first seeded event

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

	// Helper to create a quiz with timing options
	type quizOptions struct {
		publishedAt    *string
		endTime        *string
		timeoutSeconds *int
		allowRetakes   bool
	}

	createQuiz := func(t *testing.T, name string, challengeID string, opts quizOptions) string {
		input := map[string]any{
			"projectId":            projectID,
			"challengeId":          challengeID,
			"name":                 name,
			"description":          "Test quiz",
			"randomizeQuestions":   false,
			"revealCorrectAnswers": true,
			"allowRetakes":         opts.allowRetakes,
			"completionPoints":     10,
		}
		if opts.publishedAt != nil {
			input["publishedAt"] = *opts.publishedAt
		}
		if opts.endTime != nil {
			input["endTime"] = *opts.endTime
		}
		if opts.timeoutSeconds != nil {
			input["timeoutSeconds"] = *opts.timeoutSeconds
		}

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateQuiz($input: CreateQuizInput!) {
				createQuiz(input: $input) {
					id
					name
					publishedAt
					endTime
					timeoutSeconds
					allowRetakes
				}
			}
		`, map[string]any{"input": input})
		require.False(t, resp.HasErrors(), "failed to create quiz: %s", resp.ErrorMessage())

		var result struct {
			CreateQuiz struct {
				ID             string  `json:"id"`
				Name           string  `json:"name"`
				PublishedAt    *string `json:"publishedAt"`
				EndTime        *string `json:"endTime"`
				TimeoutSeconds *int    `json:"timeoutSeconds"`
				AllowRetakes   bool    `json:"allowRetakes"`
			} `json:"createQuiz"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.CreateQuiz.ID
	}

	// Helper to add a question to quiz
	addQuestion := func(t *testing.T, quizID string, questionText string, order int, correctAnswer string) string {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on PredefinedQuestion {
						id
						questionText
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":  "PREDEFINED",
				"questionText":  questionText,
				"questionOrder": order,
				"predefinedAnswers": []map[string]any{
					{"answerText": correctAnswer, "isCorrect": true, "answerOrder": 0},
					{"answerText": "Wrong answer", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion struct {
				ID string `json:"id"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AddQuizQuestion.ID
	}

	// ==================== TIMING EDGE CASE TESTS ====================

	t.Run("cannot start quiz before published", func(t *testing.T) {
		challengeID := createChallenge(t, "Unpublished Quiz Challenge")

		// Create quiz with publishedAt in the future
		futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		quizID := createQuiz(t, "Unpublished Quiz", challengeID, quizOptions{
			publishedAt: &futureTime,
		})
		addQuestion(t, quizID, "Question 1?", 0, "Answer 1")

		// Try to start the quiz
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
				}
			}
		`, map[string]any{"quizId": quizID})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "quiz is not available")
	})

	t.Run("cannot start quiz after end time", func(t *testing.T) {
		challengeID := createChallenge(t, "Ended Quiz Challenge")

		// Create quiz with endTime in the past
		pastTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		publishedTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
		quizID := createQuiz(t, "Ended Quiz", challengeID, quizOptions{
			publishedAt: &publishedTime,
			endTime:     &pastTime,
		})
		addQuestion(t, quizID, "Question 1?", 0, "Answer 1")

		// Try to start the quiz
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
				}
			}
		`, map[string]any{"quizId": quizID})

		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "quiz has ended")
	})

	t.Run("quiz timeout sets submission expiry", func(t *testing.T) {
		challengeID := createChallenge(t, "Timed Quiz Challenge")

		// Create quiz with 60 second timeout
		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		timeout := 60
		quizID := createQuiz(t, "Timed Quiz", challengeID, quizOptions{
			publishedAt:    &publishedTime,
			timeoutSeconds: &timeout,
			allowRetakes:   true,
		})
		addQuestion(t, quizID, "Question 1?", 0, "Answer 1")

		// Start the quiz
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
					expiresAt
					startedAt
				}
			}
		`, map[string]any{"quizId": quizID})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			StartQuiz struct {
				ID        string  `json:"id"`
				ExpiresAt *string `json:"expiresAt"`
				StartedAt string  `json:"startedAt"`
			} `json:"startQuiz"`
		}
		require.NoError(t, resp.UnmarshalData(&result))

		assert.NotEmpty(t, result.StartQuiz.ID)
		assert.NotNil(t, result.StartQuiz.ExpiresAt, "expiresAt should be set for timed quiz")
	})

	t.Run("cannot submit answer after submission expired", func(t *testing.T) {
		challengeID := createChallenge(t, "Expired Submission Challenge")

		// Create quiz with very short timeout (1 second)
		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		timeout := 1 // 1 second timeout
		quizID := createQuiz(t, "Quick Expire Quiz", challengeID, quizOptions{
			publishedAt:    &publishedTime,
			timeoutSeconds: &timeout,
			allowRetakes:   true,
		})

		// Get question ID and answer IDs
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on PredefinedQuestion {
						id
						predefinedAnswers {
							id
							answerText
							isCorrect
						}
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":  "PREDEFINED",
				"questionText":  "Quick question?",
				"questionOrder": 0,
				"predefinedAnswers": []map[string]any{
					{"answerText": "Correct", "isCorrect": true, "answerOrder": 0},
					{"answerText": "Wrong", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var questionResult struct {
			AddQuizQuestion struct {
				ID                string `json:"id"`
				PredefinedAnswers []struct {
					ID        string `json:"id"`
					IsCorrect bool   `json:"isCorrect"`
				} `json:"predefinedAnswers"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&questionResult))
		questionID := questionResult.AddQuizQuestion.ID

		var correctAnswerID string
		for _, ans := range questionResult.AddQuizQuestion.PredefinedAnswers {
			if ans.IsCorrect {
				correctAnswerID = ans.ID
				break
			}
		}

		// Start the quiz
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
					expiresAt
				}
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, startResp.HasErrors())

		var startResult struct {
			StartQuiz struct {
				ID        string  `json:"id"`
				ExpiresAt *string `json:"expiresAt"`
			} `json:"startQuiz"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		submissionID := startResult.StartQuiz.ID

		// Wait for the submission to expire
		time.Sleep(2 * time.Second)

		// Try to submit an answer - should fail
		submitResp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					... on PredefinedResponse {
						id
					}
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{correctAnswerID},
			},
		})

		require.True(t, submitResp.HasErrors())
		assert.Contains(t, submitResp.ErrorMessage(), "expired")
	})

	t.Run("cannot finalize after submission expired", func(t *testing.T) {
		challengeID := createChallenge(t, "Expired Finalize Challenge")

		// Create quiz with very short timeout
		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		timeout := 1
		quizID := createQuiz(t, "Quick Expire Finalize Quiz", challengeID, quizOptions{
			publishedAt:    &publishedTime,
			timeoutSeconds: &timeout,
			allowRetakes:   true,
		})
		addQuestion(t, quizID, "Question?", 0, "Answer")

		// Start the quiz
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
				}
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, startResp.HasErrors())

		var startResult struct {
			StartQuiz struct{ ID string } `json:"startQuiz"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		submissionID := startResult.StartQuiz.ID

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Try to finalize - should fail
		finalizeResp := client.WithAuth(userToken).MustExecute(t, `
			mutation FinalizeQuiz($submissionId: ID!) {
				finalizeQuiz(submissionId: $submissionId) {
					id
				}
			}
		`, map[string]any{"submissionId": submissionID})

		require.True(t, finalizeResp.HasErrors())
		assert.Contains(t, finalizeResp.ErrorMessage(), "expired")
	})

	t.Run("retakes not allowed prevents second attempt", func(t *testing.T) {
		challengeID := createChallenge(t, "No Retake Challenge")

		// Create quiz without retakes
		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		quizID := createQuiz(t, "No Retake Quiz", challengeID, quizOptions{
			publishedAt:  &publishedTime,
			allowRetakes: false,
		})

		// Add question and get answer ID
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on PredefinedQuestion {
						id
						predefinedAnswers { id isCorrect }
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":  "PREDEFINED",
				"questionText":  "Retake test question?",
				"questionOrder": 0,
				"predefinedAnswers": []map[string]any{
					{"answerText": "Yes", "isCorrect": true, "answerOrder": 0},
					{"answerText": "No", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var qResult struct {
			AddQuizQuestion struct {
				ID                string `json:"id"`
				PredefinedAnswers []struct {
					ID        string `json:"id"`
					IsCorrect bool   `json:"isCorrect"`
				} `json:"predefinedAnswers"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&qResult))

		questionID := qResult.AddQuizQuestion.ID
		var correctAnswerID string
		for _, ans := range qResult.AddQuizQuestion.PredefinedAnswers {
			if ans.IsCorrect {
				correctAnswerID = ans.ID
				break
			}
		}

		// Start and complete the quiz
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
				}
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, startResp.HasErrors())

		var startResult struct {
			StartQuiz struct{ ID string } `json:"startQuiz"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		submissionID := startResult.StartQuiz.ID

		// Submit answer
		submitResp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					... on PredefinedResponse {
						id
					}
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{correctAnswerID},
			},
		})
		require.False(t, submitResp.HasErrors())

		// Finalize
		finalizeResp := client.WithAuth(userToken).MustExecute(t, `
			mutation FinalizeQuiz($submissionId: ID!) {
				finalizeQuiz(submissionId: $submissionId) {
					id
					completedAt
				}
			}
		`, map[string]any{"submissionId": submissionID})
		require.False(t, finalizeResp.HasErrors(), "failed to finalize: %s", finalizeResp.ErrorMessage())

		// Try to start again - should fail
		retakeResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
				}
			}
		`, map[string]any{"quizId": quizID})

		require.True(t, retakeResp.HasErrors())
		assert.Contains(t, retakeResp.ErrorMessage(), "retakes not allowed")
	})

	t.Run("resume active submission instead of creating new", func(t *testing.T) {
		challengeID := createChallenge(t, "Resume Quiz Challenge")

		// Create quiz with retakes allowed
		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		quizID := createQuiz(t, "Resume Quiz", challengeID, quizOptions{
			publishedAt:  &publishedTime,
			allowRetakes: true,
		})
		addQuestion(t, quizID, "Question?", 0, "Answer")

		// Start the quiz first time
		startResp1 := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
					startedAt
				}
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, startResp1.HasErrors())

		var result1 struct {
			StartQuiz struct {
				ID        string `json:"id"`
				StartedAt string `json:"startedAt"`
			} `json:"startQuiz"`
		}
		require.NoError(t, startResp1.UnmarshalData(&result1))
		firstSubmissionID := result1.StartQuiz.ID

		// Start again - should return same submission
		startResp2 := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
					startedAt
				}
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, startResp2.HasErrors())

		var result2 struct {
			StartQuiz struct {
				ID        string `json:"id"`
				StartedAt string `json:"startedAt"`
			} `json:"startQuiz"`
		}
		require.NoError(t, startResp2.UnmarshalData(&result2))

		assert.Equal(t, firstSubmissionID, result2.StartQuiz.ID, "should return same active submission")
	})

	t.Run("full quiz lifecycle happy path", func(t *testing.T) {
		challengeID := createChallenge(t, "Full Quiz Challenge")

		// Create quiz
		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		quizID := createQuiz(t, "Full Quiz", challengeID, quizOptions{
			publishedAt:  &publishedTime,
			allowRetakes: false,
		})

		// Add multiple questions
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {
					... on PredefinedQuestion {
						id
						predefinedAnswers { id isCorrect }
					}
				}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":  "PREDEFINED",
				"questionText":  "What is 2+2?",
				"questionOrder": 0,
				"points":        5,
				"predefinedAnswers": []map[string]any{
					{"answerText": "4", "isCorrect": true, "answerOrder": 0},
					{"answerText": "5", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var q1Result struct {
			AddQuizQuestion struct {
				ID                string `json:"id"`
				PredefinedAnswers []struct {
					ID        string `json:"id"`
					IsCorrect bool   `json:"isCorrect"`
				} `json:"predefinedAnswers"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&q1Result))

		question1ID := q1Result.AddQuizQuestion.ID
		var correctAnswer1ID string
		for _, ans := range q1Result.AddQuizQuestion.PredefinedAnswers {
			if ans.IsCorrect {
				correctAnswer1ID = ans.ID
				break
			}
		}

		// Start quiz
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) {
					id
					startedAt
					maxScore
					questionOrder
				}
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, startResp.HasErrors(), "failed to start: %s", startResp.ErrorMessage())

		var startResult struct {
			StartQuiz struct {
				ID            string   `json:"id"`
				StartedAt     string   `json:"startedAt"`
				MaxScore      int      `json:"maxScore"`
				QuestionOrder []string `json:"questionOrder"`
			} `json:"startQuiz"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		submissionID := startResult.StartQuiz.ID

		assert.NotEmpty(t, submissionID)
		assert.NotEmpty(t, startResult.StartQuiz.StartedAt)
		assert.Equal(t, 1, startResult.StartQuiz.MaxScore)
		assert.Len(t, startResult.StartQuiz.QuestionOrder, 1)

		// Submit answer
		submitResp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {
					... on PredefinedResponse {
						id
						isCorrect
						pointsEarned
					}
				}
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        question1ID,
				"selectedAnswerIds": []string{correctAnswer1ID},
				"timeSpentSeconds":  10,
			},
		})
		require.False(t, submitResp.HasErrors(), "failed to submit: %s", submitResp.ErrorMessage())

		var submitResult struct {
			SubmitQuizAnswer struct {
				ID           string `json:"id"`
				IsCorrect    *bool  `json:"isCorrect"`
				PointsEarned *int   `json:"pointsEarned"`
			} `json:"submitQuizAnswer"`
		}
		require.NoError(t, submitResp.UnmarshalData(&submitResult))

		assert.NotNil(t, submitResult.SubmitQuizAnswer.IsCorrect)
		assert.True(t, *submitResult.SubmitQuizAnswer.IsCorrect)
		assert.NotNil(t, submitResult.SubmitQuizAnswer.PointsEarned)
		assert.Equal(t, 5, *submitResult.SubmitQuizAnswer.PointsEarned)

		// Finalize quiz
		finalizeResp := client.WithAuth(userToken).MustExecute(t, `
			mutation FinalizeQuiz($submissionId: ID!) {
				finalizeQuiz(submissionId: $submissionId) {
					id
					completedAt
					score
					pointsAwarded
				}
			}
		`, map[string]any{"submissionId": submissionID})
		require.False(t, finalizeResp.HasErrors(), "failed to finalize: %s", finalizeResp.ErrorMessage())

		var finalizeResult struct {
			FinalizeQuiz struct {
				ID            string  `json:"id"`
				CompletedAt   *string `json:"completedAt"`
				Score         *int    `json:"score"`
				PointsAwarded *int    `json:"pointsAwarded"`
			} `json:"finalizeQuiz"`
		}
		require.NoError(t, finalizeResp.UnmarshalData(&finalizeResult))

		assert.NotNil(t, finalizeResult.FinalizeQuiz.CompletedAt)
		assert.NotNil(t, finalizeResult.FinalizeQuiz.Score)
		assert.Equal(t, 1, *finalizeResult.FinalizeQuiz.Score) // 1 correct answer
		assert.NotNil(t, finalizeResult.FinalizeQuiz.PointsAwarded)
		// Points = completion points (10) + per-answer points (5)
		assert.Equal(t, 15, *finalizeResult.FinalizeQuiz.PointsAwarded)
	})

	t.Run("cannot finalize already completed submission", func(t *testing.T) {
		challengeID := createChallenge(t, "Double Finalize Challenge")

		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		quizID := createQuiz(t, "Double Finalize Quiz", challengeID, quizOptions{
			publishedAt:  &publishedTime,
			allowRetakes: true,
		})
		addQuestion(t, quizID, "Q?", 0, "A")

		// Start and finalize
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) { id }
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, startResp.HasErrors())

		var startResult struct {
			StartQuiz struct{ ID string } `json:"startQuiz"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		submissionID := startResult.StartQuiz.ID

		// First finalize
		finalizeResp1 := client.WithAuth(userToken).MustExecute(t, `
			mutation FinalizeQuiz($submissionId: ID!) {
				finalizeQuiz(submissionId: $submissionId) { id completedAt }
			}
		`, map[string]any{"submissionId": submissionID})
		require.False(t, finalizeResp1.HasErrors())

		// Second finalize - should fail
		finalizeResp2 := client.WithAuth(userToken).MustExecute(t, `
			mutation FinalizeQuiz($submissionId: ID!) {
				finalizeQuiz(submissionId: $submissionId) { id }
			}
		`, map[string]any{"submissionId": submissionID})

		require.True(t, finalizeResp2.HasErrors())
		assert.Contains(t, finalizeResp2.ErrorMessage(), "already completed")
	})

	t.Run("user cannot access other users submission", func(t *testing.T) {
		challengeID := createChallenge(t, "Auth Quiz Challenge")

		publishedTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		quizID := createQuiz(t, "Auth Quiz", challengeID, quizOptions{
			publishedAt:  &publishedTime,
			allowRetakes: true,
		})
		addQuestion(t, quizID, "Q?", 0, "A")

		// User 1 starts quiz
		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartQuiz($quizId: ID!) {
				startQuiz(quizId: $quizId) { id }
			}
		`, map[string]any{"quizId": quizID})
		require.False(t, startResp.HasErrors())

		var startResult struct {
			StartQuiz struct{ ID string } `json:"startQuiz"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		submissionID := startResult.StartQuiz.ID

		// User 2 tries to finalize User 1's submission
		user2ID := data.UserIDs[2]
		user2Token, err := testutil.GenerateUserToken(user2ID)
		require.NoError(t, err)

		finalizeResp := client.WithAuth(user2Token).MustExecute(t, `
			mutation FinalizeQuiz($submissionId: ID!) {
				finalizeQuiz(submissionId: $submissionId) { id }
			}
		`, map[string]any{"submissionId": submissionID})

		require.True(t, finalizeResp.HasErrors())
		assert.Contains(t, finalizeResp.ErrorMessage(), "unauthorized")
	})
}
