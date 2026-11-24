package e2e

import (
	"context"
	"errors"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMapPinQuestionSchema checks migration 00119 directly against Postgres:
// quiz_questions_map_pin_check accepts exactly the map configuration each
// question type and scoring mode needs, and UpdateQuizQuestion replaces the
// map config as a whole. The GraphQL layer validates before it gets here; this
// is the last line of defence.
func TestMapPinQuestionSchema(t *testing.T) {
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

	quizID := createMapPinSchemaQuiz(t, client.WithAuth(adminToken), data.ProjectIDs[0])
	queries := dbMgr.DB.Queries

	order := int32(0)
	create := func(modify func(*sqlc.CreateQuizQuestionParams)) (*sqlc.CreateQuizQuestionRow, error) {
		order++
		params := pointRadiusQuestionParams(quizID, order)
		modify(&params)
		return queries.CreateQuizQuestion(ctx, params)
	}

	t.Run("create accepts", func(t *testing.T) {
		tests := []struct {
			name   string
			modify func(*sqlc.CreateQuizQuestionParams)
		}{
			{name: "POINT_RADIUS", modify: func(p *sqlc.CreateQuizQuestionParams) {}},
			{name: "POINT_RADIUS with full radius 0 and the target on the map's corner", modify: func(p *sqlc.CreateQuizQuestionParams) {
				p.Mapfullpointsradius = ptr(int32(0))
				p.Maptarget = pgPoint(1000, 800)
			}},
			{name: "POLYGON with 3 vertices", modify: toPolygon(t, 3)},
			{name: "POLYGON with 100 vertices", modify: toPolygon(t, 100)},
			{name: "PREDEFINED without map columns", modify: toPredefined},
			{name: "PREDEFINED with betting, as before", modify: func(p *sqlc.CreateQuizQuestionParams) {
				toPredefined(p)
				p.Points = nil
				p.Bettingenabled = ptr(true)
			}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := create(tt.modify)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("create rejects", func(t *testing.T) {
		// Built here, not inside the cases: require must run on this test's goroutine.
		toTriangle := toPolygon(t, 3)
		triangle := polygonPoints(3)

		tests := []struct {
			name   string
			modify func(*sqlc.CreateQuizQuestionParams)
		}{
			{name: "map columns on a PREDEFINED question", modify: func(p *sqlc.CreateQuizQuestionParams) {
				toPredefined(p)
				p.Mapwidth = ptr(int32(1000))
			}},
			{name: "MAP_PIN without a scoring mode", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mapscoringmode = nil }},
			{name: "unknown scoring mode", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mapscoringmode = ptr("CIRCLE") }},
			{name: "missing width", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mapwidth = nil }},
			{name: "zero width", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mapwidth = ptr(int32(0)) }},
			{name: "negative height", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mapheight = ptr(int32(-1)) }},
			{name: "missing points", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Points = nil }},
			{name: "zero points", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Points = ptr(int32(0)) }},
			{name: "betting enabled", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Bettingenabled = ptr(true) }},
			{name: "POINT_RADIUS without a target", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Maptarget = pgtype.Point{} }},
			{name: "POINT_RADIUS target right of the map", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Maptarget = pgPoint(1001, 300) }},
			{name: "POINT_RADIUS target above the map", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Maptarget = pgPoint(500, -1) }},
			{name: "POINT_RADIUS target not on a whole pixel (x)", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Maptarget = pgPoint(500.5, 300) }},
			{name: "POINT_RADIUS target not on a whole pixel (y)", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Maptarget = pgPoint(500, 300.25) }},
			{name: "POINT_RADIUS without a zero radius", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mapzeropointsradius = nil }},
			{name: "POINT_RADIUS negative full radius", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mapfullpointsradius = ptr(int32(-1)) }},
			{name: "POINT_RADIUS zero radius equal to full radius", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mapzeropointsradius = ptr(int32(40)) }},
			{name: "POINT_RADIUS with a polygon", modify: func(p *sqlc.CreateQuizQuestionParams) { p.Mappolygon = triangle }},
			{name: "POLYGON without a polygon", modify: func(p *sqlc.CreateQuizQuestionParams) {
				toTriangle(p)
				p.Mappolygon = nil
			}},
			{name: "POLYGON with no vertices", modify: func(p *sqlc.CreateQuizQuestionParams) {
				toTriangle(p)
				p.Mappolygon = []pgtype.Point{}
			}},
			{name: "POLYGON with 2 vertices", modify: toPolygon(t, 2)},
			{name: "POLYGON with 101 vertices", modify: toPolygon(t, 101)},
			{name: "POLYGON with a target", modify: func(p *sqlc.CreateQuizQuestionParams) {
				toTriangle(p)
				p.Maptarget = pgPoint(500, 300)
			}},
			{name: "POLYGON with a full radius", modify: func(p *sqlc.CreateQuizQuestionParams) {
				toTriangle(p)
				p.Mapfullpointsradius = ptr(int32(40))
			}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := create(tt.modify)
				assertCheckViolation(t, err, "quiz_questions_map_pin_check")
			})
		}
	})

	t.Run("update", func(t *testing.T) {
		pointRadiusQuestion := func(t *testing.T) string {
			row, err := create(func(p *sqlc.CreateQuizQuestionParams) {})
			require.NoError(t, err)
			return row.ID
		}

		t.Run("without setmappin leaves the map config alone", func(t *testing.T) {
			id := pointRadiusQuestion(t)
			row, err := queries.UpdateQuizQuestion(ctx, sqlc.UpdateQuizQuestionParams{ID: id, Questiontext: ptr("Where was this taken?")})
			require.NoError(t, err)

			assert.Equal(t, "Where was this taken?", row.QuestionText)
			assert.Equal(t, ptr("POINT_RADIUS"), row.MapScoringMode)
			assert.Equal(t, pgPoint(500, 300), row.MapTarget)
			assert.Equal(t, ptr(int32(40)), row.MapFullPointsRadius)
		})

		t.Run("switching to POLYGON clears the POINT_RADIUS columns", func(t *testing.T) {
			id := pointRadiusQuestion(t)
			polygon := polygonPoints(4)
			row, err := queries.UpdateQuizQuestion(ctx, sqlc.UpdateQuizQuestionParams{
				ID:             id,
				Setmappin:      true,
				Mapscoringmode: ptr("POLYGON"),
				Mapwidth:       ptr(int32(1200)),
				Mapheight:      ptr(int32(900)),
				Mappolygon:     polygon,
			})
			require.NoError(t, err)

			assert.Equal(t, ptr("POLYGON"), row.MapScoringMode)
			assert.Equal(t, ptr(int32(1200)), row.MapWidth)
			assert.Equal(t, polygon, row.MapPolygon)
			assert.False(t, row.MapTarget.Valid)
			assert.Nil(t, row.MapFullPointsRadius)
			assert.Nil(t, row.MapZeroPointsRadius)
		})

		t.Run("switching back to POINT_RADIUS clears the polygon", func(t *testing.T) {
			row, err := create(toPolygon(t, 3))
			require.NoError(t, err)

			updated, err := queries.UpdateQuizQuestion(ctx, sqlc.UpdateQuizQuestionParams{
				ID:                  row.ID,
				Setmappin:           true,
				Mapscoringmode:      ptr("POINT_RADIUS"),
				Mapwidth:            ptr(int32(1000)),
				Mapheight:           ptr(int32(800)),
				Maptarget:           pgPoint(100, 200),
				Mapfullpointsradius: ptr(int32(10)),
				Mapzeropointsradius: ptr(int32(50)),
			})
			require.NoError(t, err)

			assert.Equal(t, ptr("POINT_RADIUS"), updated.MapScoringMode)
			assert.Equal(t, pgPoint(100, 200), updated.MapTarget)
			assert.Nil(t, updated.MapPolygon)
		})

		t.Run("an incomplete config is rejected, not merged with the stored one", func(t *testing.T) {
			id := pointRadiusQuestion(t)
			_, err := queries.UpdateQuizQuestion(ctx, sqlc.UpdateQuizQuestionParams{
				ID:             id,
				Setmappin:      true,
				Mapscoringmode: ptr("POINT_RADIUS"),
				Mapwidth:       ptr(int32(1000)),
				Mapheight:      ptr(int32(800)),
			})
			assertCheckViolation(t, err, "quiz_questions_map_pin_check")

			stored, err := queries.GetQuizQuestionByID(ctx, id)
			require.NoError(t, err)
			assert.Equal(t, pgPoint(500, 300), stored.MapTarget, "the rejected update must leave the row as it was")
		})

		t.Run("enabling betting on a MAP_PIN question is rejected", func(t *testing.T) {
			id := pointRadiusQuestion(t)
			_, err := queries.UpdateQuizQuestion(ctx, sqlc.UpdateQuizQuestionParams{ID: id, Bettingenabled: ptr(true)})
			assertCheckViolation(t, err, "quiz_questions_map_pin_check")
		})

		t.Run("a PREDEFINED question updates as before", func(t *testing.T) {
			row, err := create(toPredefined)
			require.NoError(t, err)

			_, err = queries.UpdateQuizQuestion(ctx, sqlc.UpdateQuizQuestionParams{ID: row.ID, Questiontext: ptr("Renamed"), Points: ptr(int32(5))})
			assert.NoError(t, err)
		})
	})
}

