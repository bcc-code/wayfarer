package quizgrading

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectionCorrect(t *testing.T) {
	single := []Answer{{ID: "A", IsCorrect: true}, {ID: "B"}, {ID: "C"}}
	multi := []Answer{{ID: "A", IsCorrect: true}, {ID: "B", IsCorrect: true}, {ID: "C"}}

	tests := []struct {
		name     string
		selected []string
		answers  []Answer
		want     bool
	}{
		{name: "single: correct", selected: []string{"A"}, answers: single, want: true},
		{name: "single: wrong", selected: []string{"B"}, answers: single, want: false},
		{name: "single: correct plus wrong", selected: []string{"A", "B"}, answers: single, want: false},
		{name: "single: nothing selected", selected: []string{}, answers: single, want: false},
		{name: "multi: all correct", selected: []string{"A", "B"}, answers: multi, want: true},
		{name: "multi: all correct, any order", selected: []string{"B", "A"}, answers: multi, want: true},
		{name: "multi: only some correct", selected: []string{"A"}, answers: multi, want: false},
		{name: "multi: all correct plus wrong", selected: []string{"A", "B", "C"}, answers: multi, want: false},
		{name: "multi: duplicate does not fill a missing answer", selected: []string{"A", "A"}, answers: multi, want: false},
		{name: "duplicate of a correct answer is ignored", selected: []string{"A", "A"}, answers: single, want: true},
		{name: "unknown answer ID", selected: []string{"X"}, answers: single, want: false},
		{name: "no correct answers, nothing selected", selected: []string{}, answers: []Answer{{ID: "A"}}, want: true},
		{name: "no correct answers, something selected", selected: []string{"A"}, answers: []Answer{{ID: "A"}}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SelectionCorrect(tt.selected, tt.answers))
		})
	}
}

func TestOrderCorrect(t *testing.T) {
	answers := []Answer{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}}

	tests := []struct {
		name      string
		submitted []string
		want      bool
	}{
		{name: "correct order", submitted: []string{"A", "B", "C", "D"}, want: true},
		{name: "partly correct is wrong", submitted: []string{"A", "B", "D", "C"}, want: false},
		{name: "all wrong", submitted: []string{"D", "C", "B", "A"}, want: false},
		{name: "too short", submitted: []string{"A", "B", "C"}, want: false},
		{name: "too long", submitted: []string{"A", "B", "C", "D", "E"}, want: false},
		{name: "empty", submitted: []string{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, OrderCorrect(tt.submitted, answers))
		})
	}
}

func TestGrade(t *testing.T) {
	answers := []Answer{{ID: "A", IsCorrect: true}, {ID: "B"}}

	t.Run("predefined correct", func(t *testing.T) {
		got := Grade(QuestionTypePredefined, Response{SelectedAnswerIDs: []string{"A"}}, answers)
		if assert.NotNil(t, got) {
			assert.True(t, *got)
		}
	})

	t.Run("predefined wrong", func(t *testing.T) {
		got := Grade(QuestionTypePredefined, Response{SelectedAnswerIDs: []string{"B"}}, answers)
		if assert.NotNil(t, got) {
			assert.False(t, *got)
		}
	})

	t.Run("predefined without selection is not graded", func(t *testing.T) {
		assert.Nil(t, Grade(QuestionTypePredefined, Response{}, answers))
	})

	t.Run("ordering correct", func(t *testing.T) {
		got := Grade(QuestionTypeOrdering, Response{SubmittedOrder: []string{"A", "B"}}, answers)
		if assert.NotNil(t, got) {
			assert.True(t, *got)
		}
	})

	t.Run("ordering wrong", func(t *testing.T) {
		got := Grade(QuestionTypeOrdering, Response{SubmittedOrder: []string{"B", "A"}}, answers)
		if assert.NotNil(t, got) {
			assert.False(t, *got)
		}
	})

	t.Run("ordering without order is not graded", func(t *testing.T) {
		assert.Nil(t, Grade(QuestionTypeOrdering, Response{}, answers))
	})

	for _, qt := range []string{"FREE_TEXT", "NUMBER", "JSON", ""} {
		t.Run(qt+" is not graded", func(t *testing.T) {
			resp := Response{SelectedAnswerIDs: []string{"A"}, SubmittedOrder: []string{"A", "B"}}
			assert.Nil(t, Grade(qt, resp, answers))
		})
	}
}

func TestIsGradable(t *testing.T) {
	assert.True(t, IsGradable(QuestionTypePredefined))
	assert.True(t, IsGradable(QuestionTypeOrdering))
	assert.False(t, IsGradable("FREE_TEXT"))
	assert.False(t, IsGradable("NUMBER"))
	assert.False(t, IsGradable("JSON"))
}
