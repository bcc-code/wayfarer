package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuizAnswerVisibility covers the session gates on the two fields that give
// away a correct answer: QuizPredefinedAnswer.isCorrect (which answers are right)
// and QuizResponse.isCorrect (whether the user's own answer was right).
func TestQuizAnswerVisibility(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	userID := data.UserIDs[0]
	adminUserID := data.UserIDs[2]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

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

	createChallenge := func(t *testing.T, name string) string {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateChallenge($projectId: ID!, $eventId: ID!, $input: CreateChallengeInput!) {
				createChallenge(projectId: $projectId, eventId: $eventId, input: $input) { id }
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

		publishedAt := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		for _, m := range []string{
			`mutation P($id: ID!, $t: DateTime!) { publishChallenge(id: $id, publishedAt: $t) { id } }`,
			`mutation V($id: ID!, $t: DateTime!) { setChallengeVisibility(id: $id, visibleAt: $t) { id } }`,
		} {
			r := client.WithAuth(adminToken).MustExecute(t, m, map[string]any{"id": result.CreateChallenge.ID, "t": publishedAt})
			require.False(t, r.HasErrors(), "failed to publish/make visible: %s", r.ErrorMessage())
		}
		return result.CreateChallenge.ID
	}

	createQuiz := func(t *testing.T, name, challengeID string) string {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateQuiz($input: CreateQuizInput!) {
				createQuiz(input: $input) { id }
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":            projectID,
				"challengeId":          challengeID,
				"name":                 name,
				"description":          "Test quiz",
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

	// addQuestion returns the question ID and the ID of its correct answer.
	addQuestion := func(t *testing.T, quizID string) (string, string) {
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
				"questionText":  "Which one is right?",
				"questionOrder": 0,
				"predefinedAnswers": []map[string]any{
					{"answerText": "Correct Answer", "isCorrect": true, "answerOrder": 0},
					{"answerText": "Wrong Answer", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion struct {
				ID                string `json:"id"`
				PredefinedAnswers []struct {
					ID        string `json:"id"`
					IsCorrect *bool  `json:"isCorrect"`
				} `json:"predefinedAnswers"`
			} `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		require.Len(t, result.AddQuizQuestion.PredefinedAnswers, 2)

		var correctAnswerID string
		for _, a := range result.AddQuizQuestion.PredefinedAnswers {
			if a.IsCorrect != nil && *a.IsCorrect {
				correctAnswerID = a.ID
			}
		}
		require.NotEmpty(t, correctAnswerID, "admin should see which answer is correct")
		return result.AddQuizQuestion.ID, correctAnswerID
	}

	// startSession creates a session, grants the user access, opens it and starts
	// the user's submission. It returns the session and submission IDs.
	startSession := func(t *testing.T, quizID string) (string, string) {
		createResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) {
				createQuizSession(input: $input) { id }
			}
		`, map[string]any{"input": map[string]any{"quizId": quizID}})
		require.False(t, createResp.HasErrors(), "failed to create session: %s", createResp.ErrorMessage())

		var createResult struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, createResp.UnmarshalData(&createResult))
		sessionID := createResult.CreateQuizSession.ID

		grantResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) {
				grantQuizSessionAccess(input: $input)
			}
		`, map[string]any{"input": map[string]any{"sessionId": sessionID, "userIds": []string{userID}}})
		require.False(t, grantResp.HasErrors(), "failed to grant access: %s", grantResp.ErrorMessage())

		openResp := client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) { openQuizSession(id: $id) { id } }
		`, map[string]any{"id": sessionID})
		require.False(t, openResp.HasErrors(), "failed to open session: %s", openResp.ErrorMessage())

		startResp := client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) { startQuizSession(sessionId: $sessionId) { id } }
		`, map[string]any{"sessionId": sessionID})
		require.False(t, startResp.HasErrors(), "failed to start session: %s", startResp.ErrorMessage())

		var startResult struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		require.NoError(t, startResp.UnmarshalData(&startResult))
		return sessionID, startResult.StartQuizSession.ID
	}

	transition := func(t *testing.T, mutation, sessionID string) {
		resp := client.WithAuth(adminToken).MustExecute(t, mutation, map[string]any{"id": sessionID})
		require.False(t, resp.HasErrors(), "failed to transition session: %s", resp.ErrorMessage())
	}
	lockSession := func(t *testing.T, sessionID string) {
		transition(t, `mutation L($id: ID!) { lockQuizSession(id: $id) { id state } }`, sessionID)
	}
	finishSession := func(t *testing.T, sessionID string) {
		transition(t, `mutation F($id: ID!) { finishQuizSession(id: $id) { id state } }`, sessionID)
	}

	// readAnswerVerdicts returns isCorrect for each of the question's answers, as
	// seen by the holder of the given token.
	readAnswerVerdicts := func(t *testing.T, quizID, token string) []*bool {
		resp := client.WithAuth(token).MustExecute(t, `
			query GetQuiz($id: ID!) {
				quiz(id: $id) {
					questions {
						... on PredefinedQuestion {
							predefinedAnswers { id isCorrect }
						}
					}
				}
			}
		`, map[string]any{"id": quizID})
		require.False(t, resp.HasErrors(), "failed to read quiz: %s", resp.ErrorMessage())

		var result struct {
			Quiz struct {
				Questions []struct {
					PredefinedAnswers []struct {
						ID        string `json:"id"`
						IsCorrect *bool  `json:"isCorrect"`
					} `json:"predefinedAnswers"`
				} `json:"questions"`
			} `json:"quiz"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		require.Len(t, result.Quiz.Questions, 1)

		require.Len(t, result.Quiz.Questions[0].PredefinedAnswers, 2, "the question's answers must be readable even when the verdict is not")

		verdicts := make([]*bool, 0, len(result.Quiz.Questions[0].PredefinedAnswers))
		for _, a := range result.Quiz.Questions[0].PredefinedAnswers {
			verdicts = append(verdicts, a.IsCorrect)
		}
		return verdicts
	}

	submitAnswer := func(t *testing.T, submissionID, questionID, answerID string) {
		resp := client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) { id }
			}
		`, map[string]any{
			"submissionId": submissionID,
			"input": map[string]any{
				"questionId":        questionID,
				"selectedAnswerIds": []string{answerID},
				"timeSpentSeconds":  5,
			},
		})
		require.False(t, resp.HasErrors(), "failed to submit answer: %s", resp.ErrorMessage())
	}

	// readResponseVerdict returns the user's own response verdict for the quiz.
	readResponseVerdict := func(t *testing.T, submissionID string) *bool {
		resp := client.WithAuth(userToken).MustExecute(t, `
			query GetSubmission($id: ID!) {
				quizSubmission(id: $id) {
					responses {
						... on PredefinedResponse { isCorrect }
					}
				}
			}
		`, map[string]any{"id": submissionID})
		require.False(t, resp.HasErrors(), "failed to read submission: %s", resp.ErrorMessage())

		var result struct {
			QuizSubmission struct {
				Responses []struct {
					IsCorrect *bool `json:"isCorrect"`
				} `json:"responses"`
			} `json:"quizSubmission"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		require.Len(t, result.QuizSubmission.Responses, 1)
		return result.QuizSubmission.Responses[0].IsCorrect
	}

	assertAllNil := func(t *testing.T, verdicts []*bool, msg string) {
		for _, v := range verdicts {
			assert.Nil(t, v, msg)
		}
	}

	t.Run("answer isCorrect is hidden from the player while the session is OPEN", func(t *testing.T) {
		challengeID := createChallenge(t, "Visibility Open Challenge")
		quizID := createQuiz(t, "Visibility Open Quiz", challengeID)
		addQuestion(t, quizID)
		startSession(t, quizID)

		assertAllNil(t, readAnswerVerdicts(t, quizID, userToken),
			"a player must not be told which answer is correct while the session is OPEN")
	})

	t.Run("answer isCorrect stays hidden while the session is LOCKED", func(t *testing.T) {
		challengeID := createChallenge(t, "Visibility Locked Challenge")
		quizID := createQuiz(t, "Visibility Locked Quiz", challengeID)
		addQuestion(t, quizID)
		sessionID, _ := startSession(t, quizID)
		lockSession(t, sessionID)

		assertAllNil(t, readAnswerVerdicts(t, quizID, userToken),
			"a player must not see the correct answer in a LOCKED session")
	})

	t.Run("answer isCorrect is revealed to the player once the session is FINISHED", func(t *testing.T) {
		challengeID := createChallenge(t, "Visibility Finished Challenge")
		quizID := createQuiz(t, "Visibility Finished Quiz", challengeID)
		addQuestion(t, quizID)
		sessionID, _ := startSession(t, quizID)
		lockSession(t, sessionID)
		finishSession(t, sessionID)

		verdicts := readAnswerVerdicts(t, quizID, userToken)
		require.Len(t, verdicts, 2)
		for _, v := range verdicts {
			require.NotNil(t, v, "a FINISHED session reveals which answers were correct")
		}
		assert.True(t, *verdicts[0])
		assert.False(t, *verdicts[1])
	})

	t.Run("admins see answer isCorrect even while the session is OPEN", func(t *testing.T) {
		challengeID := createChallenge(t, "Visibility Admin Challenge")
		quizID := createQuiz(t, "Visibility Admin Quiz", challengeID)
		addQuestion(t, quizID)
		startSession(t, quizID)

		verdicts := readAnswerVerdicts(t, quizID, adminToken)
		require.Len(t, verdicts, 2)
		for _, v := range verdicts {
			require.NotNil(t, v, "the admin editor needs isCorrect regardless of session state")
		}
		assert.True(t, *verdicts[0])
		assert.False(t, *verdicts[1])
	})

	t.Run("response isCorrect is hidden while the session is OPEN and revealed after", func(t *testing.T) {
		challengeID := createChallenge(t, "Visibility Response Challenge")
		quizID := createQuiz(t, "Visibility Response Quiz", challengeID)
		questionID, correctAnswerID := addQuestion(t, quizID)
		sessionID, submissionID := startSession(t, quizID)
		submitAnswer(t, submissionID, questionID, correctAnswerID)

		assert.Nil(t, readResponseVerdict(t, submissionID),
			"a player must not learn their own verdict while they can still change the answer")

		lockSession(t, sessionID)
		verdict := readResponseVerdict(t, submissionID)
		require.NotNil(t, verdict, "a LOCKED session reveals the verdict")
		assert.True(t, *verdict)
	})
}
