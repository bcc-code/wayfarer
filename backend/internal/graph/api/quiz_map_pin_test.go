package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
)

func pointRadiusInput() *model.MapPinConfigInput {
	return &model.MapPinConfigInput{
		ScoringMode:      model.MapPinScoringModePointRadius,
		MapWidth:         1000,
		MapHeight:        800,
		Target:           &model.MapPointInput{X: 500, Y: 300},
		FullPointsRadius: ptrTo(40.0),
		ZeroPointsRadius: ptrTo(200.0),
	}
}

func triangleInput() *model.MapPinConfigInput {
	return &model.MapPinConfigInput{
		ScoringMode: model.MapPinScoringModePolygon,
		MapWidth:    1000,
		MapHeight:   800,
		Polygon:     []model.MapPointInput{{X: 100, Y: 100}, {X: 300, Y: 100}, {X: 200, Y: 300}},
	}
}

func ptrTo[T any](v T) *T { return &v }

// assertFieldError checks that err is a field error for field.
func assertFieldError(t *testing.T, err error, field string) {
	t.Helper()
	var fieldErr *BetValidationError
	require.True(t, errors.As(err, &fieldErr), "want a field error, got %v", err)
	assert.Equal(t, field, fieldErr.Field)
}

func TestCheckMapPinCreate(t *testing.T) {
	mapPin := func(modify func(*model.CreateQuizQuestionInput)) model.CreateQuizQuestionInput {
		in := model.CreateQuizQuestionInput{
			QuestionType: model.QuizQuestionTypeMapPin,
			Points:       ptrTo(100),
			MapPin:       pointRadiusInput(),
		}
		if modify != nil {
			modify(&in)
		}
		return in
	}

	t.Run("POINT_RADIUS", func(t *testing.T) {
		c, err := checkMapPinCreate(mapPin(nil))
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, quizgrading.MapPinModePointRadius, c.Mode)
		assert.Equal(t, 1000, c.Width)
		assert.Equal(t, 800, c.Height)
		assert.Equal(t, &quizgrading.MapPoint{X: 500, Y: 300}, c.Target)
		assert.Equal(t, 40.0, *c.FullRadius)
		assert.Equal(t, 200.0, *c.ZeroRadius)
		assert.Nil(t, c.Polygon)
	})

	t.Run("POLYGON is normalized", func(t *testing.T) {
		c, err := checkMapPinCreate(mapPin(func(in *model.CreateQuizQuestionInput) {
			in.MapPin = triangleInput()
			in.MapPin.Polygon = append(in.MapPin.Polygon, model.MapPointInput{X: 100, Y: 100})
		}))
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, []quizgrading.MapPoint{{X: 100, Y: 100}, {X: 300, Y: 100}, {X: 200, Y: 300}}, c.Polygon,
			"the closing vertex is dropped")
		assert.Nil(t, c.Target)
	})

	t.Run("betting explicitly off is fine", func(t *testing.T) {
		_, err := checkMapPinCreate(mapPin(func(in *model.CreateQuizQuestionInput) { in.BettingEnabled = ptrTo(false) }))
		assert.NoError(t, err)
	})

	t.Run("other types have no config", func(t *testing.T) {
		c, err := checkMapPinCreate(model.CreateQuizQuestionInput{QuestionType: model.QuizQuestionTypePredefined})
		require.NoError(t, err)
		assert.Nil(t, c)
	})

	rejected := []struct {
		name  string
		input model.CreateQuizQuestionInput
		field string
	}{
		{
			name:  "mapPin on another type",
			input: model.CreateQuizQuestionInput{QuestionType: model.QuizQuestionTypeFreeText, MapPin: pointRadiusInput()},
			field: "mapPin",
		},
		{name: "no mapPin", input: mapPin(func(in *model.CreateQuizQuestionInput) { in.MapPin = nil }), field: "mapPin"},
		{name: "no points", input: mapPin(func(in *model.CreateQuizQuestionInput) { in.Points = nil }), field: "points"},
		{name: "zero points", input: mapPin(func(in *model.CreateQuizQuestionInput) { in.Points = ptrTo(0) }), field: "points"},
		{name: "negative points", input: mapPin(func(in *model.CreateQuizQuestionInput) { in.Points = ptrTo(-5) }), field: "points"},
		{name: "betting", input: mapPin(func(in *model.CreateQuizQuestionInput) { in.BettingEnabled = ptrTo(true) }), field: "bettingEnabled"},
		{
			name:  "invalid geometry names the input field",
			input: mapPin(func(in *model.CreateQuizQuestionInput) { in.MapPin.ZeroPointsRadius = ptrTo(10.0) }),
			field: "mapPin.zeroPointsRadius",
		},
		{
			name: "other mode's field",
			input: mapPin(func(in *model.CreateQuizQuestionInput) {
				in.MapPin = triangleInput()
				in.MapPin.Target = &model.MapPointInput{X: 1, Y: 1}
			}),
			field: "mapPin.target",
		},
		{
			name: "self-intersecting polygon",
			input: mapPin(func(in *model.CreateQuizQuestionInput) {
				in.MapPin = triangleInput()
				in.MapPin.Polygon = []model.MapPointInput{{X: 100, Y: 100}, {X: 300, Y: 300}, {X: 300, Y: 100}, {X: 100, Y: 400}}
			}),
			field: "mapPin.polygon",
		},
		{
			name: "vertex outside the map",
			input: mapPin(func(in *model.CreateQuizQuestionInput) {
				in.MapPin = triangleInput()
				in.MapPin.Polygon[2].Y = 801
			}),
			field: "mapPin.polygon[2]",
		},
	}
	for _, tt := range rejected {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			c, err := checkMapPinCreate(tt.input)
			assertFieldError(t, err, tt.field)
			assert.Nil(t, c)
		})
	}
}

