package betting

import (
	"context"
	"errors"
	"fmt"

	"github.com/bcc-media/wayfarer/i18n"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SkipReason explains why a bet was not settled.
type SkipReason string

const (
	SkipNone  SkipReason = ""
	SkipNoBet SkipReason = "no_bet"
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
	// Void means the response was not graded (no answer given): the stake was
	// returned, NetPoints is 0 and Correct and Multiplier carry no meaning.
	Void       bool
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
// A bet on a response that was not graded (no answer given) is void: the
// second entry returns the stake, so the net is 0.
// Settle must run in a transaction: on ErrAlreadySettled the caller rolls back
// so concurrent settlements of the same response pay out only once.
func Settle(ctx context.Context, q Querier, in SettleInput) (Result, error) {
	p, err := planBet(in)
	if err != nil || !p.result.Settled {
		return p.result, err
	}

	if _, err := q.CreateScoreJournalEntry(ctx, p.stake); err != nil {
		return Result{}, fmt.Errorf("failed to create stake journal entry: %w", err)
	}
	if _, err := q.CreateScoreJournalEntry(ctx, p.winnings); err != nil {
		return Result{}, fmt.Errorf("failed to create winnings journal entry: %w", err)
	}

	// Stored last: score_journal_id references the entries written above
	claimed, err := q.SettleBetResult(ctx, sqlc.SettleBetResultParams{
		ID:             in.Bet.ResponseID,
		Pointsearned:   int32(p.result.NetPoints),
		Scorejournalid: p.result.JournalID,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to store bet result: %w", err)
	}
	if claimed == 0 {
		return Result{}, ErrAlreadySettled
	}
	return p.result, nil
}

// betPlan is what settling one bet writes: the stake and winnings journal
// entries, and the result whose NetPoints and JournalID the response stores.
type betPlan struct {
	result   Result
	stake    sqlc.CreateScoreJournalEntryParams
	winnings sqlc.CreateScoreJournalEntryParams
}

// planBet computes the payout of a bet without touching the database, so
// Settle and SettleBatch pay out identically. A bet that is not settled
// (no stake) has result.Settled false and no entries.
func planBet(in SettleInput) (betPlan, error) {
	bet := in.Bet

	if bet.Amount <= 0 {
		return betPlan{result: Result{SkipReason: SkipNoBet}}, nil
	}
	void := bet.IsCorrect == nil
	correct := !void && *bet.IsCorrect
	stake := bet.Amount
	var multiplier int64
	var winnings int32
	if void {
		// Nothing to judge: the whole stake is returned
		winnings = stake
	} else {
		multiplier = Multiplier(correct, bet.MultiplierCorrect, bet.MultiplierWrong)
		var err error
		winnings, err = Winnings(stake, multiplier)
		if err != nil {
			return betPlan{}, fmt.Errorf("stake %d, multiplier %d: %w", stake, multiplier, err)
		}
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
	// Always created, even when 0, so every bet has a matching winnings entry
	winningsReason := i18n.FormatBetWinningsReason(in.Language, in.ChallengeName)
	if void {
		winningsReason = i18n.FormatBetRefundReason(in.Language, in.ChallengeName)
	}

	return betPlan{
		result: Result{
			Settled:    true,
			Void:       void,
			Correct:    correct,
			Multiplier: multiplier,
			Stake:      int(stake),
			Winnings:   int(winnings),
			NetPoints:  int(netPoints),
			JournalID:  journalID,
		},
		stake: sqlc.CreateScoreJournalEntryParams{
			ID:         stakeJournalID,
			ProjectID:  in.ProjectID,
			UserID:     bet.UserID,
			EventID:    in.EventID,
			Points:     -stake,
			SourceType: SourceTypeBet,
			SourceID:   &bet.ResponseID,
			Reason:     &stakeReason,
		},
		winnings: sqlc.CreateScoreJournalEntryParams{
			ID:         winningsJournalID,
			ProjectID:  in.ProjectID,
			UserID:     bet.UserID,
			EventID:    in.EventID,
			Points:     winnings,
			SourceType: SourceTypeBet,
			SourceID:   &bet.ResponseID,
			Reason:     &winningsReason,
		},
	}, nil
}

// BatchResult is the outcome of one bet in SettleBatch. Err is set when this
// bet could not be settled (e.g. ErrAlreadySettled, or a payout out of
// range); the rest of the batch is not affected.
type BatchResult struct {
	Result
	Err error
}

// SettleBatch settles many bets with three statements instead of three per
// bet: it claims the bets that are still unsettled (row locks, so concurrent
// settlements never pay a bet twice), writes all their journal entries, then
// stores all results. Payouts are computed exactly like Settle.
//
// It must run in a transaction (q bound to it), which the caller commits;
// an error means nothing of the batch may be committed. Results are in the
// order of ins. A bet another settlement got first has Err ErrAlreadySettled
// and nothing written for it.
func SettleBatch(ctx context.Context, q BatchQuerier, ins []SettleInput) ([]BatchResult, error) {
	results := make([]BatchResult, len(ins))
	plans := make([]betPlan, len(ins))
	var claimIDs []string
	for i, in := range ins {
		p, err := planBet(in)
		if err != nil {
			results[i].Err = err
			continue
		}
		plans[i] = p
		results[i].Result = p.result
		if p.result.Settled {
			claimIDs = append(claimIDs, in.Bet.ResponseID)
		}
	}
	if len(claimIDs) == 0 {
		return results, nil
	}

	claimedIDs, err := q.ClaimUnsettledBets(ctx, claimIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to claim bets: %w", err)
	}
	claimed := make(map[string]bool, len(claimedIDs))
	for _, id := range claimedIDs {
		claimed[id] = true
	}

	var entries sqlc.CreateScoreJournalEntriesParams
	var settled sqlc.SettleBetResultsParams
	for i, in := range ins {
		if !results[i].Settled {
			continue
		}
		if !claimed[in.Bet.ResponseID] {
			results[i] = BatchResult{Err: ErrAlreadySettled}
			continue
		}
		for _, e := range []sqlc.CreateScoreJournalEntryParams{plans[i].stake, plans[i].winnings} {
			entries.Ids = append(entries.Ids, e.ID)
			entries.ProjectIds = append(entries.ProjectIds, e.ProjectID)
			entries.UserIds = append(entries.UserIds, e.UserID)
			entries.EventIds = append(entries.EventIds, derefString(e.EventID))
			entries.Points = append(entries.Points, e.Points)
			entries.SourceTypes = append(entries.SourceTypes, e.SourceType)
			entries.SourceIds = append(entries.SourceIds, derefString(e.SourceID))
			entries.Reasons = append(entries.Reasons, derefString(e.Reason))
		}
		settled.Ids = append(settled.Ids, in.Bet.ResponseID)
		settled.PointsEarned = append(settled.PointsEarned, int32(results[i].NetPoints))
		settled.ScoreJournalIds = append(settled.ScoreJournalIds, results[i].JournalID)
	}
	if len(settled.Ids) == 0 {
		return results, nil
	}

	if err := q.CreateScoreJournalEntries(ctx, entries); err != nil {
		return nil, fmt.Errorf("failed to create journal entries: %w", err)
	}
	// Stored last: score_journal_id references the entries written above
	if err := q.SettleBetResults(ctx, settled); err != nil {
		return nil, fmt.Errorf("failed to store bet results: %w", err)
	}
	return results, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
