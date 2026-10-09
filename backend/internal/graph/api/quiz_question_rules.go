package api

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/betting"
	"github.com/jackc/pgx/v5/pgtype"
)

// QuestionBettingSettings are a question's points and betting settings after a
// create or update: the new value where one is given, otherwise the stored one.
type QuestionBettingSettings struct {
	Points            *int
	BettingEnabled    bool
	MinPercentage     *float64
	MaxPercentage     *float64
	MinAbsolute       *int
	MaxAbsolute       *int
	MultiplierCorrect *float64
	MultiplierWrong   *float64
}

// ValidateQuestionBetting checks a question's betting settings: limits in
// range and min <= max, valid multipliers, and, when checkPointsWithBetting is
// set, that a betting question doesn't also award points (a settled bet would
// overwrite them on the response). Updates only check the last rule when they
// change points or betting, so older questions with both stay editable.
func ValidateQuestionBetting(s QuestionBettingSettings, checkPointsWithBetting bool) error {
	if checkPointsWithBetting && s.BettingEnabled && s.Points != nil && *s.Points > 0 {
		return &BetValidationError{Field: "points", Message: "a question with betting cannot also award points"}
	}

	for _, p := range []struct {
		field string
		value *float64
	}{{"bettingMinPercentage", s.MinPercentage}, {"bettingMaxPercentage", s.MaxPercentage}} {
		if p.value != nil && (*p.value < 0 || *p.value > 100) {
			return &BetValidationError{Field: p.field, Message: "must be between 0 and 100"}
		}
	}
	if s.MinPercentage != nil && s.MaxPercentage != nil && *s.MinPercentage > *s.MaxPercentage {
		return &BetValidationError{Field: "bettingMinPercentage", Message: "must not be greater than bettingMaxPercentage"}
	}

	for _, a := range []struct {
		field string
		value *int
	}{{"bettingMinAbsolute", s.MinAbsolute}, {"bettingMaxAbsolute", s.MaxAbsolute}} {
		if a.value != nil && *a.value < 0 {
			return &BetValidationError{Field: a.field, Message: "must not be negative"}
		}
	}
	if s.MinAbsolute != nil && s.MaxAbsolute != nil && *s.MinAbsolute > *s.MaxAbsolute {
		return &BetValidationError{Field: "bettingMinAbsolute", Message: "must not be greater than bettingMaxAbsolute"}
	}

	return ValidateBettingMultipliers(s.MultiplierCorrect, s.MultiplierWrong)
}

// storedBettingSettings returns a stored question's points and betting settings.
func storedBettingSettings(q *sqlc.GetQuizQuestionByIDRow) QuestionBettingSettings {
	s := QuestionBettingSettings{
		BettingEnabled:    q.BettingEnabled,
		MinPercentage:     numericToFloat(q.BettingMinPercentage),
		MaxPercentage:     numericToFloat(q.BettingMaxPercentage),
		MultiplierCorrect: numericToFloat(q.BettingMultiplierCorrect),
		MultiplierWrong:   numericToFloat(q.BettingMultiplierWrong),
	}
	if q.Points != nil {
		v := int(*q.Points)
		s.Points = &v
	}
	if q.BettingMinAbsolute != nil {
		v := int(*q.BettingMinAbsolute)
		s.MinAbsolute = &v
	}
	if q.BettingMaxAbsolute != nil {
		v := int(*q.BettingMaxAbsolute)
		s.MaxAbsolute = &v
	}
	return s
}

// payoutChanged reports whether new multipliers (nil = keep the stored one)
// would pay a correct or wrong answer differently than the stored ones,
// defaults included. The new values must be valid (ValidateBettingMultipliers).
func payoutChanged(storedCorrect, storedWrong pgtype.Numeric, newCorrect, newWrong *float64) (bool, error) {
	oldCorrect, err := numericToMultiplier(storedCorrect)
	if err != nil {
		return false, err
	}
	oldWrong, err := numericToMultiplier(storedWrong)
	if err != nil {
		return false, err
	}
	nextCorrect, nextWrong := oldCorrect, oldWrong
	if newCorrect != nil {
		if nextCorrect, err = multiplierToFixed("bettingMultiplierCorrect", newCorrect); err != nil {
			return false, err
		}
	}
	if newWrong != nil {
		if nextWrong, err = multiplierToFixed("bettingMultiplierWrong", newWrong); err != nil {
			return false, err
		}
	}
	for _, correct := range []bool{true, false} {
		if betting.Multiplier(correct, oldCorrect, oldWrong) != betting.Multiplier(correct, nextCorrect, nextWrong) {
			return true, nil
		}
	}
	return false, nil
}

// answerSpec is an answer (or ordering item) as an admin edits it.
type answerSpec struct {
	Text    string
	Correct bool
	Order   int32
}

// sameAnswers reports whether two answer lists are the same, ignoring list order.
func sameAnswers(a, b []answerSpec) bool {
	if len(a) != len(b) {
		return false
	}
	sorted := func(s []answerSpec) []answerSpec {
		s = slices.Clone(s)
		slices.SortFunc(s, func(x, y answerSpec) int {
			if x.Order != y.Order {
				return int(x.Order - y.Order)
			}
			return strings.Compare(x.Text, y.Text)
		})
		return s
	}
	return slices.Equal(sorted(a), sorted(b))
}

// storedAnswerSpecs converts stored answers for sameAnswers.
func storedAnswerSpecs(rows []*sqlc.QuizPredefinedAnswer) []answerSpec {
	specs := make([]answerSpec, len(rows))
	for i, r := range rows {
		specs[i] = answerSpec{Text: r.AnswerText, Correct: r.IsCorrect, Order: r.AnswerOrder}
	}
	return specs
}

// openBetsError is returned for a change that would alter bets already placed.
func openBetsError(field string, count int64) error {
	return &BetValidationError{
		Field:   field,
		Message: fmt.Sprintf("the question has %d open bets; finish the session first, or delete the question to cancel them", count),
	}
}

// answeredError is returned for replacing answers that users already chose from.
func answeredError(field string) error {
	return &BetValidationError{
		Field:   field,
		Message: "can't be changed after users have answered this question; delete it and add a new one",
	}
}