// pointRadiusQuestionParams is a valid POINT_RADIUS question on a 1000x800 map:
// full points within 40 px of (500, 300), none from 200 px on.
func pointRadiusQuestionParams(quizID string, order int32) sqlc.CreateQuizQuestionParams {
	return sqlc.CreateQuizQuestionParams{
		ID:                  ulid.NewQuizQuestionID(),
		Quizid:              quizID,
		Questiontype:        "MAP_PIN",
		Questiontext:        "Where in the temple was this photo taken?",
		Questionorder:       order,
		Points:              ptr(int32(100)),
		Mapscoringmode:      ptr("POINT_RADIUS"),
		Mapwidth:            ptr(int32(1000)),
		Mapheight:           ptr(int32(800)),
		Maptarget:           pgPoint(500, 300),
		Mapfullpointsradius: ptr(int32(40)),
		Mapzeropointsradius: ptr(int32(200)),
	}
}

// toPolygon turns valid POINT_RADIUS params into POLYGON params with n
// vertices. The CHECK only looks at the vertex count; geometry is Go's job.
func toPolygon(t *testing.T, n int) func(*sqlc.CreateQuizQuestionParams) {
	t.Helper()
	polygon := polygonPoints(n)
	return func(p *sqlc.CreateQuizQuestionParams) {
		p.Mapscoringmode = ptr("POLYGON")
		p.Maptarget = pgtype.Point{}
		p.Mapfullpointsradius, p.Mapzeropointsradius = nil, nil
		p.Mappolygon = polygon
	}
}

