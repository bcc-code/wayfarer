package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bcc-media/wayfarer/common"
	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/translations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
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

	t.Run("descriptions round-trip through translation export and import", func(t *testing.T) {
		quizID := createQuiz(t, "Answer Description Translation")
		addQuestion(t, quizID, []map[string]any{
			{"answerText": "Kjøttgryten fra Egypt", "description": "En gryte som symboliserer maten.", "isCorrect": true, "answerOrder": 0},
			{"answerText": "Fangelenken fra Egypt", "isCorrect": false, "answerOrder": 1},
		})

		provider := &capturingTranslationsProvider{}
		svc := translations.NewService(dbMgr.DB.Queries, provider)

		// Export: the described answer carries its description, the other has none
		require.NoError(t, svc.SendCollectionToTranslation(ctx, translations.CollectionQuizzes))
		exported := provider.find(t, quizID)

		var quiz translations.QuizTranslation
		require.NoError(t, json.Unmarshal(exported.Value, &quiz))
		require.Len(t, quiz.Questions, 1)
		require.Len(t, quiz.Questions[0].Answers, 2)
		pot, chain := quiz.Questions[0].Answers[0], quiz.Questions[0].Answers[1]
		assert.Equal(t, null.StringFrom("En gryte som symboliserer maten."), pot.Description)
		assert.False(t, chain.Description.Valid, "answer without description must not export one")

		// Import an English translation, the way the Phrase webhook does
		quiz.Name = "Answer Description Translation (en)"
		quiz.Questions[0].QuestionText = "Which item is this week's?"
		quiz.Questions[0].Answers[0].AnswerText = "The meat pot from Egypt"
		quiz.Questions[0].Answers[0].Description = null.StringFrom("A pot symbolising the food.")
		quiz.Questions[0].Answers[1].AnswerText = "The chain from Egypt"
		value, err := json.Marshal(quiz)
		require.NoError(t, err)

		errs := svc.UpdateTranslations(ctx, &translations.CollectionQuizzes, []common.TranslationData{
			{Language: "en", ID: quizID, Value: value},
		})
		require.Empty(t, errs)

		type translatedAnswer struct {
			AnswerText        string  `json:"answerText"`
			Description       *string `json:"description"`
			TranslationStatus []struct {
				LanguageCode string   `json:"languageCode"`
				Fields       []string `json:"fields"`
			} `json:"translationStatus"`
		}
		readAnswers := func(t *testing.T, lang string) []translatedAnswer {
			resp := client.WithAuth(adminToken).WithLanguage(lang).MustExecute(t, `
				query GetQuiz($id: ID!) {
					quiz(id: $id) {
						questions {
							... on PredefinedQuestion {
								predefinedAnswers {
									answerText
									description
									translationStatus {
										languageCode
										fields
									}
								}
							}
						}
					}
				}
			`, map[string]any{"id": quizID})
			require.False(t, resp.HasErrors(), "failed to read quiz: %s", resp.ErrorMessage())

			var result struct {
				Quiz struct {
					Questions []struct {
						PredefinedAnswers []translatedAnswer `json:"predefinedAnswers"`
					} `json:"questions"`
				} `json:"quiz"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			require.Len(t, result.Quiz.Questions, 1)
			require.Len(t, result.Quiz.Questions[0].PredefinedAnswers, 2)
			return result.Quiz.Questions[0].PredefinedAnswers
		}

		english := readAnswers(t, "en")
		assert.Equal(t, "The meat pot from Egypt", english[0].AnswerText)
		assert.Equal(t, strPtr("A pot symbolising the food."), english[0].Description)
		assert.Equal(t, "The chain from Egypt", english[1].AnswerText)
		assert.Nil(t, english[1].Description, "answer without a Norwegian description gets none in English")

		require.Len(t, english[0].TranslationStatus, 1)
		assert.Equal(t, "en", english[0].TranslationStatus[0].LanguageCode)
		assert.ElementsMatch(t, []string{"answerText", "description"}, english[0].TranslationStatus[0].Fields)
		require.Len(t, english[1].TranslationStatus, 1)
		assert.ElementsMatch(t, []string{"answerText"}, english[1].TranslationStatus[0].Fields)

		// Norwegian is untouched by the import
		norwegian := readAnswers(t, "no")
		assert.Equal(t, "Kjøttgryten fra Egypt", norwegian[0].AnswerText)
		assert.Equal(t, strPtr("En gryte som symboliserer maten."), norwegian[0].Description)
	})
}

// capturingTranslationsProvider records what the translation service exports
// instead of sending it to Phrase.
type capturingTranslationsProvider struct {
	sent []common.TranslationData
}

func (p *capturingTranslationsProvider) SendToTranslation(_ context.Context, _ string, data []common.TranslationData) error {
	p.sent = append(p.sent, data...)
	return nil
}

func (p *capturingTranslationsProvider) ProcessWebhook(context.Context, *http.Request, []byte) (*translations.TranslatableCollection, []common.TranslationData, error) {
	return nil, nil, nil
}

func (p *capturingTranslationsProvider) find(t *testing.T, id string) common.TranslationData {
	t.Helper()
	for _, d := range p.sent {
		if d.ID == id {
			return d
		}
	}
	require.FailNow(t, "entity was not exported", id)
	return common.TranslationData{}
}
