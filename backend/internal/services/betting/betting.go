// Package betting settles bets placed on quiz questions.
//
// A bet is all or nothing: the response is either correct or wrong (see
// quizgrading), and the question's payout multiplier for that result decides
// the winnings. winnings = bet * multiplier; net points = winnings - bet.
package betting

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// SourceTypeBet is the score_journal.source_type used for bet entries.
const SourceTypeBet = "BET"

// Default payout multipliers, used when a question does not set its own.
const (
	DefaultMultiplierCorrect = 2.0 // bet doubled
	DefaultMultiplierWrong   = 0.0 // bet lost
)

// Querier defines the database operations needed to settle a bet.
// Pass a transaction-bound querier so a bet is settled completely or not at all.
type Querier interface {
	CreateScoreJournalEntry(ctx context.Context, arg sqlc.CreateScoreJournalEntryParams) (*sqlc.ScoreJournal, error)
	SettleBetResult(ctx context.Context, arg sqlc.SettleBetResultParams) (int64, error)
}

// Bet is one response with a bet that is ready to be settled.
type Bet struct {
	ResponseID string
	UserID     string
	Amount     int32
	// IsCorrect is the response's stored is_correct. Nil means not graded.
	IsCorrect *bool
	// Question payout multipliers. Nil means the default.
	MultiplierCorrect *float64
	MultiplierWrong   *float64
}

// Multiplier returns the payout multiplier for a correct or wrong answer,
// falling back to the defaults when the question does not set one.
func Multiplier(correct bool, multiplierCorrect, multiplierWrong *float64) float64 {
	if correct {
		if multiplierCorrect != nil {
			return *multiplierCorrect
		}
		return DefaultMultiplierCorrect
	}
	if multiplierWrong != nil {
		return *multiplierWrong
	}
	return DefaultMultiplierWrong
}
