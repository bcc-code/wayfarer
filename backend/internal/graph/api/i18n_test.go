package api

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func TestApplyOptionalStringTranslation(t *testing.T) {
	tests := []struct {
		name       string
		translated *string
		base       *string
		expected   *string
	}{
		{name: "translation replaces base", translated: strPtr("A pot"), base: strPtr("En gryte"), expected: strPtr("A pot")},
		{name: "missing translation falls back to base", translated: nil, base: strPtr("En gryte"), expected: strPtr("En gryte")},
		{name: "empty translation falls back to base", translated: strPtr(""), base: strPtr("En gryte"), expected: strPtr("En gryte")},
		{name: "no base stays empty despite a translation", translated: strPtr("A pot"), base: nil, expected: nil},
		{name: "no base and no translation stays empty", translated: nil, base: nil, expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, applyOptionalStringTranslation(tt.translated, tt.base))
		})
	}
}

func TestApplyQuizAnswerTranslation(t *testing.T) {
	newAnswer := func() *model.QuizPredefinedAnswer {
		return &model.QuizPredefinedAnswer{
			ID:             "QA01ARZ3NDEKTSV4RRFFQ69G5FAV",
			AnswerText:     "Kjøttgryten fra Egypt",
			Description:    strPtr("En gryte som symboliserer maten."),
			AnswerOrder:    1,
			IsCorrectValue: true,
			QuestionID:     "QQ01ARZ3NDEKTSV4RRFFQ69G5FAV",
		}
	}

	t.Run("translates text and description", func(t *testing.T) {
		answer := newAnswer()
		got := applyQuizAnswerTranslation(answer, &loaders.Translation{
			AnswerText:  strPtr("The meat pot from Egypt"),
			Description: strPtr("A pot symbolising the food."),
		})

		require.NotNil(t, got)
		assert.Equal(t, "The meat pot from Egypt", got.AnswerText)
		assert.Equal(t, strPtr("A pot symbolising the food."), got.Description)
		// Untranslated fields are carried over unchanged
		assert.Equal(t, answer.ID, got.ID)
		assert.Equal(t, answer.AnswerOrder, got.AnswerOrder)
		assert.Equal(t, answer.IsCorrectValue, got.IsCorrectValue)
		assert.Equal(t, answer.QuestionID, got.QuestionID)
	})

	t.Run("untranslated description falls back to the base description", func(t *testing.T) {
		got := applyQuizAnswerTranslation(newAnswer(), &loaders.Translation{
			AnswerText: strPtr("The meat pot from Egypt"),
		})

		assert.Equal(t, "The meat pot from Egypt", got.AnswerText)
		assert.Equal(t, strPtr("En gryte som symboliserer maten."), got.Description)
	})

	t.Run("does not modify the cached base answer", func(t *testing.T) {
		answer := newAnswer()
		applyQuizAnswerTranslation(answer, &loaders.Translation{
			AnswerText:  strPtr("The meat pot from Egypt"),
			Description: strPtr("A pot symbolising the food."),
		})

		assert.Equal(t, "Kjøttgryten fra Egypt", answer.AnswerText)
		assert.Equal(t, strPtr("En gryte som symboliserer maten."), answer.Description)
	})
}
