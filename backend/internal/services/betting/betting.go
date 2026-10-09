// Package betting settles bets placed on quiz questions.
//
// A bet is all or nothing: the response is either correct or wrong (see
// quizgrading), and the question's payout multiplier for that result decides
// the winnings. winnings = bet * multiplier; net points = winnings - bet.
//
// Multipliers are fixed-point (hundredths, see MultiplierScale) so payouts are
// computed with integers only. Winnings are rounded down.
package betting

import (
	"context"
	"errors"
	"math"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// SourceTypeBet is the score_journal.source_type used for bet entries.
const SourceTypeBet = "BET"

// MultiplierScale is the fixed-point scale of multipliers: 2.5x is 250.
const MultiplierScale = 100

// Multiplier limits and defaults, in hundredths.
const (
	DefaultMultiplierCorrect int64 = 2 * MultiplierScale // bet doubled
	DefaultMultiplierWrong   int64 = 0                   // bet lost
	MaxMultiplier            int64 = 100 * MultiplierScale
)

// ErrMultiplierOutOfRange is returned for a multiplier below 0 or above MaxMultiplier.
var ErrMultiplierOutOfRange = errors.New("multiplier out of range")

// ErrMultiplierPrecision is returned for a multiplier with more than 2 decimals.
var ErrMultiplierPrecision = errors.New("multiplier has more than 2 decimals")

// ErrPayoutOutOfRange is returned when winnings do not fit in the points range.
var ErrPayoutOutOfRange = errors.New("payout out of points range")

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
	// IsCorrect is the response's stored is_correct. Nil means not graded (no
	// answer given): the bet is void and the stake is returned.
	IsCorrect *bool
	// Question payout multipliers in hundredths. Nil means the default.
	MultiplierCorrect *int64
	MultiplierWrong   *int64
}

// Multiplier returns the payout multiplier (hundredths) for a correct or wrong
// answer, falling back to the defaults when the question does not set one.
func Multiplier(correct bool, multiplierCorrect, multiplierWrong *int64) int64 {
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

// MultiplierFromFloat converts an API multiplier (e.g. 2.5) to hundredths.
// It rejects values outside 0..MaxMultiplier and values with more than 2 decimals.
func MultiplierFromFloat(v float64) (int64, error) {
	if math.IsNaN(v) || v < 0 || v > float64(MaxMultiplier)/MultiplierScale {
		return 0, ErrMultiplierOutOfRange
	}
	scaled := math.Round(v * MultiplierScale)
	// Tolerance only absorbs float representation error (e.g. 0.29 * 100)
	if math.Abs(v*MultiplierScale-scaled) > 1e-6 {
		return 0, ErrMultiplierPrecision
	}
	return int64(scaled), nil
}

// Winnings returns floor(stake * multiplier), using integer arithmetic only.
func Winnings(stake int32, multiplier int64) (int32, error) {
	if stake < 0 || multiplier < 0 || multiplier > MaxMultiplier {
		return 0, ErrPayoutOutOfRange
	}
	// stake <= MaxInt32 and multiplier <= MaxMultiplier, so this cannot overflow int64
	winnings := int64(stake) * multiplier / MultiplierScale
	if winnings > math.MaxInt32 {
		return 0, ErrPayoutOutOfRange
	}
	return int32(winnings), nil
}