// storedMapPinRow is a stored POINT_RADIUS question matching pointRadiusInput
// with 100 points.
func storedMapPinRow(t *testing.T) *sqlc.GetQuizQuestionByIDRow {
	t.Helper()
	c, err := checkMapPinCreate(model.CreateQuizQuestionInput{
		QuestionType: model.QuizQuestionTypeMapPin,
		Points:       ptrTo(100),
		MapPin:       pointRadiusInput(),
	})
	require.NoError(t, err)
	cols, err := c.Columns()
	require.NoError(t, err)
	return &sqlc.GetQuizQuestionByIDRow{
		ID:                  "QQ001",
		QuestionType:        "MAP_PIN",
		Points:              ptrTo(int32(100)),
		MapScoringMode:      cols.Mode,
		MapWidth:            cols.Width,
		MapHeight:           cols.Height,
		MapTargetX:          cols.TargetX,
		MapTargetY:          cols.TargetY,
		MapFullPointsRadius: cols.FullRadius,
		MapZeroPointsRadius: cols.ZeroRadius,
		MapPolygon:          cols.Polygon,
	}
}

func TestCheckMapPinUpdate(t *testing.T) {
	t.Run("no map fields leaves the config alone", func(t *testing.T) {
		u, err := checkMapPinUpdate(storedMapPinRow(t), model.UpdateQuizQuestionInput{QuestionText: ptrTo("Where?")})
		require.NoError(t, err)
		assert.Equal(t, mapPinUpdate{}, u)
	})

	t.Run("resending the same config and points doesn't lock", func(t *testing.T) {
		u, err := checkMapPinUpdate(storedMapPinRow(t), model.UpdateQuizQuestionInput{
			Points: ptrTo(100),
			MapPin: pointRadiusInput(),
		})
		require.NoError(t, err)
		require.NotNil(t, u.Config)
		assert.Empty(t, u.LockedField)
	})

	t.Run("resending a polygon with a closing vertex doesn't lock", func(t *testing.T) {
		stored := storedMapPinRow(t)
		c, err := checkMapPinCreate(model.CreateQuizQuestionInput{
			QuestionType: model.QuizQuestionTypeMapPin, Points: ptrTo(100), MapPin: triangleInput(),
		})
		require.NoError(t, err)
		cols, err := c.Columns()
		require.NoError(t, err)
		stored.MapScoringMode, stored.MapPolygon = cols.Mode, cols.Polygon
		stored.MapTargetX, stored.MapTargetY = nil, nil
		stored.MapFullPointsRadius, stored.MapZeroPointsRadius = nil, nil

		resent := triangleInput()
		resent.Polygon = append(resent.Polygon, resent.Polygon[0])
		u, err := checkMapPinUpdate(stored, model.UpdateQuizQuestionInput{MapPin: resent})
		require.NoError(t, err)
		assert.Empty(t, u.LockedField)
	})

	t.Run("changing the config locks on mapPin", func(t *testing.T) {
		changed := pointRadiusInput()
		changed.FullPointsRadius = ptrTo(50.0)
		u, err := checkMapPinUpdate(storedMapPinRow(t), model.UpdateQuizQuestionInput{MapPin: changed})
		require.NoError(t, err)
		require.NotNil(t, u.Config)
		assert.Equal(t, 50.0, *u.Config.FullRadius)
		assert.Equal(t, "mapPin", u.LockedField)
	})

	t.Run("switching mode returns only the new mode's fields", func(t *testing.T) {
		u, err := checkMapPinUpdate(storedMapPinRow(t), model.UpdateQuizQuestionInput{MapPin: triangleInput()})
		require.NoError(t, err)
		require.NotNil(t, u.Config)
		assert.Equal(t, "mapPin", u.LockedField)

		cols, err := u.Config.Columns()
		require.NoError(t, err)
		assert.Equal(t, quizgrading.MapPinModePolygon, *cols.Mode)
		assert.NotNil(t, cols.Polygon)
		assert.Nil(t, cols.TargetX, "the old mode's columns must be written as NULL")
		assert.Nil(t, cols.TargetY)
		assert.Nil(t, cols.FullRadius)
		assert.Nil(t, cols.ZeroRadius)
	})

	t.Run("changing points locks on points", func(t *testing.T) {
		u, err := checkMapPinUpdate(storedMapPinRow(t), model.UpdateQuizQuestionInput{Points: ptrTo(50)})
		require.NoError(t, err)
		assert.Nil(t, u.Config)
		assert.Equal(t, "points", u.LockedField)
	})

	t.Run("an unreadable stored config counts as changed", func(t *testing.T) {
		stored := storedMapPinRow(t)
		stored.MapScoringMode = ptrTo(quizgrading.MapPinModePolygon)
		stored.MapTargetX, stored.MapTargetY = nil, nil
		stored.MapFullPointsRadius, stored.MapZeroPointsRadius = nil, nil
		stored.MapPolygon = []byte(`[1, 2, 3]`)

		u, err := checkMapPinUpdate(stored, model.UpdateQuizQuestionInput{MapPin: triangleInput()})
		require.NoError(t, err)
		assert.Equal(t, "mapPin", u.LockedField)
	})

	t.Run("other types ignore everything but a mapPin", func(t *testing.T) {
		u, err := checkMapPinUpdate(&sqlc.GetQuizQuestionByIDRow{QuestionType: "PREDEFINED"}, model.UpdateQuizQuestionInput{
			Points:         ptrTo(0),
			BettingEnabled: ptrTo(true),
		})
		require.NoError(t, err)
		assert.Equal(t, mapPinUpdate{}, u)
	})

	rejected := []struct {
		name   string
		stored func(*testing.T) *sqlc.GetQuizQuestionByIDRow
		input  model.UpdateQuizQuestionInput
		field  string
	}{
		{
			name: "mapPin on another type",
			stored: func(*testing.T) *sqlc.GetQuizQuestionByIDRow {
				return &sqlc.GetQuizQuestionByIDRow{QuestionType: "ORDERING"}
			},
			input: model.UpdateQuizQuestionInput{MapPin: pointRadiusInput()},
			field: "mapPin",
		},
		{name: "zero points", stored: storedMapPinRow, input: model.UpdateQuizQuestionInput{Points: ptrTo(0)}, field: "points"},
		{name: "betting", stored: storedMapPinRow, input: model.UpdateQuizQuestionInput{BettingEnabled: ptrTo(true)}, field: "bettingEnabled"},
		{
			name:   "invalid config",
			stored: storedMapPinRow,
			input: model.UpdateQuizQuestionInput{MapPin: &model.MapPinConfigInput{
				ScoringMode: model.MapPinScoringModePolygon, MapWidth: 1000, MapHeight: 800,
				Polygon: []model.MapPointInput{{X: 1, Y: 1}, {X: 2, Y: 2}},
			}},
			field: "mapPin.polygon",
		},
	}
	for _, tt := range rejected {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			_, err := checkMapPinUpdate(tt.stored(t), tt.input)
			assertFieldError(t, err, tt.field)
		})
	}
}

