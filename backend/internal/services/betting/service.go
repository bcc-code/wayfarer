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

// SkipReason explains why a response was not settled.
type SkipReason string

const (
	SkipNone            SkipReason = ""
	SkipNoBet           SkipReason = "no_bet"
	SkipBettingDisabled SkipReason = "betting_disabled"
	SkipAlreadySettled  SkipReason = "already_settled"
	SkipUnsupportedType SkipReason = "unsupported_question_type"
	SkipNotEvaluable    SkipReason = "not_evaluable"
)

// SettleInput is everything needed to settle one response.
type SettleInput struct {
	Response  Response
	Answers   []Answer
	UserID    string
	ProjectID string
	EventID   *string
	// Language is the user's language, used for the journal reasons.
	Language string
	// ChallengeName is already translated into Language.
	ChallengeName string
}

// Result is the outcome of settling one response.
// When Settled is false, SkipReason says why and no data was written.
type Result struct {
	Settled    bool
	SkipReason SkipReason
	Outcome    Outcome
	Multiplier float64
	Stake      int
	Winnings   int
	NetPoints  int
	// JournalID is the journal entry stored on the response:
	// the winnings entry on a win, the stake entry otherwise.
	JournalID string
}

// Service settles bets using a Strategy per question type.
type Service struct {
	queries    Querier
	strategies map[QuestionType]Strategy
}

// NewService creates a betting service. Question types missing from
// strategies are skipped with SkipUnsupportedType.
func NewService(queries Querier, strategies map[QuestionType]Strategy) *Service {
	return &Service{
		queries:    queries,
		strategies: strategies,
	}
}

// Supports reports whether bets on the question type can be settled.
func (s *Service) Supports(questionType QuestionType) bool {
	_, ok := s.strategies[questionType]
	return ok
}

// Settle evaluates a response, pays out its bet and records the result.
//
// It writes two score journal entries: the stake (negative) and the winnings
// (zero or positive), then stores the net points and journal ID on the response.
// A response that already has a journal ID is never settled twice.
func (s *Service) Settle(ctx context.Context, in SettleInput) (Result, error) {
	resp := in.Response

	if resp.BetAmount == nil || *resp.BetAmount == 0 {
		return Result{SkipReason: SkipNoBet}, nil
	}
	if !resp.BettingEnabled {
		return Result{SkipReason: SkipBettingDisabled}, nil
	}
	if resp.ScoreJournalID != nil && *resp.ScoreJournalID != "" {
		return Result{SkipReason: SkipAlreadySettled}, nil
	}

	strategy, ok := s.strategies[resp.QuestionType]
	if !ok {
		slog.Warn("betting: no settlement strategy for question type, bet left unsettled",
			"response_id", resp.ID, "question_type", resp.QuestionType)
		return Result{SkipReason: SkipUnsupportedType}, nil
	}

	outcome, err := strategy.Evaluator.Evaluate(resp, in.Answers)
	if err != nil {
		if errors.Is(err, ErrNotEvaluable) {
			slog.Warn("betting: response could not be evaluated, bet left unsettled",
				"error", err, "response_id", resp.ID, "question_id", resp.QuestionID)
			return Result{SkipReason: SkipNotEvaluable}, nil
		}
		return Result{}, fmt.Errorf("failed to evaluate response: %w", err)
	}

	stake := int(*resp.BetAmount)
	multiplier := strategy.PayoutRule.Multiplier(outcome)
	winnings := int(float64(stake) * multiplier)
	netPoints := winnings - stake

	stakeJournalID := ulid.NewScoreJournalID()
	stakeReason := i18n.FormatBetStakeReason(in.Language, in.ChallengeName)
	_, err = s.queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
		ID:         stakeJournalID,
		ProjectID:  in.ProjectID,
		UserID:     in.UserID,
		EventID:    in.EventID,
		Points:     int32(-stake),
		SourceType: SourceTypeBet,
		SourceID:   &resp.ID,
		Reason:     &stakeReason,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to create stake journal entry: %w", err)
	}

	// Always created, even when 0, so every bet has a matching winnings entry
	winningsJournalID := ulid.NewScoreJournalID()
	winningsReason := i18n.FormatBetWinningsReason(in.Language, in.ChallengeName)
	_, err = s.queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
		ID:         winningsJournalID,
		ProjectID:  in.ProjectID,
		UserID:     in.UserID,
		EventID:    in.EventID,
		Points:     int32(winnings),
		SourceType: SourceTypeBet,
		SourceID:   &resp.ID,
		Reason:     &winningsReason,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to create winnings journal entry: %w", err)
	}

	journalID := stakeJournalID
	if winnings > 0 {
		journalID = winningsJournalID
	}

	_, err = s.queries.UpdateBetResultWithJournal(ctx, sqlc.UpdateBetResultWithJournalParams{
		ID:             resp.ID,
		Pointsearned:   int32(netPoints),
		Scorejournalid: journalID,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to store bet result: %w", err)
	}

	slog.Debug("betting: settled bet",
		"response_id", resp.ID,
		"user_id", in.UserID,
		"question_type", resp.QuestionType,
		"bet_amount", stake,
		"correct", outcome.Correct,
		"total", outcome.Total,
		"multiplier", multiplier,
		"winnings", winnings,
		"net_points", netPoints)

	return Result{
		Settled:    true,
		Outcome:    outcome,
		Multiplier: multiplier,
		Stake:      stake,
		Winnings:   winnings,
		NetPoints:  netPoints,
		JournalID:  journalID,
	}, nil
}
