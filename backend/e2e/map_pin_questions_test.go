package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mapPinQuestionFields = `
	__typename
	... on MapPinQuestion {
		id
		points
		bettingEnabled
		scoringMode
		mapWidth
		mapHeight
		fullPointsRadius
		zeroPointsRadius
		correctPosition { x y }
		correctArea { x y }
	}
`

type mapPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type mapPinQuestion struct {
	Typename         string     `json:"__typename"`
	ID               string     `json:"id"`
	Points           *int       `json:"points"`
	BettingEnabled   bool       `json:"bettingEnabled"`
	ScoringMode      string     `json:"scoringMode"`
	MapWidth         int        `json:"mapWidth"`
	MapHeight        int        `json:"mapHeight"`
	FullPointsRadius *float64   `json:"fullPointsRadius"`
	ZeroPointsRadius *float64   `json:"zeroPointsRadius"`
	CorrectPosition  *mapPoint  `json:"correctPosition"`
	CorrectArea      []mapPoint `json:"correctArea"`
}

// pointRadiusMapPin is a 1000x800 map with full points within 40 px of
// (500, 300) and none from 200 px on.
func pointRadiusMapPin() map[string]any {
	return map[string]any{
		"scoringMode":      "POINT_RADIUS",
		"mapWidth":         1000,
		"mapHeight":        800,
		"target":           map[string]any{"x": 500, "y": 300},
		"fullPointsRadius": 40,
		"zeroPointsRadius": 200,
	}
}

// trianglePinArea is a triangle on a 1000x800 map, sent with a closing vertex.
func trianglePinArea() map[string]any {
	return map[string]any{
		"scoringMode": "POLYGON",
		"mapWidth":    1000,
		"mapHeight":   800,
		"polygon": []map[string]any{
			{"x": 100, "y": 100}, {"x": 300, "y": 100}, {"x": 200, "y": 300}, {"x": 100, "y": 100},
		},
	}
}

var triangleArea = []mapPoint{{100, 100}, {300, 100}, {200, 300}}