func TestConvertMapPinQuestion(t *testing.T) {
	t.Run("POINT_RADIUS", func(t *testing.T) {
		row := &mockQuestionRow{
			id: "QQ001", quizID: "QZ001", questionType: "MAP_PIN", questionText: "Where was the photo taken?",
			points: int32Ptr(100),
			mapPin: quizgrading.MapPinColumns{
				Mode: ptrTo("POINT_RADIUS"), Width: ptrTo(int32(1000)), Height: ptrTo(int32(800)),
				TargetX: ptrTo(500.0), TargetY: ptrTo(300.0), FullRadius: ptrTo(40.0), ZeroRadius: ptrTo(200.0),
			},
		}

		q, ok := convertQuizQuestionRowToInterface(row).(*model.MapPinQuestion)
		require.True(t, ok, "MAP_PIN must not fall back to another question type")
		assert.Equal(t, "QQ001", q.ID)
		assert.Equal(t, "QZ001", q.QuizID)
		assert.Equal(t, "Where was the photo taken?", q.QuestionText)
		assert.Equal(t, ptrTo(100), q.Points)
		assert.Equal(t, quizgrading.MapPinConfig{
			Mode: "POINT_RADIUS", Width: 1000, Height: 800,
			Target:     &quizgrading.MapPoint{X: 500, Y: 300},
			FullRadius: ptrTo(40.0), ZeroRadius: ptrTo(200.0),
		}, q.MapPin)
		assert.Equal(t, "MAP_PIN", quizQuestionTypeString(q))
	})

	t.Run("POLYGON", func(t *testing.T) {
		row := &mockQuestionRow{
			id: "QQ002", questionType: "MAP_PIN",
			mapPin: quizgrading.MapPinColumns{
				Mode: ptrTo("POLYGON"), Width: ptrTo(int32(1000)), Height: ptrTo(int32(800)),
				Polygon: []byte(`[{"x":100,"y":100},{"x":300,"y":100},{"x":200,"y":300}]`),
			},
		}

		q, ok := convertQuizQuestionRowToInterface(row).(*model.MapPinQuestion)
		require.True(t, ok)
		assert.Equal(t, []quizgrading.MapPoint{{X: 100, Y: 100}, {X: 300, Y: 100}, {X: 200, Y: 300}}, q.MapPin.Polygon)
		assert.Nil(t, q.MapPin.Target)
	})

	t.Run("an unreadable polygon still loads the question", func(t *testing.T) {
		row := &mockQuestionRow{
			id: "QQ003", questionType: "MAP_PIN",
			mapPin: quizgrading.MapPinColumns{
				Mode: ptrTo("POLYGON"), Width: ptrTo(int32(1000)), Height: ptrTo(int32(800)),
				Polygon: []byte(`"not a polygon"`),
			},
		}

		q, ok := convertQuizQuestionRowToInterface(row).(*model.MapPinQuestion)
		require.True(t, ok)
		assert.Equal(t, "POLYGON", q.MapPin.Mode)
		assert.Nil(t, q.MapPin.Polygon)
	})

	t.Run("the question text is translated", func(t *testing.T) {
		q := &model.MapPinQuestion{QuestionText: "Where?", MapPin: quizgrading.MapPinConfig{Mode: "POINT_RADIUS"}}
		translated, ok := applyQuizQuestionTranslation(q, &loaders.Translation{QuestionText: ptrTo("Hvor?")}).(*model.MapPinQuestion)
		require.True(t, ok)
		assert.Equal(t, "Hvor?", translated.QuestionText)
		assert.Equal(t, "POINT_RADIUS", translated.MapPin.Mode)
		assert.Equal(t, "Where?", q.QuestionText, "the cached question must not be changed")
	})
}
