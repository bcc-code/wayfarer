package betting

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bcc-media/wayfarer/i18n"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SkipReason explains why a bet was not settled.
type SkipReason string

const (
	SkipNone      SkipReason = ""
	SkipNoBet     SkipReason = "no_bet"
	SkipNotGraded SkipReason = "not_graded"
)

// ErrAlreadySettled is returned when the response was settled by someone else
// in the meantime. The journal entries written by this call must be rolled back.
var ErrAlreadySettled = errors.New("bet already settled")

// SettleInput is everything needed to settle one bet.
type SettleInput struct {
	Bet       Bet
	ProjectID string
	EventID   *string
	// Language is the user's language, used for the journal reasons.
	Language string
	// ChallengeName is already translated into Language.
	ChallengeName string
}

// Result is the outcome of settling one bet.
// When Settled is false, SkipReason says why and nothing was written.
type Result struct {
	Settled    bool
	SkipReason SkipReason
	Correct    bool
	Multiplier int64
	Stake      int
	Winnings   int
	NetPoints  int
	JournalID  string
}

// Settle pays out a bet and records the result.
//
// It writes two score journal entries, the stake (negative) and the winnings
// (zero or positive), then stores the net points and journal ID on the
// response, which only succeeds while the response has no score_journal_id.
// Settle must run in a transaction: on ErrAlreadySettled the caller rolls back
// so concurrent settlements of the same response pay out only once.
func Settle(ctx context.Context, q Querier, in SettleInput) (Result, error) {
	bet := in.Bet

	if bet.Amount <= 0 {
		return Result{SkipReason: SkipNoBet}, nil
	}
	if bet.IsCorrect == nil {
		slog.Warn("betting: response is not graded, bet left unsettled",
			"response_id", bet.ResponseID)
		return Result{SkipReason: SkipNotGraded}, nil
	}

	correct := *bet.IsCorrect
	stake := bet.Amount
	multiplier := Multiplier(correct, bet.MultiplierCorrect, bet.MultiplierWrong)
	winnings, err := Winnings(stake, multiplier)
	if err != nil {
		return Result{}, fmt.Errorf("stake %d, multiplier %d: %w", stake, multiplier, err)
	}
	// Both are within 0..MaxInt32, so the difference fits in int32
	netPoints := winnings - stake

	stakeJournalID := ulid.NewScoreJournalID()
	winningsJournalID := ulid.NewScoreJournalID()
	journalID := stakeJournalID
	if winnings > 0 {
		journalID = winningsJournalID
	}

	stakeReason := i18n.FormatBetStakeReason(in.Language, in.ChallengeName)
	_, err = q.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
		ID:         stakeJournalID,
		ProjectID:  in.ProjectID,
		UserID:     bet.UserID,
		EventID:    in.EventID,
		Points:     -stake,
		SourceType: SourceTypeBet,
		SourceID:   &bet.ResponseID,
		Reason:     &stakeReason,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to create stake journal entry: %w", err)
	}

	// Always created, even when 0, so every bet has a matching winnings entry
	winningsReason := i18n.FormatBetWinningsReason(in.Language, in.ChallengeName)
	_, err = q.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
		ID:         winningsJournalID,
		ProjectID:  in.ProjectID,
		UserID:     bet.UserID,
		EventID:    in.EventID,
		Points:     winnings,
		SourceType: SourceTypeBet,
		SourceID:   &bet.ResponseID,
		Reason:     &winningsReason,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to create winnings journal entry: %w", err)
	}

	// Stored last: score_journal_id references the entries written above
	claimed, err := q.SettleBetResult(ctx, sqlc.SettleBetResultParams{
		ID:             bet.ResponseID,
		Pointsearned:   netPoints,
		Scorejournalid: journalID,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to store bet result: %w", err)
	}
	if claimed == 0 {
		return Result{}, ErrAlreadySettled
	}

	return Result{
		Settled:    true,
		Correct:    correct,
		Multiplier: multiplier,
		Stake:      int(stake),
		Winnings:   int(winnings),
		NetPoints:  int(netPoints),
		JournalID:  journalID,
	}, nil
}
