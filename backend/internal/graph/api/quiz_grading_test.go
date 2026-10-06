package api

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
	"github.com/stretchr/testify/assert"
)

func TestToGradingAnswers(t *testing.T) {
	answers := []*model.QuizPredefinedAnswer{
		{ID: "A", IsCorrectValue: true},
		{ID: "B"},
		{ID: "C", IsCorrectValue: true},
	}

	assert.Equal(t, []quizgrading.Answer{
		{ID: "A", IsCorrect: true},
		{ID: "B", IsCorrect: false},
		{ID: "C", IsCorrect: true},
	}, toGradingAnswers(answers))
}

func TestToGradingAnswers_Empty(t *testing.T) {
	assert.Empty(t, toGradingAnswers(nil))
}
