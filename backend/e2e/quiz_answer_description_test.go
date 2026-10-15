package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuizAnswerDescription(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	adminUserID := data.UserIDs[2]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]

	type answerResult struct {
		ID          string  `json:"id"`
		AnswerText  string  `json:"answerText"`
		Description *string `json:"description"`
		AnswerOrder int     `json:"answerOrder"`
	}

	type questionResult struct {
		ID                string         `json:"id"`
		PredefinedAnswers []answerResult `json:"predefinedAnswers"`
	}

	const questionFields = `
		... on PredefinedQuestion {
			id
			predefinedAnswers {
				id
				answerText
				description
				answerOrder
			}
		}
	`

	// createQuiz creates a quiz challenge with an empty quiz and returns the quiz ID.
	createQuiz := func(t *testing.T, name string) string {
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
				"buttonText":  "Gjett",
				"description": "<p>Guess this week's item</p>",
			},
		})
		require.False(t, resp.HasErrors(), "failed to create challenge: %s", resp.ErrorMessage())

		var challenge struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, resp.UnmarshalData(&challenge))

		resp = client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateQuiz($input: CreateQuizInput!) {
				createQuiz(input: $input) {
					id
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":            projectID,
				"challengeId":          challenge.CreateChallenge.ID,
				"name":                 name,
				"description":          "Which item is this week's?",
				"randomizeQuestions":   false,
				"revealCorrectAnswers": true,
				"allowRetakes":         false,
				"completionPoints":     0,
			},
		})
		require.False(t, resp.HasErrors(), "failed to create quiz: %s", resp.ErrorMessage())

		var quiz struct {
			CreateQuiz struct{ ID string } `json:"createQuiz"`
		}
		require.NoError(t, resp.UnmarshalData(&quiz))
		return quiz.CreateQuiz.ID
	}

	addQuestion := func(t *testing.T, quizID string, answers []map[string]any) questionResult {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {`+questionFields+`}
			}
		`, map[string]any{
			"quizId": quizID,
			"input": map[string]any{
				"questionType":      "PREDEFINED",
				"questionText":      "Hvilken gjenstand er ukens?",
				"questionOrder":     0,
				"points":            10,
				"predefinedAnswers": answers,
			},
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())

		var result struct {
			AddQuizQuestion questionResult `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AddQuizQuestion
	}

	// readQuestion loads the quiz's only question through the quiz query, a
	// read path separate from the mutation response.
	readQuestion := func(t *testing.T, quizID string) questionResult {
		resp := client.WithAuth(adminToken).MustExecute(t, `
			query GetQuiz($id: ID!) {
				quiz(id: $id) {
					questions {`+questionFields+`}
				}
			}
		`, map[string]any{"id": quizID})
		require.False(t, resp.HasErrors(), "failed to read quiz: %s", resp.ErrorMessage())

		var result struct {
			Quiz struct {
				Questions []questionResult `json:"questions"`
			} `json:"quiz"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		require.Len(t, result.Quiz.Questions, 1)
		return result.Quiz.Questions[0]
	}

	// descriptionsByText maps each answer's text to its description so
	// assertions don't depend on answer order.
	descriptionsByText := func(t *testing.T, answers []answerResult) map[string]*string {
		out := make(map[string]*string, len(answers))
		for _, a := range answers {
			out[a.AnswerText] = a.Description
		}
		return out
	}

	strPtr := func(s string) *string { return &s }

	t.Run("descriptions are stored and returned", func(t *testing.T) {
		quizID := createQuiz(t, "Answer Description Create")

		created := addQuestion(t, quizID, []map[string]any{
			{
				"answerText":  "Kjøttgryten fra Egypt",
				"description": "En gryte som symboliserer maten israelittene lengtet tilbake til.",
				"isCorrect":   true,
				"answerOrder": 0,
			},
			{"answerText": "Fangelenken fra Egypt", "isCorrect": false, "answerOrder": 1},
		})

		want := map[string]*string{
			"Kjøttgryten fra Egypt": strPtr("En gryte som symboliserer maten israelittene lengtet tilbake til."),
			"Fangelenken fra Egypt": nil,
		}
		assert.Equal(t, want, descriptionsByText(t, created.PredefinedAnswers), "mutation response")
		assert.Equal(t, want, descriptionsByText(t, readQuestion(t, quizID).PredefinedAnswers), "quiz query")
	})

	t.Run("blank descriptions are stored as null and others are trimmed", func(t *testing.T) {
		quizID := createQuiz(t, "Answer Description Blank")

		addQuestion(t, quizID, []map[string]any{
			{"answerText": "Melkekrukken", "description": "", "isCorrect": true, "answerOrder": 0},
			{"answerText": "Melkisedeks kappe", "description": "  \n\t ", "isCorrect": false, "answerOrder": 1},
			{"answerText": "Abrahams ættetavle", "description": "  Slekt og etterkommere.  ", "isCorrect": false, "answerOrder": 2},
		})

		assert.Equal(t, map[string]*string{
			"Melkekrukken":       nil,
			"Melkisedeks kappe":  nil,
			"Abrahams ættetavle": strPtr("Slekt og etterkommere."),
		}, descriptionsByText(t, readQuestion(t, quizID).PredefinedAnswers))
	})

	t.Run("updating answers replaces their descriptions", func(t *testing.T) {
		quizID := createQuiz(t, "Answer Description Update")

		created := addQuestion(t, quizID, []map[string]any{
			{"answerText": "Den blodstrøkne dørkarmen", "description": "Gammel beskrivelse.", "isCorrect": true, "answerOrder": 0},
			{"answerText": "Melkekrukken", "description": "Skal fjernes.", "isCorrect": false, "answerOrder": 1},
		})

		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateQuestion($id: ID!, $input: UpdateQuizQuestionInput!) {
				updateQuizQuestion(id: $id, input: $input) {`+questionFields+`}
			}
		`, map[string]any{
			"id": created.ID,
			"input": map[string]any{
				"predefinedAnswers": []map[string]any{
					{"answerText": "Den blodstrøkne dørkarmen", "description": "Ny beskrivelse.", "isCorrect": true, "answerOrder": 0},
					{"answerText": "Melkekrukken", "isCorrect": false, "answerOrder": 1},
				},
			},
		})
		require.False(t, resp.HasErrors(), "failed to update question: %s", resp.ErrorMessage())

		var updated struct {
			UpdateQuizQuestion questionResult `json:"updateQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&updated))

		want := map[string]*string{
			"Den blodstrøkne dørkarmen": strPtr("Ny beskrivelse."),
			"Melkekrukken":              nil,
		}
		assert.Equal(t, want, descriptionsByText(t, updated.UpdateQuizQuestion.PredefinedAnswers), "mutation response")
		assert.Equal(t, want, descriptionsByText(t, readQuestion(t, quizID).PredefinedAnswers), "quiz query")
	})
}
