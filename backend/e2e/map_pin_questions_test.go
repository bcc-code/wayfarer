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

const mapPinResponseFields = `
	__typename
	... on MapPinResponse {
		id
		submittedPosition { x y }
		pointsEarned
		distance
		isCorrect
	}
`

type mapPinResponse struct {
	Typename          string    `json:"__typename"`
	ID                string    `json:"id"`
	SubmittedPosition *mapPoint `json:"submittedPosition"`
	PointsEarned      *int      `json:"pointsEarned"`
	Distance          *float64  `json:"distance"`
	IsCorrect         *bool     `json:"isCorrect"`
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

	// addMapPinQuestionAt adds a MAP_PIN question at the given position in the
	// quiz and returns its ID.
	addMapPinQuestionAt := func(t *testing.T, quizID string, order, points int, mapPin map[string]any) string {
		t.Helper()
		resp := addQuestionResp(t, quizID, map[string]any{
			"questionType": "MAP_PIN", "questionOrder": order, "points": points, "mapPin": mapPin,
		})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())
		var result struct {
			AddQuizQuestion mapPinQuestion `json:"addQuizQuestion"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.AddQuizQuestion.ID
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
	submitAnswerResp := func(t *testing.T, submissionID string, input map[string]any) *testutil.GraphQLResponse {
		t.Helper()
		return client.WithAuth(userToken).MustExecute(t, `
			mutation SubmitAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
				submitQuizAnswer(submissionId: $submissionId, input: $input) {`+mapPinResponseFields+`}
			}
		`, map[string]any{"submissionId": submissionID, "input": input})
	}

	// submitPin answers a question with a pin and returns the response.
	submitPin := func(t *testing.T, submissionID, questionID string, x, y float64) mapPinResponse {
		t.Helper()
		resp := submitAnswerResp(t, submissionID, map[string]any{
			"questionId":  questionID,
			"mapPosition": map[string]any{"x": x, "y": y},
		})
		require.False(t, resp.HasErrors(), "failed to submit pin: %s", resp.ErrorMessage())
		var result struct {
			SubmitQuizAnswer mapPinResponse `json:"submitQuizAnswer"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		return result.SubmitQuizAnswer
	}

	updateAnswerResp := func(t *testing.T, responseID string, input map[string]any) *testutil.GraphQLResponse {
		t.Helper()
		return client.WithAuth(userToken).MustExecute(t, `
			mutation UpdateAnswer($responseId: ID!, $input: UpdateQuizAnswerInput!) {
				updateQuizAnswer(responseId: $responseId, input: $input) {`+mapPinResponseFields+`}
			}
		`, map[string]any{"responseId": responseID, "input": input})
	}

	// readResponses returns the submission's responses by question ID, as seen
	// by the holder of token, through quizSubmission (the responses dataloader).
	readResponses := func(t *testing.T, submissionID, token string) map[string]mapPinResponse {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, `
			query GetSubmission($id: ID!) {
				quizSubmission(id: $id) { responses {`+mapPinResponseFields+` question { id } } }
			}
		`, map[string]any{"id": submissionID})
		require.False(t, resp.HasErrors(), "failed to read submission: %s", resp.ErrorMessage())
		var result struct {
			QuizSubmission struct {
				Responses []struct {
					mapPinResponse
					Question struct{ ID string } `json:"question"`
				} `json:"responses"`
			} `json:"quizSubmission"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		byQuestion := make(map[string]mapPinResponse, len(result.QuizSubmission.Responses))
		for _, r := range result.QuizSubmission.Responses {
			byQuestion[r.Question.ID] = r.mapPinResponse
		}
		return byQuestion
	}

	storedPoints := func(t *testing.T, responseID string) *int32 {
		t.Helper()
		row, err := dbMgr.DB.Queries.GetQuizResponseByID(ctx, responseID)
		require.NoError(t, err)
		return row.PointsEarned
	}

	t.Run("answering scores the pin", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin answers")
		near := addMapPinQuestionAt(t, quizID, 0, 100, pointRadiusMapPin())
		halfway := addMapPinQuestionAt(t, quizID, 1, 100, pointRadiusMapPin())
		far := addMapPinQuestionAt(t, quizID, 2, 100, pointRadiusMapPin())
		inside := addMapPinQuestionAt(t, quizID, 3, 30, trianglePinArea())
		onEdge := addMapPinQuestionAt(t, quizID, 4, 30, trianglePinArea())
		outside := addMapPinQuestionAt(t, quizID, 5, 30, trianglePinArea())
		sessionID, submissionID := startSession(t, quizID)

		pins := []struct {
			name       string
			questionID string
			x, y       float64
			points     int32
			distance   *float64
		}{
			{name: "POINT_RADIUS within the full radius", questionID: near, x: 530, y: 300, points: 100, distance: ptr(30.0)},
			{name: "POINT_RADIUS halfway to the zero radius", questionID: halfway, x: 500, y: 420, points: 50, distance: ptr(120.0)},
			{name: "POINT_RADIUS beyond the zero radius", questionID: far, x: 900, y: 700, points: 0, distance: ptr(565.685424949238)},
			{name: "POLYGON inside", questionID: inside, x: 200, y: 150, points: 30},
			{name: "POLYGON on an edge", questionID: onEdge, x: 200, y: 100, points: 30},
			{name: "POLYGON outside", questionID: outside, x: 500, y: 500, points: 0},
		}

		for _, p := range pins {
			t.Run("submit: "+p.name, func(t *testing.T) {
				r := submitPin(t, submissionID, p.questionID, p.x, p.y)
				assert.Equal(t, "MapPinResponse", r.Typename)
				assert.Equal(t, &mapPoint{p.x, p.y}, r.SubmittedPosition)
				assert.Nil(t, r.IsCorrect, "MAP_PIN is never graded")
				assert.Nil(t, r.PointsEarned, "points are hidden while the session is OPEN")
				assert.Nil(t, r.Distance, "distance is hidden while the session is OPEN")

				assert.Equal(t, ptr(p.points), storedPoints(t, r.ID), "stored points_earned")
			})
		}

		t.Run("hidden from everyone while the session is OPEN", func(t *testing.T) {
			for name, token := range map[string]string{"respondent": userToken, "manager": adminToken} {
				responses := readResponses(t, submissionID, token)
				require.Len(t, responses, len(pins), name)
				for _, p := range pins {
					r := responses[p.questionID]
					assert.Equal(t, "MapPinResponse", r.Typename, name)
					assert.Equal(t, &mapPoint{p.x, p.y}, r.SubmittedPosition, name)
					assert.Nil(t, r.PointsEarned, "%s: %s", name, p.name)
					assert.Nil(t, r.Distance, "%s: %s", name, p.name)
				}
			}
		})

		t.Run("another user can't read the pins", func(t *testing.T) {
			resp := client.WithAuth(otherUserToken).MustExecute(t, `
				query GetSubmission($id: ID!) {
					quizSubmission(id: $id) { responses { ... on MapPinResponse { submittedPosition { x y } } } }
				}
			`, map[string]any{"id": submissionID})
			require.True(t, resp.HasErrors())
			assert.Contains(t, resp.ErrorMessage(), "unauthorized")
		})

		lockSession(t, sessionID)
		t.Run("shown once the session is LOCKED", func(t *testing.T) {
			responses := readResponses(t, submissionID, userToken)
			for _, p := range pins {
				r := responses[p.questionID]
				assert.Equal(t, ptr(int(p.points)), r.PointsEarned, p.name)
				if p.distance == nil {
					assert.Nil(t, r.Distance, "%s: POLYGON has no distance", p.name)
				} else if assert.NotNil(t, r.Distance, p.name) {
					assert.InDelta(t, *p.distance, *r.Distance, 1e-9, p.name)
				}
			}
		})
	})

	t.Run("submit rejects", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin answer rules")
		questionID := addMapPinQuestionAt(t, quizID, 0, 100, pointRadiusMapPin())
		resp := addQuestionResp(t, quizID, map[string]any{"questionType": "FREE_TEXT", "questionOrder": 1})
		require.False(t, resp.HasErrors(), "failed to add question: %s", resp.ErrorMessage())
		questions, err := dbMgr.DB.Queries.GetQuizQuestionsByQuizID(ctx, quizID)
		require.NoError(t, err)
		require.Len(t, questions, 2)
		freeTextID := questions[1].ID
		_, submissionID := startSession(t, quizID)

		for _, tt := range []struct {
			name    string
			input   map[string]any
			message string
		}{
			{name: "no mapPosition", input: map[string]any{"questionId": questionID}, message: "mapPosition: is required"},
			{name: "a pin right of the map", input: map[string]any{"questionId": questionID, "mapPosition": map[string]any{"x": 1000.5, "y": 10}}, message: "mapPosition"},
			{name: "a pin above the map", input: map[string]any{"questionId": questionID, "mapPosition": map[string]any{"x": 10, "y": -1}}, message: "mapPosition"},
			{name: "jsonResponse on MAP_PIN", input: map[string]any{"questionId": questionID, "jsonResponse": `{"x": 500, "y": 300}`}, message: "cannot set a JSON answer on a MAP_PIN question"},
			{name: "a bet", input: map[string]any{"questionId": questionID, "mapPosition": map[string]any{"x": 500, "y": 300}, "betAmount": 5}, message: "betAmount"},
			{name: "mapPosition on another type", input: map[string]any{"questionId": freeTextID, "mapPosition": map[string]any{"x": 1, "y": 1}}, message: "cannot set a MAP_PIN answer on a FREE_TEXT question"},
		} {
			t.Run(tt.name, func(t *testing.T) {
				resp := submitAnswerResp(t, submissionID, tt.input)
				require.True(t, resp.HasErrors(), "expected the answer to be rejected")
				assert.Contains(t, resp.ErrorMessage(), tt.message)
			})
		}

		responses, err := dbMgr.DB.Queries.GetQuizResponsesBySubmissionID(ctx, submissionID)
		require.NoError(t, err)
		assert.Empty(t, responses, "rejected answers must not be stored")
	})

	t.Run("updating a pin rescores it", func(t *testing.T) {
		quizID := createQuiz(t, "Map pin answer updates")
		pointQuestion := addMapPinQuestionAt(t, quizID, 0, 100, pointRadiusMapPin())
		areaQuestion := addMapPinQuestionAt(t, quizID, 1, 30, trianglePinArea())
		_, submissionID := startSession(t, quizID)

		pointResponse := submitPin(t, submissionID, pointQuestion, 500, 300)
		areaResponse := submitPin(t, submissionID, areaQuestion, 200, 150)
		require.Equal(t, ptr(int32(100)), storedPoints(t, pointResponse.ID))
		require.Equal(t, ptr(int32(30)), storedPoints(t, areaResponse.ID))

		update := func(t *testing.T, responseID string, x, y float64) {
			t.Helper()
			resp := updateAnswerResp(t, responseID, map[string]any{"mapPosition": map[string]any{"x": x, "y": y}})
			require.False(t, resp.HasErrors(), "failed to update pin: %s", resp.ErrorMessage())
			var result struct {
				UpdateQuizAnswer mapPinResponse `json:"updateQuizAnswer"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			assert.Equal(t, "MapPinResponse", result.UpdateQuizAnswer.Typename)
			assert.Equal(t, &mapPoint{x, y}, result.UpdateQuizAnswer.SubmittedPosition)
		}

		t.Run("POINT_RADIUS from the target to beyond the zero radius", func(t *testing.T) {
			update(t, pointResponse.ID, 900, 300)
			assert.Equal(t, ptr(int32(0)), storedPoints(t, pointResponse.ID), "the old 100 points must not be kept")
			update(t, pointResponse.ID, 620, 300)
			assert.Equal(t, ptr(int32(50)), storedPoints(t, pointResponse.ID))
		})

		t.Run("POLYGON from inside to outside", func(t *testing.T) {
			update(t, areaResponse.ID, 500, 500)
			assert.Equal(t, ptr(int32(0)), storedPoints(t, areaResponse.ID), "the old 30 points must not be kept")
			update(t, areaResponse.ID, 100, 100)
			assert.Equal(t, ptr(int32(30)), storedPoints(t, areaResponse.ID), "a vertex counts as inside")
		})

		t.Run("rejects", func(t *testing.T) {
			for name, tt := range map[string]struct {
				input   map[string]any
				message string
			}{
				"a pin outside the map": {input: map[string]any{"mapPosition": map[string]any{"x": 10, "y": 801}}, message: "mapPosition"},
				"jsonResponse":          {input: map[string]any{"jsonResponse": `{"x": 500, "y": 300}`}, message: "cannot set a JSON answer on a MAP_PIN question"},
				"a bet":                 {input: map[string]any{"betAmount": 5}, message: "betAmount"},
			} {
				t.Run(name, func(t *testing.T) {
					resp := updateAnswerResp(t, pointResponse.ID, tt.input)
					require.True(t, resp.HasErrors(), "expected the update to be rejected")
					assert.Contains(t, resp.ErrorMessage(), tt.message)
				})
			}
			assert.Equal(t, ptr(int32(50)), storedPoints(t, pointResponse.ID), "rejected updates leave the points alone")
			row, err := dbMgr.DB.Queries.GetQuizResponseByID(ctx, pointResponse.ID)
			require.NoError(t, err)
			assert.JSONEq(t, `{"x": 620, "y": 300}`, string(row.JsonResponse))
		})
	})

	t.Run("M2M submissions score pins", func(t *testing.T) {
		m2mToken, err := testutil.GenerateM2MToken()
		require.NoError(t, err)

		quizID := createQuiz(t, "Map pin M2M")
		pointQuestion := addMapPinQuestionAt(t, quizID, 0, 100, pointRadiusMapPin())
		areaQuestion := addMapPinQuestionAt(t, quizID, 1, 30, trianglePinArea())

		submit := func(t *testing.T, responses []map[string]any) *testutil.GraphQLResponse {
			t.Helper()
			return client.WithAuth(m2mToken).MustExecute(t, `
				mutation Submit($quizId: ID!, $userId: ID!, $responses: [SubmitQuizAnswerInput!]!, $completedAt: DateTime) {
					createQuizSubmission(quizId: $quizId, userId: $userId, responses: $responses, completedAt: $completedAt) {
						id score maxScore
					}
				}
			`, map[string]any{
				"quizId":      quizID,
				"userId":      userID,
				"responses":   responses,
				"completedAt": time.Now().Format(time.RFC3339),
			})
		}

		t.Run("rejects jsonResponse on MAP_PIN", func(t *testing.T) {
			resp := submit(t, []map[string]any{{"questionId": pointQuestion, "jsonResponse": `{"x": 500, "y": 300}`}})
			require.True(t, resp.HasErrors())
			assert.Contains(t, resp.ErrorMessage(), "cannot set a JSON answer on a MAP_PIN question")
		})

		resp := submit(t, []map[string]any{
			{"questionId": pointQuestion, "mapPosition": map[string]any{"x": 620, "y": 300}},
			{"questionId": areaQuestion, "mapPosition": map[string]any{"x": 200, "y": 150}},
		})
		require.False(t, resp.HasErrors(), "failed to create submission: %s", resp.ErrorMessage())
		var result struct {
			CreateQuizSubmission struct {
				ID       string `json:"id"`
				Score    *int   `json:"score"`
				MaxScore *int   `json:"maxScore"`
			} `json:"createQuizSubmission"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		assert.Equal(t, ptr(50+30), result.CreateQuizSubmission.Score)
		assert.Equal(t, ptr(100+30), result.CreateQuizSubmission.MaxScore, "MAP_PIN points count towards the max score")

		responses, err := dbMgr.DB.Queries.GetQuizResponsesBySubmissionID(ctx, result.CreateQuizSubmission.ID)
		require.NoError(t, err)
		require.Len(t, responses, 2)
		points := map[string]*int32{}
		for _, r := range responses {
			points[r.QuestionID] = r.PointsEarned
			assert.Nil(t, r.IsCorrect)
		}
		assert.Equal(t, ptr(int32(50)), points[pointQuestion])
		assert.Equal(t, ptr(int32(30)), points[areaQuestion])

		t.Run("results are visible without a session", func(t *testing.T) {
			read := readResponses(t, result.CreateQuizSubmission.ID, userToken)
			require.Len(t, read, 2)

			pin := read[pointQuestion]
			assert.Equal(t, "MapPinResponse", pin.Typename)
			assert.Equal(t, &mapPoint{620, 300}, pin.SubmittedPosition)
			assert.Equal(t, ptr(50), pin.PointsEarned)
			if assert.NotNil(t, pin.Distance) {
				assert.InDelta(t, 120.0, *pin.Distance, 1e-9)
			}

			area := read[areaQuestion]
			assert.Equal(t, ptr(30), area.PointsEarned)
			assert.Nil(t, area.Distance, "POLYGON has no distance")
		})

		t.Run("recordBetResult rejects a MAP_PIN response", func(t *testing.T) {
			var responseID string
			for _, r := range responses {
				if r.QuestionID == pointQuestion {
					responseID = r.ID
				}
			}

			single := client.WithAuth(m2mToken).MustExecute(t, `
				mutation Record($input: RecordBetResultInput!) { recordBetResult(input: $input) { id } }
			`, map[string]any{"input": map[string]any{"responseId": responseID, "pointsEarned": 999}})
			require.True(t, single.HasErrors())
			assert.Contains(t, single.ErrorMessage(), "MAP_PIN")

			batch := client.WithAuth(m2mToken).MustExecute(t, `
				mutation Record($inputs: [RecordBetResultInput!]!) { recordBetResults(inputs: $inputs) { id } }
			`, map[string]any{"inputs": []map[string]any{{"responseId": responseID, "pointsEarned": 999}}})
			require.True(t, batch.HasErrors())
			assert.Contains(t, batch.ErrorMessage(), "MAP_PIN")

			assert.Equal(t, ptr(int32(50)), storedPoints(t, responseID), "the pin's points must not be overwritten")
		})
	})

	t.Run("finalize and results", func(t *testing.T) {
		m2mToken, err := testutil.GenerateM2MToken()
		require.NoError(t, err)

		quizID := createQuiz(t, "Map pin results")
		resp := client.WithAuth(adminToken).MustExecute(t, `
			mutation UpdateQuiz($id: ID!, $input: UpdateQuizInput!) { updateQuiz(id: $id, input: $input) { id } }
		`, map[string]any{"id": quizID, "input": map[string]any{"completionPoints": 10}})
		require.False(t, resp.HasErrors(), "failed to set completion points: %s", resp.ErrorMessage())

		pointQuestion := addMapPinQuestionAt(t, quizID, 0, 100, pointRadiusMapPin())
		areaQuestion := addMapPinQuestionAt(t, quizID, 1, 30, trianglePinArea())

		// 50 + 30 points. Locked first: finalizing while OPEN is the known gap.
		sessionID, submissionID := startSession(t, quizID)
		submitPin(t, submissionID, pointQuestion, 620, 300)
		submitPin(t, submissionID, areaQuestion, 200, 150)
		lockSession(t, sessionID)

		resp = client.WithAuth(userToken).MustExecute(t, `
			mutation Finalize($id: ID!) { finalizeQuiz(submissionId: $id) { score maxScore pointsAwarded } }
		`, map[string]any{"id": submissionID})
		require.False(t, resp.HasErrors(), "failed to finalize: %s", resp.ErrorMessage())
		var finalized struct {
			FinalizeQuiz struct {
				Score         *int `json:"score"`
				MaxScore      *int `json:"maxScore"`
				PointsAwarded *int `json:"pointsAwarded"`
			} `json:"finalizeQuiz"`
		}
		require.NoError(t, resp.UnmarshalData(&finalized))

		t.Run("finalize sums the pins' points", func(t *testing.T) {
			assert.Equal(t, ptr(50+30), finalized.FinalizeQuiz.Score)
			assert.Equal(t, ptr(100+30), finalized.FinalizeQuiz.MaxScore)
			assert.Equal(t, ptr(10+50+30), finalized.FinalizeQuiz.PointsAwarded, "completion points plus the pins")

			var points, entries int
			require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx, `
				SELECT COALESCE(SUM(points), 0), COUNT(*) FROM score_journal
				WHERE user_id = $1 AND source_type = 'QUIZ' AND source_id = $2
			`, userID, quizID).Scan(&points, &entries))
			assert.Equal(t, 1, entries)
			assert.Equal(t, 10+50+30, points, "the QUIZ journal entry carries the pins' points")
		})

		// M2M: 40.5 px out rounds up to full points; the second pin is outside.
		resp = client.WithAuth(m2mToken).MustExecute(t, `
			mutation Submit($quizId: ID!, $userId: ID!, $responses: [SubmitQuizAnswerInput!]!, $completedAt: DateTime) {
				createQuizSubmission(quizId: $quizId, userId: $userId, responses: $responses, completedAt: $completedAt) { id }
			}
		`, map[string]any{
			"quizId": quizID,
			"userId": otherUserID,
			"responses": []map[string]any{
				{"questionId": pointQuestion, "mapPosition": map[string]any{"x": 540.5, "y": 300}},
				{"questionId": areaQuestion, "mapPosition": map[string]any{"x": 500, "y": 500}},
			},
			"completedAt": time.Now().Format(time.RFC3339),
		})
		require.False(t, resp.HasErrors(), "failed to create submission: %s", resp.ErrorMessage())

		type pinResult struct {
			Position     mapPoint `json:"position"`
			PointsEarned *int     `json:"pointsEarned"`
			Distance     *float64 `json:"distance"`
		}
		type questionResults struct {
			Typename        string         `json:"__typename"`
			Question        mapPinQuestion `json:"question"`
			ResponseCount   int            `json:"responseCount"`
			FullPointsCount int            `json:"fullPointsCount"`
			AveragePoints   *float64       `json:"averagePoints"`
			AverageDistance *float64       `json:"averageDistance"`
			Pins            []pinResult    `json:"pins"`
		}
		readResults := func(t *testing.T, token string) (*testutil.GraphQLResponse, []questionResults) {
			t.Helper()
			resp := client.WithAuth(token).MustExecute(t, `
				query Results($quizId: ID!) {
					quizResults(quizId: $quizId) {
						questions {
							__typename
							responseCount
							question {`+mapPinQuestionFields+`}
							... on MapPinQuestionResults {
								fullPointsCount
								averagePoints
								averageDistance
								pins { position { x y } pointsEarned distance }
							}
						}
					}
				}
			`, map[string]any{"quizId": quizID})
			if resp.HasErrors() {
				return resp, nil
			}
			var result struct {
				QuizResults struct {
					Questions []questionResults `json:"questions"`
				} `json:"quizResults"`
			}
			require.NoError(t, resp.UnmarshalData(&result))
			return resp, result.QuizResults.Questions
		}

		_, results := readResults(t, adminToken)
		require.Len(t, results, 2)

		t.Run("POINT_RADIUS results", func(t *testing.T) {
			r := results[0]
			assert.Equal(t, "MapPinQuestionResults", r.Typename)
			assert.Equal(t, "MapPinQuestion", r.Question.Typename)
			assert.Equal(t, pointQuestion, r.Question.ID)
			assert.Equal(t, &mapPoint{500, 300}, r.Question.CorrectPosition, "the manager sees the target with the results")
			assert.Equal(t, 2, r.ResponseCount)
			assert.Equal(t, 1, r.FullPointsCount, "the pin 40.5 px out rounded up to full points")
			if assert.NotNil(t, r.AveragePoints) {
				assert.InDelta(t, (50+100)/2.0, *r.AveragePoints, 1e-9)
			}
			if assert.NotNil(t, r.AverageDistance) {
				assert.InDelta(t, (120+40.5)/2, *r.AverageDistance, 1e-9)
			}
			assert.Equal(t, []pinResult{
				{Position: mapPoint{620, 300}, PointsEarned: ptr(50), Distance: ptr(120.0)},
				{Position: mapPoint{540.5, 300}, PointsEarned: ptr(100), Distance: ptr(40.5)},
			}, r.Pins, "in the order they were answered")
		})

		t.Run("POLYGON results", func(t *testing.T) {
			r := results[1]
			assert.Equal(t, "MapPinQuestionResults", r.Typename)
			assert.Equal(t, areaQuestion, r.Question.ID)
			assert.Equal(t, triangleArea, r.Question.CorrectArea)
			assert.Equal(t, 2, r.ResponseCount)
			assert.Equal(t, 1, r.FullPointsCount)
			if assert.NotNil(t, r.AveragePoints) {
				assert.InDelta(t, 15.0, *r.AveragePoints, 1e-9)
			}
			assert.Nil(t, r.AverageDistance, "POLYGON has no distance")
			assert.Equal(t, []pinResult{
				{Position: mapPoint{200, 150}, PointsEarned: ptr(30)},
				{Position: mapPoint{500, 500}, PointsEarned: ptr(0)},
			}, r.Pins, "in the order they were answered")
		})

		t.Run("results are for managers only", func(t *testing.T) {
			resp, _ := readResults(t, userToken)
			require.True(t, resp.HasErrors())
			assert.Contains(t, resp.ErrorMessage(), "unauthorized")
		})
	})
}