// TestMapPinQuestions covers MAP_PIN questions through GraphQL: creating and
// updating them in both scoring modes, the rules on points, betting and edits
// after answers, and who may see the correct position or area.
func TestMapPinQuestions(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	userID := data.UserIDs[0]
	otherUserID := data.UserIDs[1]
	adminUserID := data.UserIDs[2]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

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

	createQuiz := func(t *testing.T, name string) string {
		t.Helper()
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
				"buttonText":  "Start",
				"description": "<p>Find the position in the temple</p>",
			},
		})
		require.False(t, resp.HasErrors(), "failed to create challenge: %s", resp.ErrorMessage())
		var challenge struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		require.NoError(t, resp.UnmarshalData(&challenge))

		publishedAt := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		for _, m := range []string{
			`mutation P($id: ID!, $t: DateTime!) { publishChallenge(id: $id, publishedAt: $t) { id } }`,
			`mutation V($id: ID!, $t: DateTime!) { setChallengeVisibility(id: $id, visibleAt: $t) { id } }`,
		} {
			r := client.WithAuth(adminToken).MustExecute(t, m, map[string]any{"id": challenge.CreateChallenge.ID, "t": publishedAt})
			require.False(t, r.HasErrors(), "failed to publish challenge: %s", r.ErrorMessage())
		}

		resp = client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateQuiz($input: CreateQuizInput!) {
				createQuiz(input: $input) { id }
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":            projectID,
				"challengeId":          challenge.CreateChallenge.ID,
				"name":                 name,
				"description":          "Map pin",
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

	addQuestionResp := func(t *testing.T, quizID string, input map[string]any) *testutil.GraphQLResponse {
		t.Helper()
		base := map[string]any{"questionText": "Where was this photo taken?", "questionOrder": 0}
		for k, v := range input {
			base[k] = v
		}
		return client.WithAuth(adminToken).MustExecute(t, `
			mutation AddQuestion($quizId: ID!, $input: CreateQuizQuestionInput!) {
				addQuizQuestion(quizId: $quizId, input: $input) {`+mapPinQuestionFields+`}
			}
		`, map[string]any{"quizId": quizID, "input": base})
	}

	addMapPinQuestion := func(t *testing.T, quizID string, mapPin map[string]any) mapPinQuestion {
		t.Helper()
		resp := addQuestionResp(t, quizID, map[string]any{"questionType": "MAP_PIN", "points": 100, "mapPin": mapPin})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())
		var result struct {
			AddQuizQuestion mapPinQuestion `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AddQuizQuestion
	}

	updateQuestionResp := func(t *testing.T, questionID string, input map[string]any) *testutil.GraphQLResponse {
		t.Helper()
		return client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateQuestion($id: ID!, $input: UpdateQuizQuestionInput!) {
				updateQuizQuestion(id: $id, input: $input) {`+mapPinQuestionFields+`}
			}
		`, map[string]any{"id": questionID, "input": input})
	}

	// readQuestion reads the quiz's only question through quiz.questions, which
	// goes through the questions dataloader rather than the mutation's row.
	readQuestion := func(t *testing.T, quizID, token string) mapPinQuestion {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, `
			query GetQuiz($id: ID!) {
				quiz(id: $id) { questions {`+mapPinQuestionFields+`} }
			}
		`, map[string]any{"id": quizID})
		require.False(t, resp.HasErrors(), "failed to read quiz: %s", resp.ErrorMessage())
		var result struct {
			Quiz struct {
				Questions []mapPinQuestion `json:"questions"`
			} `json:"quiz"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		require.Len(t, result.Quiz.Questions, 1)
		return result.Quiz.Questions[0]
	}

	// startSession creates a session for the user, opens it and starts their
	// submission. It returns the session and submission IDs.
	startSession := func(t *testing.T, quizID string) (string, string) {
		t.Helper()
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation CreateSession($input: CreateQuizSessionInput!) { createQuizSession(input: $input) { id } }
		`, map[string]any{"input": map[string]any{"quizId": quizID}})
		require.False(t, resp.HasErrors(), "failed to create session: %s", resp.ErrorMessage())
		var session struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		require.NoError(t, resp.UnmarshalData(&session))
		sessionID := session.CreateQuizSession.ID

		resp = client.WithAuth(adminToken).MustExecute(t, `
			mutation GrantAccess($input: GrantQuizSessionAccessInput!) { grantQuizSessionAccess(input: $input) }
		`, map[string]any{"input": map[string]any{"sessionId": sessionID, "userIds": []string{userID}}})
		require.False(t, resp.HasErrors(), "failed to grant access: %s", resp.ErrorMessage())

		resp = client.WithAuth(adminToken).MustExecute(t, `
			mutation OpenSession($id: ID!) { openQuizSession(id: $id) { id } }
		`, map[string]any{"id": sessionID})
		require.False(t, resp.HasErrors(), "failed to open session: %s", resp.ErrorMessage())

		resp = client.WithAuth(userToken).MustExecute(t, `
			mutation StartSession($sessionId: ID!) { startQuizSession(sessionId: $sessionId) { id } }
		`, map[string]any{"sessionId": sessionID})
		require.False(t, resp.HasErrors(), "failed to start session: %s", resp.ErrorMessage())
		var submission struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		require.NoError(t, resp.UnmarshalData(&submission))
		return sessionID, submission.StartQuizSession.ID
	}

	transition := func(t *testing.T, mutation, sessionID string) {
		t.Helper()
		resp := client.WithAuth(adminToken).MustExecute(t, mutation, map[string]any{"id": sessionID})
		require.False(t, resp.HasErrors(), "failed to transition session: %s", resp.ErrorMessage())
	}
	lockSession := func(t *testing.T, sessionID string) {
		transition(t, `mutation L($id: ID!) { lockQuizSession(id: $id) { id } }`, sessionID)
	}
	finishSession := func(t *testing.T, sessionID string) {
		transition(t, `mutation F($id: ID!) { finishQuizSession(id: $id) { id } }`, sessionID)
	}

	t.Run("create POINT_RADIUS", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin point radius")
		created := addMapPinQuestion(t, quizID, pointRadiusMapPin())

		for name, q := range map[string]mapPinQuestion{
			"mutation result": created,
			"quiz.questions":  readQuestion(t, quizID, adminToken),
		} {
			t.Run(name, func(t *testing.T) {
				assert.Equal(t, "MapPinQuestion", q.Typename)
				assert.Equal(t, created.ID, q.ID)
				assert.Equal(t, ptr(100), q.Points)
				assert.False(t, q.BettingEnabled)
				assert.Equal(t, "POINT_RADIUS", q.ScoringMode)
				assert.Equal(t, 1000, q.MapWidth)
				assert.Equal(t, 800, q.MapHeight)
				assert.Equal(t, ptr(40.0), q.FullPointsRadius)
				assert.Equal(t, ptr(200.0), q.ZeroPointsRadius)
				assert.Equal(t, &mapPoint{500, 300}, q.CorrectPosition, "a manager sees the target")
				assert.Nil(t, q.CorrectArea, "POINT_RADIUS has no area")
			})
		}
	})

	t.Run("create POLYGON drops the closing vertex", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin polygon")
		created := addMapPinQuestion(t, quizID, trianglePinArea())

		for name, q := range map[string]mapPinQuestion{
			"mutation result": created,
			"quiz.questions":  readQuestion(t, quizID, adminToken),
		} {
			t.Run(name, func(t *testing.T) {
				assert.Equal(t, "MapPinQuestion", q.Typename)
				assert.Equal(t, "POLYGON", q.ScoringMode)
				assert.Equal(t, 1000, q.MapWidth)
				assert.Equal(t, 800, q.MapHeight)
				assert.Equal(t, triangleArea, q.CorrectArea)
				assert.Nil(t, q.CorrectPosition, "POLYGON has no target")
				assert.Nil(t, q.FullPointsRadius)
				assert.Nil(t, q.ZeroPointsRadius)
			})
		}
	})

	t.Run("create rejects", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin create rules")

		tests := []struct {
			name    string
			input   map[string]any
			message string
		}{
			{
				name:    "MAP_PIN without mapPin",
				input:   map[string]any{"questionType": "MAP_PIN", "points": 100},
				message: "mapPin: is required",
			},
			{
				name:    "mapPin on another type",
				input:   map[string]any{"questionType": "FREE_TEXT", "points": 5, "mapPin": pointRadiusMapPin()},
				message: "mapPin: is only used by MAP_PIN questions",
			},
			{
				name:    "no points",
				input:   map[string]any{"questionType": "MAP_PIN", "mapPin": pointRadiusMapPin()},
				message: "points: must be greater than 0",
			},
			{
				name:    "zero points",
				input:   map[string]any{"questionType": "MAP_PIN", "points": 0, "mapPin": pointRadiusMapPin()},
				message: "points: must be greater than 0",
			},
			{
				name:    "betting",
				input:   map[string]any{"questionType": "MAP_PIN", "points": 100, "bettingEnabled": true, "mapPin": pointRadiusMapPin()},
				message: "bettingEnabled: MAP_PIN questions can't have betting",
			},
			{
				name: "a self-intersecting polygon",
				input: map[string]any{"questionType": "MAP_PIN", "points": 100, "mapPin": map[string]any{
					"scoringMode": "POLYGON", "mapWidth": 1000, "mapHeight": 800,
					"polygon": []map[string]any{{"x": 100, "y": 100}, {"x": 300, "y": 300}, {"x": 300, "y": 100}, {"x": 100, "y": 400}},
				}},
				message: "mapPin.polygon",
			},
			{
				name: "the other mode's field",
				input: map[string]any{"questionType": "MAP_PIN", "points": 100, "mapPin": func() map[string]any {
					m := trianglePinArea()
					m["fullPointsRadius"] = 40
					return m
				}()},
				message: "mapPin.fullPointsRadius: is only used in POINT_RADIUS mode",
			},
			{
				name: "a target outside the map",
				input: map[string]any{"questionType": "MAP_PIN", "points": 100, "mapPin": func() map[string]any {
					m := pointRadiusMapPin()
					m["target"] = map[string]any{"x": 500, "y": 801}
					return m
				}()},
				message: "mapPin.target",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp := addQuestionResp(t, quizID, tt.input)
				require.True(t, resp.HasErrors(), "expected the question to be rejected")
				assert.Contains(t, resp.ErrorMessage(), tt.message)
			})
		}

		// Nothing was stored by the rejected creates.
		questions, err := dbMgr.DB.Queries.GetQuizQuestionsByQuizID(ctx, quizID)
		require.NoError(t, err)
		assert.Empty(t, questions)
	})

	t.Run("update switches mode and clears the old mode's fields", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin mode switch")
		q := addMapPinQuestion(t, quizID, pointRadiusMapPin())

		resp := updateQuestionResp(t, q.ID, map[string]any{"mapPin": trianglePinArea()})
		require.False(t, resp.HasErrors(), "failed to update: %s", resp.ErrorMessage())

		updated := readQuestion(t, quizID, adminToken)
		assert.Equal(t, "POLYGON", updated.ScoringMode)
		assert.Equal(t, triangleArea, updated.CorrectArea)
		assert.Nil(t, updated.CorrectPosition)
		assert.Nil(t, updated.FullPointsRadius)
		assert.Nil(t, updated.ZeroPointsRadius)

		row, err := dbMgr.DB.Queries.GetQuizQuestionByID(ctx, q.ID)
		require.NoError(t, err)
		assert.Nil(t, row.MapTargetX)
		assert.Nil(t, row.MapTargetY)
		assert.Nil(t, row.MapFullPointsRadius)
		assert.Nil(t, row.MapZeroPointsRadius)

		resp = updateQuestionResp(t, q.ID, map[string]any{"mapPin": pointRadiusMapPin()})
		require.False(t, resp.HasErrors(), "failed to switch back: %s", resp.ErrorMessage())
		back := readQuestion(t, quizID, adminToken)
		assert.Equal(t, "POINT_RADIUS", back.ScoringMode)
		assert.Equal(t, &mapPoint{500, 300}, back.CorrectPosition)
		assert.Nil(t, back.CorrectArea)
	})

	t.Run("update without mapPin keeps the config", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin text edit")
		q := addMapPinQuestion(t, quizID, trianglePinArea())

		resp := updateQuestionResp(t, q.ID, map[string]any{"questionText": "Where was this taken?"})
		require.False(t, resp.HasErrors(), "failed to update: %s", resp.ErrorMessage())
		assert.Equal(t, triangleArea, readQuestion(t, quizID, adminToken).CorrectArea)
	})

	t.Run("update rejects", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin update rules")
		q := addMapPinQuestion(t, quizID, pointRadiusMapPin())

		for name, tt := range map[string]struct {
			input   map[string]any
			message string
		}{
			"zero points": {input: map[string]any{"points": 0}, message: "points: must be greater than 0"},
			"betting":     {input: map[string]any{"bettingEnabled": true}, message: "bettingEnabled: MAP_PIN questions can't have betting"},
			"an invalid config": {
				input:   map[string]any{"mapPin": map[string]any{"scoringMode": "POINT_RADIUS", "mapWidth": 0, "mapHeight": 800}},
				message: "mapPin.mapWidth",
			},
		} {
			t.Run(name, func(t *testing.T) {
				resp := updateQuestionResp(t, q.ID, tt.input)
				require.True(t, resp.HasErrors(), "expected the update to be rejected")
				assert.Contains(t, resp.ErrorMessage(), tt.message)
			})
		}

		t.Run("mapPin on another type", func(t *testing.T) {
			resp := addQuestionResp(t, quizID, map[string]any{"questionType": "FREE_TEXT", "questionOrder": 1})
			require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())
			var result struct {
				AddQuizQuestion struct {
					Typename string `json:"__typename"`
				} `json:"addQuizQuestion"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			require.Equal(t, "FreeTextQuestion", result.AddQuizQuestion.Typename)

			questions, err := dbMgr.DB.Queries.GetQuizQuestionsByQuizID(ctx, quizID)
			require.NoError(t, err)
			require.Len(t, questions, 2)
			freeTextID := questions[1].ID

			resp = updateQuestionResp(t, freeTextID, map[string]any{"mapPin": pointRadiusMapPin()})
			require.True(t, resp.HasErrors(), "expected the update to be rejected")
			assert.Contains(t, resp.ErrorMessage(), "mapPin: is only used by MAP_PIN questions")
		})

		// The rejected updates left the question as it was.
		row, err := dbMgr.DB.Queries.GetQuizQuestionByID(ctx, q.ID)
		require.NoError(t, err)
		assert.Equal(t, ptr(int32(100)), row.Points)
		assert.False(t, row.BettingEnabled)
		assert.Equal(t, ptr(int32(1000)), row.MapWidth)
	})

	t.Run("scoring can't change once the question is answered", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin edit lock")
		q := addMapPinQuestion(t, quizID, pointRadiusMapPin())
		_, submissionID := startSession(t, quizID)

		// Stored directly: answering MAP_PIN through GraphQL is a later step.
		_, err := dbMgr.DB.Queries.CreateQuizResponse(ctx, sqlc.CreateQuizResponseParams{
			ID:           ulid.NewQuizResponseID(),
			Submissionid: submissionID,
			Questionid:   q.ID,
			Jsonresponse: []byte(`{"x": 510, "y": 300}`),
			Pointsearned: ptr(int32(100)),
		})
		require.NoError(t, err)

		changed := pointRadiusMapPin()
		changed["zeroPointsRadius"] = 300

		for name, tt := range map[string]struct {
			input   map[string]any
			message string
		}{
			"a changed config": {input: map[string]any{"mapPin": changed}, message: "mapPin: can't be changed after users have answered"},
			"a mode switch":    {input: map[string]any{"mapPin": trianglePinArea()}, message: "mapPin: can't be changed after users have answered"},
			"changed points":   {input: map[string]any{"points": 50}, message: "points: can't be changed after users have answered"},
		} {
			t.Run("rejects "+name, func(t *testing.T) {
				resp := updateQuestionResp(t, q.ID, tt.input)
				require.True(t, resp.HasErrors(), "expected the update to be rejected")
				assert.Contains(t, resp.ErrorMessage(), tt.message)
			})
		}

		t.Run("allows resending the same config and points with a new text", func(t *testing.T) {
			resp := updateQuestionResp(t, q.ID, map[string]any{
				"questionText": "Where exactly was this taken?",
				"points":       100,
				"mapPin":       pointRadiusMapPin(),
			})
			require.False(t, resp.HasErrors(), "an unchanged config must not lock: %s", resp.ErrorMessage())
		})

		row, err := dbMgr.DB.Queries.GetQuizQuestionByID(ctx, q.ID)
		require.NoError(t, err)
		assert.Equal(t, ptr(int32(100)), row.Points)
		assert.Equal(t, ptr(200.0), row.MapZeroPointsRadius)
		assert.Equal(t, "Where exactly was this taken?", row.QuestionText)
	})

	t.Run("correct geometry visibility", func(t *testing.T) {
		for _, mode := range []struct {
			name   string
			mapPin func() map[string]any
			hidden func(*testing.T, mapPinQuestion)
			shown  func(*testing.T, mapPinQuestion)
		}{
			{
				name:   "POINT_RADIUS",
				mapPin: pointRadiusMapPin,
				hidden: func(t *testing.T, q mapPinQuestion) { assert.Nil(t, q.CorrectPosition) },
				shown:  func(t *testing.T, q mapPinQuestion) { assert.Equal(t, &mapPoint{500, 300}, q.CorrectPosition) },
			},
			{
				name:   "POLYGON",
				mapPin: trianglePinArea,
				hidden: func(t *testing.T, q mapPinQuestion) { assert.Nil(t, q.CorrectArea) },
				shown:  func(t *testing.T, q mapPinQuestion) { assert.Equal(t, triangleArea, q.CorrectArea) },
			},
		} {
			t.Run(mode.name, func(t *testing.T) {
				quizID := createQuiz(t, "Map pin visibility "+mode.name)
				addMapPinQuestion(t, quizID, mode.mapPin())

				// The public parts of the question stay readable whatever the answer's visibility.
				assertPublic := func(t *testing.T, q mapPinQuestion) {
					t.Helper()
					assert.Equal(t, "MapPinQuestion", q.Typename)
					assert.Equal(t, mode.name, q.ScoringMode)
					assert.Equal(t, 1000, q.MapWidth)
					assert.Equal(t, 800, q.MapHeight)
				}

				t.Run("hidden from a user before any session exists", func(t *testing.T) {
					q := readQuestion(t, quizID, userToken)
					assertPublic(t, q)
					mode.hidden(t, q)
				})

				sessionID, _ := startSession(t, quizID)

				t.Run("hidden while the session is OPEN", func(t *testing.T) {
					mode.hidden(t, readQuestion(t, quizID, userToken))
				})
				t.Run("shown to a manager while the session is OPEN", func(t *testing.T) {
					mode.shown(t, readQuestion(t, quizID, adminToken))
				})

				lockSession(t, sessionID)
				t.Run("hidden while the session is LOCKED", func(t *testing.T) {
					mode.hidden(t, readQuestion(t, quizID, userToken))
				})

				finishSession(t, sessionID)
				t.Run("shown once the session is FINISHED", func(t *testing.T) {
					q := readQuestion(t, quizID, userToken)
					assertPublic(t, q)
					mode.shown(t, q)
				})
				t.Run("still hidden from a user without access to any session", func(t *testing.T) {
					q := readQuestion(t, quizID, otherUserToken)
					assertPublic(t, q)
					mode.hidden(t, q)
				})
			})
		}
	})
}
