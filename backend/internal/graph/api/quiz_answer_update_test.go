package api

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyAnswerUpdate(t *testing.T) {
	text := "a new answer"
	number := 42.5
	jsonVal := `{"k":"v"}`

	t.Run("writes the field matching the question type", func(t *testing.T) {
		cases := []struct {
			questionType string
			input        model.UpdateQuizAnswerInput
			assert       func(t *testing.T, p sqlc.UpdateQuizResponseParams)
		}{
			{
				questionType: "PREDEFINED",
				input:        model.UpdateQuizAnswerInput{SelectedAnswerIds: []string{"A1", "A2"}},
				assert: func(t *testing.T, p sqlc.UpdateQuizResponseParams) {
					assert.JSONEq(t, `["A1","A2"]`, string(p.Selectedanswerids))
				},
			},
			{
				questionType: "FREE_TEXT",
				input:        model.UpdateQuizAnswerInput{TextResponse: &text},
				assert: func(t *testing.T, p sqlc.UpdateQuizResponseParams) {
					require.NotNil(t, p.Textresponse)
					assert.Equal(t, text, *p.Textresponse)
				},
			},
			{
				questionType: "NUMBER",
				input:        model.UpdateQuizAnswerInput{NumberResponse: &number},
				assert: func(t *testing.T, p sqlc.UpdateQuizResponseParams) {
					assert.True(t, p.Numberresponse.Valid)
				},
			},
			{
				questionType: "JSON",
				input:        model.UpdateQuizAnswerInput{JSONResponse: &jsonVal},
				assert: func(t *testing.T, p sqlc.UpdateQuizResponseParams) {
					assert.JSONEq(t, jsonVal, string(p.Jsonresponse))
				},
			},
			{
				questionType: "ORDERING",
				input:        model.UpdateQuizAnswerInput{SubmittedOrder: []string{"i2", "i1"}},
				assert: func(t *testing.T, p sqlc.UpdateQuizResponseParams) {
					assert.JSONEq(t, `["i2","i1"]`, string(p.Jsonresponse))
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.questionType, func(t *testing.T) {
				var params sqlc.UpdateQuizResponseParams
				changed, err := applyAnswerUpdate(&params, tc.questionType, quizgrading.MapPinConfig{}, tc.input)
				require.NoError(t, err)
				assert.True(t, changed)
				tc.assert(t, params)
			})
		}
	})

	t.Run("a bet-only update leaves the answer alone", func(t *testing.T) {
		bet := 10
		var params sqlc.UpdateQuizResponseParams
		changed, err := applyAnswerUpdate(&params, "PREDEFINED", quizgrading.MapPinConfig{}, model.UpdateQuizAnswerInput{BetAmount: &bet})
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Nil(t, params.Selectedanswerids)
	})

	t.Run("an empty selection is a change, not an omission", func(t *testing.T) {
		var params sqlc.UpdateQuizResponseParams
		changed, err := applyAnswerUpdate(&params, "PREDEFINED", quizgrading.MapPinConfig{}, model.UpdateQuizAnswerInput{SelectedAnswerIds: []string{}})
		require.NoError(t, err)
		assert.True(t, changed)
		assert.JSONEq(t, `[]`, string(params.Selectedanswerids))
	})

	t.Run("rejects an answer field belonging to another question type", func(t *testing.T) {
		var params sqlc.UpdateQuizResponseParams
		_, err := applyAnswerUpdate(&params, "PREDEFINED", quizgrading.MapPinConfig{}, model.UpdateQuizAnswerInput{SubmittedOrder: []string{"i1"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot set a ORDERING answer on a PREDEFINED question")
	})

	t.Run("rejects any answer field on an unrecognised question type", func(t *testing.T) {
		var params sqlc.UpdateQuizResponseParams
		_, err := applyAnswerUpdate(&params, "MYSTERY", quizgrading.MapPinConfig{}, model.UpdateQuizAnswerInput{SelectedAnswerIds: []string{"A1"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "on a MYSTERY question")
	})

	t.Run("a bet-only update on an unrecognised question type is not a change", func(t *testing.T) {
		var params sqlc.UpdateQuizResponseParams
		changed, err := applyAnswerUpdate(&params, "MYSTERY", quizgrading.MapPinConfig{}, model.UpdateQuizAnswerInput{})
		require.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("MAP_PIN", func(t *testing.T) {
		mapPin := quizgrading.MapPinConfig{Mode: "POLYGON", Width: 1000, Height: 800}

		t.Run("stores the pin as JSON", func(t *testing.T) {
			var params sqlc.UpdateQuizResponseParams
			changed, err := applyAnswerUpdate(&params, "MAP_PIN", mapPin,
				model.UpdateQuizAnswerInput{MapPosition: &model.MapPointInput{X: 12.5, Y: 800}})
			require.NoError(t, err)
			assert.True(t, changed)
			assert.JSONEq(t, `{"x":12.5,"y":800}`, string(params.Jsonresponse))
		})

		t.Run("rejects a pin outside the map", func(t *testing.T) {
			var params sqlc.UpdateQuizResponseParams
			_, err := applyAnswerUpdate(&params, "MAP_PIN", mapPin,
				model.UpdateQuizAnswerInput{MapPosition: &model.MapPointInput{X: 1000.5, Y: 10}})
			assertFieldError(t, err, "mapPosition")
			assert.Nil(t, params.Jsonresponse)
		})

		t.Run("rejects jsonResponse, so the stored JSON is always a validated pin", func(t *testing.T) {
			var params sqlc.UpdateQuizResponseParams
			_, err := applyAnswerUpdate(&params, "MAP_PIN", mapPin, model.UpdateQuizAnswerInput{JSONResponse: &jsonVal})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot set a JSON answer on a MAP_PIN question")
			assert.Nil(t, params.Jsonresponse)
		})

		t.Run("rejects mapPosition on another type", func(t *testing.T) {
			var params sqlc.UpdateQuizResponseParams
			_, err := applyAnswerUpdate(&params, "JSON", quizgrading.MapPinConfig{},
				model.UpdateQuizAnswerInput{MapPosition: &model.MapPointInput{X: 1, Y: 1}})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot set a MAP_PIN answer on a JSON question")
		})
	})
}
