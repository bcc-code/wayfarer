package loaders

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
)

func ptr[T any](v T) *T { return &v }

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
			MapTargetX:          ptr(500.0),
			MapTargetY:          ptr(300.0),
			MapFullPointsRadius: ptr(40.0),
			MapZeroPointsRadius: ptr(200.0),
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
			FullRadius: ptr(40.0),
			ZeroRadius: ptr(200.0),
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
			MapPolygon:     []byte(`[{"x": 100, "y": 100}, {"x": 300, "y": 100}, {"x": 200, "y": 300}]`),
		}).(*model.MapPinQuestion)
		require.True(t, ok)

		assert.Equal(t, quizgrading.MapPinConfig{
			Mode:    "POLYGON",
			Width:   1000,
			Height:  800,
			Polygon: []quizgrading.MapPoint{{X: 100, Y: 100}, {X: 300, Y: 100}, {X: 200, Y: 300}},
		}, q.MapPin)
	})

	t.Run("an unreadable polygon still loads the question", func(t *testing.T) {
		q, ok := quizQuestionFromRow(&sqlc.GetQuizQuestionsByQuizIDsRow{
			ID:             "QQ003",
			QuestionType:   "MAP_PIN",
			MapScoringMode: ptr("POLYGON"),
			MapWidth:       ptr(int32(1000)),
			MapHeight:      ptr(int32(800)),
			MapPolygon:     []byte(`[1, 2, 3]`),
		}).(*model.MapPinQuestion)
		require.True(t, ok)
		assert.Equal(t, "POLYGON", q.MapPin.Mode)
		assert.Nil(t, q.MapPin.Polygon)
	})
}