func toPredefined(p *sqlc.CreateQuizQuestionParams) {
	p.Questiontype = "PREDEFINED"
	p.Mapscoringmode, p.Mapwidth, p.Mapheight = nil, nil, nil
	p.Maptarget = pgtype.Point{}
	p.Mapfullpointsradius, p.Mapzeropointsradius = nil, nil
	p.Mappolygon = nil
}

// polygonPoints is n whole-pixel vertices. The CHECK only counts them.
func polygonPoints(n int) []pgtype.Point {
	vertices := make([]pgtype.Point, n)
	for i := range vertices {
		vertices[i] = pgPoint(float64(10+i), float64(10+(i%2)*5))
	}
	return vertices
}

// pgPoint is a non-NULL Postgres point.
func pgPoint(x, y float64) pgtype.Point {
	return pgtype.Point{P: pgtype.Vec2{X: x, Y: y}, Valid: true}
}

func ptr[T any](v T) *T { return &v }

func assertCheckViolation(t *testing.T, err error, constraint string) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !assert.True(t, errors.As(err, &pgErr), "expected a Postgres error, got %v", err) {
		return
	}
	assert.Equal(t, "23514", pgErr.Code, "expected check_violation, got %s: %s", pgErr.Code, pgErr.Message)
	assert.Equal(t, constraint, pgErr.ConstraintName)
}

func createMapPinSchemaQuiz(t *testing.T, client *testutil.GraphQLClient, projectID string) string {
	t.Helper()

	resp := client.MustExecute(t, `
		mutation CreateChallenge($projectId: ID!, $input: CreateChallengeInput!) {
			createChallenge(projectId: $projectId, input: $input) { id }
		}
	`, map[string]any{
		"projectId": projectID,
		"input": map[string]any{
			"type":        "QUIZ",
			"name":        "Find the position in the temple",
			"buttonText":  "Start",
			"description": "<p>Map pin</p>",
		},
	})
	require.False(t, resp.HasErrors(), "failed to create challenge: %s", resp.ErrorMessage())
	var challenge struct {
		CreateChallenge struct{ ID string } `json:"createChallenge"`
	}
	require.NoError(t, resp.UnmarshalData(&challenge))

	resp = client.MustExecute(t, `
		mutation CreateQuiz($input: CreateQuizInput!) {
			createQuiz(input: $input) { id }
		}
	`, map[string]any{
		"input": map[string]any{
			"projectId":            projectID,
			"challengeId":          challenge.CreateChallenge.ID,
			"name":                 "Map pin schema",
			"description":          "MAP_PIN schema checks",
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
