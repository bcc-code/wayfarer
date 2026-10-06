// Package betting settles bets placed on quiz questions.
//
// Settlement is split into two per-question-type parts:
//   - an Evaluator decides how correct a response was (an Outcome)
//   - a PayoutRule turns that Outcome into a winnings multiplier
//
// The Service combines them and writes the result to the score journal.
// Question types without a registered Strategy are skipped, never guessed.
package betting

import (
	"context"
	"errors"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// QuestionType mirrors quiz_questions.question_type.
type QuestionType string

const (
	QuestionTypePredefined QuestionType = "PREDEFINED"
	QuestionTypeFreeText   QuestionType = "FREE_TEXT"
	QuestionTypeNumber     QuestionType = "NUMBER"
	QuestionTypeJSON       QuestionType = "JSON"
	QuestionTypeOrdering   QuestionType = "ORDERING"
)

// SourceTypeBet is the score_journal.source_type used for bet entries.
const SourceTypeBet = "BET"

// ErrNotEvaluable is returned by an Evaluator when a response cannot be judged,
// e.g. the stored answer is malformed or the question has no correct answers.
var ErrNotEvaluable = errors.New("response cannot be evaluated")

// Response is the part of a quiz response that settlement needs.
// It is decoupled from any specific sqlc row so every caller can build one.
type Response struct {
	ID                string
	QuestionID        string
	QuestionType      QuestionType
	BettingEnabled    bool
	BetAmount         *int32
	ScoreJournalID    *string
	SelectedAnswerIDs []byte // raw JSON array of answer IDs (PREDEFINED)
	JSONResponse      []byte // raw JSON (ORDERING, JSON)
}

// Answer is one predefined answer of a question.
// Answers are passed in their correct order (answer_order), which ORDERING relies on.
type Answer struct {
	ID        string
	IsCorrect bool
}

// Outcome describes how correct a response was: Correct out of Total parts.
// Binary question types use Total = 1.
type Outcome struct {
	Correct int
	Total   int
}

// AllCorrect reports whether every part of the response was correct.
func (o Outcome) AllCorrect() bool {
	return o.Total > 0 && o.Correct == o.Total
}

// Evaluator judges a response for one question type.
type Evaluator interface {
	Evaluate(response Response, answers []Answer) (Outcome, error)
}

// EvaluatorFunc adapts a function to the Evaluator interface.
type EvaluatorFunc func(response Response, answers []Answer) (Outcome, error)

func (f EvaluatorFunc) Evaluate(response Response, answers []Answer) (Outcome, error) {
	return f(response, answers)
}

// PayoutRule returns the winnings multiplier for an outcome.
// Winnings = stake * multiplier; net points = winnings - stake.
type PayoutRule interface {
	Multiplier(outcome Outcome) float64
}

// PayoutRuleFunc adapts a function to the PayoutRule interface.
type PayoutRuleFunc func(outcome Outcome) float64

func (f PayoutRuleFunc) Multiplier(outcome Outcome) float64 {
	return f(outcome)
}

// Strategy is how bets on one question type are settled.
type Strategy struct {
	Evaluator  Evaluator
	PayoutRule PayoutRule
}

// Querier defines the database operations needed to settle bets.
type Querier interface {
	CreateScoreJournalEntry(ctx context.Context, arg sqlc.CreateScoreJournalEntryParams) (*sqlc.ScoreJournal, error)
	UpdateBetResultWithJournal(ctx context.Context, arg sqlc.UpdateBetResultWithJournalParams) (*sqlc.UpdateBetResultWithJournalRow, error)
}
