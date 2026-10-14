package api

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateQuestionBetting(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	i := func(v int) *int { return &v }

	tests := []struct {
		name        string
		settings    QuestionBettingSettings
		checkPoints bool
		wantField   string // empty = valid
	}{
		{name: "nothing set", settings: QuestionBettingSettings{}},
		{name: "betting with limits", settings: QuestionBettingSettings{BettingEnabled: true,
			MinPercentage: f(10), MaxPercentage: f(50), MinAbsolute: i(5), MaxAbsolute: i(100)}},
		{name: "points without betting", settings: QuestionBettingSettings{Points: i(10)}, checkPoints: true},
		{name: "betting with 0 points", settings: QuestionBettingSettings{BettingEnabled: true, Points: i(0)}, checkPoints: true},
		{name: "points and betting", settings: QuestionBettingSettings{BettingEnabled: true, Points: i(10)}, checkPoints: true, wantField: "points"},
		{name: "points and betting, not checked", settings: QuestionBettingSettings{BettingEnabled: true, Points: i(10)}},
		{name: "percentage above 100", settings: QuestionBettingSettings{MaxPercentage: f(150)}, wantField: "bettingMaxPercentage"},
		{name: "negative percentage", settings: QuestionBettingSettings{MinPercentage: f(-1)}, wantField: "bettingMinPercentage"},
		{name: "percentages 0 and 100", settings: QuestionBettingSettings{MinPercentage: f(0), MaxPercentage: f(100)}},
		{name: "min percentage above max", settings: QuestionBettingSettings{MinPercentage: f(60), MaxPercentage: f(40)}, wantField: "bettingMinPercentage"},
		{name: "equal percentages", settings: QuestionBettingSettings{MinPercentage: f(40), MaxPercentage: f(40)}},
		{name: "negative absolute", settings: QuestionBettingSettings{MinAbsolute: i(-5)}, wantField: "bettingMinAbsolute"},
		{name: "negative max absolute", settings: QuestionBettingSettings{MaxAbsolute: i(-5)}, wantField: "bettingMaxAbsolute"},
		{name: "min absolute above max", settings: QuestionBettingSettings{MinAbsolute: i(100), MaxAbsolute: i(10)}, wantField: "bettingMinAbsolute"},
		{name: "invalid multiplier", settings: QuestionBettingSettings{MultiplierCorrect: f(1.234)}, wantField: "bettingMultiplierCorrect"},
		{name: "wrong above correct", settings: QuestionBettingSettings{MultiplierCorrect: f(1), MultiplierWrong: f(2)}, wantField: "bettingMultiplierWrong"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuestionBetting(tt.settings, tt.checkPoints)
			if tt.wantField == "" {
				assert.NoError(t, err)
				return
			}
			var betErr *BetValidationError
			require.ErrorAs(t, err, &betErr)
			assert.Equal(t, tt.wantField, betErr.Field)
			assert.NotContains(t, betErr.Error(), "SQLSTATE")
		})
	}
}

func TestPayoutChanged(t *testing.T) {
	num := func(s string) pgtype.Numeric {
		var n pgtype.Numeric
		require.NoError(t, n.Scan(s))
		return n
	}
	f := func(v float64) *float64 { return &v }
	unset := pgtype.Numeric{}

	tests := []struct {
		name                     string
		storedCorrect, storedWrg pgtype.Numeric
		newCorrect, newWrong     *float64
		want                     bool
	}{
		{name: "nothing new", storedCorrect: unset, storedWrg: unset},
		{name: "default set explicitly", storedCorrect: unset, storedWrg: unset, newCorrect: f(2), newWrong: f(0)},
		{name: "same stored value", storedCorrect: num("3.00"), storedWrg: num("0.5"), newCorrect: f(3), newWrong: f(0.5)},
		{name: "correct from default to 5", storedCorrect: unset, storedWrg: unset, newCorrect: f(5), want: true},
		{name: "wrong changed", storedCorrect: num("3"), storedWrg: num("0.5"), newWrong: f(0.25), want: true},
		{name: "stored back to default value", storedCorrect: num("3"), storedWrg: unset, newCorrect: f(2), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := payoutChanged(tt.storedCorrect, tt.storedWrg, tt.newCorrect, tt.newWrong)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSameAnswers(t *testing.T) {
	stored := []answerSpec{{Text: "a", Correct: true, Order: 0}, {Text: "b", Order: 1}}
	tests := []struct {
		name string
		next []answerSpec
		want bool
	}{
		{name: "identical", next: []answerSpec{{Text: "a", Correct: true, Order: 0}, {Text: "b", Order: 1}}, want: true},
		{name: "other list order", next: []answerSpec{{Text: "b", Order: 1}, {Text: "a", Correct: true, Order: 0}}, want: true},
		{name: "text changed", next: []answerSpec{{Text: "A", Correct: true, Order: 0}, {Text: "b", Order: 1}}},
		{name: "correct answer changed", next: []answerSpec{{Text: "a", Order: 0}, {Text: "b", Correct: true, Order: 1}}},
		{name: "order changed", next: []answerSpec{{Text: "a", Correct: true, Order: 1}, {Text: "b", Order: 0}}},
		{name: "answer added", next: []answerSpec{{Text: "a", Correct: true, Order: 0}, {Text: "b", Order: 1}, {Text: "c", Order: 2}}},
		{name: "answer removed", next: []answerSpec{{Text: "a", Correct: true, Order: 0}}},
		{name: "empty", next: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sameAnswers(stored, tt.next))
		})
	}
	assert.True(t, sameAnswers(nil, nil))
}
