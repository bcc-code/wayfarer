package loaders

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
	"github.com/jackc/pgx/v5/pgtype"
)

func ptr[T any](v T) *T { return &v }

func pgPoint(x, y float64) pgtype.Point {
	return pgtype.Point{P: pgtype.Vec2{X: x, Y: y}, Valid: true}
}

func TestQuizQuestionFromRowMapPin(t *testing.T) {
	t.Run("POINT_RADIUS", func(t *testing.T) {
		q, ok := quizQuestionFromRow(&sqlc.GetQuizQuestionsByQuizIDsRow{
			ID:                  "QQ001",
			QuizID:              "QZ001",
			QuestionType:        "MAP_PIN",
			QuestionText:        "Where was the photo taken?",
			ImageUrl:            ptr("https://example.com/map.png"),
			QuestionOrder:       2,
			Points:              ptr(int32(100)),
			MapScoringMode:      ptr("POINT_RADIUS"),
			MapWidth:            ptr(int32(1000)),
			MapHeight:           ptr(int32(800)),
			MapTarget:           pgPoint(500, 300),
			MapFullPointsRadius: ptr(int32(40)),
			MapZeroPointsRadius: ptr(int32(200)),
		}).(*model.MapPinQuestion)
		require.True(t, ok, "MAP_PIN must not fall back to FreeTextQuestion")

		assert.Equal(t, "QQ001", q.ID)
		assert.Equal(t, "QZ001", q.QuizID)
		assert.Equal(t, "Where was the photo taken?", q.QuestionText)
		assert.Equal(t, ptr("https://example.com/map.png"), q.ImageURL)
		assert.Equal(t, 2, q.QuestionOrder)
		assert.Equal(t, ptr(100), q.Points)
		assert.Equal(t, quizgrading.MapPinConfig{
			Mode:       "POINT_RADIUS",
			Width:      1000,
			Height:     800,
			Target:     &quizgrading.MapPoint{X: 500, Y: 300},
			FullRadius: ptr(40),
			ZeroRadius: ptr(200),
		}, q.MapPin)
	})

	t.Run("POLYGON", func(t *testing.T) {
		q, ok := quizQuestionFromRow(&sqlc.GetQuizQuestionsByQuizIDsRow{
			ID:             "QQ002",
			QuestionType:   "MAP_PIN",
			Points:         ptr(int32(50)),
			MapScoringMode: ptr("POLYGON"),
			MapWidth:       ptr(int32(1000)),
			MapHeight:      ptr(int32(800)),
			MapPolygon:     []pgtype.Point{pgPoint(100, 100), pgPoint(300, 100), pgPoint(200, 300)},
		}).(*model.MapPinQuestion)
		require.True(t, ok)

		assert.Equal(t, quizgrading.MapPinConfig{
			Mode:    "POLYGON",
			Width:   1000,
			Height:  800,
			Polygon: []quizgrading.MapPoint{{X: 100, Y: 100}, {X: 300, Y: 100}, {X: 200, Y: 300}},
		}, q.MapPin)
	})

	t.Run("a polygon with a NULL vertex still loads the question", func(t *testing.T) {
		q, ok := quizQuestionFromRow(&sqlc.GetQuizQuestionsByQuizIDsRow{
			ID:             "QQ003",
			QuestionType:   "MAP_PIN",
			MapScoringMode: ptr("POLYGON"),
			MapWidth:       ptr(int32(1000)),
			MapHeight:      ptr(int32(800)),
			MapPolygon:     []pgtype.Point{pgPoint(100, 100), {}, pgPoint(200, 300)},
		}).(*model.MapPinQuestion)
		require.True(t, ok)
		assert.Equal(t, "POLYGON", q.MapPin.Mode)
		assert.Nil(t, q.MapPin.Polygon)
	})
}
