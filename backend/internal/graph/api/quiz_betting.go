package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/betting"
	"github.com/jackc/pgx/v5/pgtype"
)

// BetValidationError represents an error during bet validation
type BetValidationError struct {
	Field   string
	Message string
}

func (e *BetValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// BetValidationConfig contains the betting configuration for a question
type BetValidationConfig struct {
	BettingEnabled       bool
	BettingMinPercentage pgtype.Numeric
	BettingMaxPercentage pgtype.Numeric
	BettingMinAbsolute   *int32
	BettingMaxAbsolute   *int32
}

// ValidateBet validates a bet amount against the question's betting
// configuration and the points the user can still bet (see availableBetPoints).
//
// Returns nil if the bet is valid, or a BetValidationError if invalid.
//
// Validation rules:
// - If betting is enabled, a bet must be provided (nil is rejected)
// - If betting is disabled and bet is nil, it's valid (no bet placed)
// - If betting is disabled on the question, any non-zero bet is rejected
// - Bet must be >= 0
// - Bet must not exceed the available points
// - If bettingMinAbsolute is set, bet must be >= that value
// - If bettingMaxAbsolute is set, bet must be <= that value
// - If bettingMinPercentage is set and available > 0, bet must be >= (available * minPercentage / 100)
// - If bettingMaxPercentage is set and available > 0, bet must be <= (available * maxPercentage / 100)
func ValidateBet(config BetValidationConfig, available int, betAmount *int) error {
	// If betting is enabled, a bet is required
	if config.BettingEnabled && betAmount == nil {
		return &BetValidationError{
			Field:   "betAmount",
			Message: "bet is required when betting is enabled",
		}
	}

	// No bet is valid when betting is not enabled
	if betAmount == nil {
		return nil
	}

	bet := *betAmount

	// Bet must be non-negative
	if bet < 0 {
		return &BetValidationError{
			Field:   "betAmount",
			Message: "bet amount cannot be negative",
		}
	}

	// At this point betting must be enabled (we checked above for nil/zero bets)
	// but we keep this check for safety in case the function is called directly
	if !config.BettingEnabled {
		return &BetValidationError{
			Field:   "betAmount",
			Message: "betting is not enabled for this question",
		}
	}

	// Bet cannot exceed what the user has left after their open bets
	if bet > available {
		return &BetValidationError{
			Field:   "betAmount",
			Message: fmt.Sprintf("bet amount (%d) exceeds available points (%d)", bet, available),
		}
	}

	// Validate against absolute limits
	if config.BettingMinAbsolute != nil {
		minAbs := int(*config.BettingMinAbsolute)
		if bet < minAbs {
			return &BetValidationError{
				Field:   "betAmount",
				Message: fmt.Sprintf("bet amount (%d) is below minimum (%d)", bet, minAbs),
			}
		}
	}

	if config.BettingMaxAbsolute != nil {
		maxAbs := int(*config.BettingMaxAbsolute)
		if bet > maxAbs {
			return &BetValidationError{
				Field:   "betAmount",
				Message: fmt.Sprintf("bet amount (%d) exceeds maximum (%d)", bet, maxAbs),
			}
		}
	}

	// Validate against percentage limits of the available points (only if available > 0)
	if available > 0 {
		if config.BettingMinPercentage.Valid {
			val, _ := config.BettingMinPercentage.Float64Value()
			minPct := val.Float64
			minAmount := int(float64(available) * minPct / 100)
			if bet < minAmount {
				return &BetValidationError{
					Field:   "betAmount",
					Message: fmt.Sprintf("bet amount (%d) is below minimum percentage (%.1f%% = %d)", bet, minPct, minAmount),
				}
			}
		}

		if config.BettingMaxPercentage.Valid {
			val, _ := config.BettingMaxPercentage.Float64Value()
			maxPct := val.Float64
			maxAmount := int(float64(available) * maxPct / 100)
			if bet > maxAmount {
				return &BetValidationError{
					Field:   "betAmount",
					Message: fmt.Sprintf("bet amount (%d) exceeds maximum percentage (%.1f%% = %d)", bet, maxPct, maxAmount),
				}
			}
		}
	}

	return nil
}

// availableBetPoints returns the points a user can still bet in a project:
// their score minus their open stakes, excluding the response being changed
// (nil for a new answer). Call it after betCheck.lock, in the transaction that
// stores the bet, so two concurrent bets cannot both spend the same points
func availableBetPoints(ctx context.Context, q *sqlc.Queries, userID, projectID string, excludeResponseID *string) (int, error) {
	available, err := q.GetUserAvailableBetPoints(ctx, sqlc.GetUserAvailableBetPointsParams{
		Userid:            userID,
		Projectid:         projectID,
		Excluderesponseid: excludeResponseID,
		Coresettledtypes:  sessionBetQuestionTypes,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get available bet points: %w", err)
	}
	return int(available), nil
}

// betCheck validates a bet inside the transaction that stores it. Call lock
// first, before any other row lock, then validate
type betCheck struct {
	userID    string
	projectID string
	// excludeResponseID is the response whose bet is being replaced, nil for a new answer
	excludeResponseID *string
	config            BetValidationConfig
	betAmount         *int
}

// newBetCheck returns a betCheck for a bet on a question of the given project.
func newBetCheck(userID, projectID string, excludeResponseID *string, config BetValidationConfig, betAmount *int) *betCheck {
	return &betCheck{
		userID:            userID,
		projectID:         projectID,
		excludeResponseID: excludeResponseID,
		config:            config,
		betAmount:         betAmount,
	}
}

// lock takes the row locks of all the user's submissions that can still get a
// bet in the project, so the user's concurrent bets run one after another.
// Other users are not affected.
func (b *betCheck) lock(ctx context.Context, q *sqlc.Queries) error {
	if _, err := q.LockUserBettableSubmissions(ctx, sqlc.LockUserBettableSubmissionsParams{
		Userid:    b.userID,
		Projectid: b.projectID,
	}); err != nil {
		return fmt.Errorf("failed to lock submissions: %w", err)
	}
	return nil
}

// validate checks the bet against the question limits and the points the user
// has left. Under READ COMMITTED each statement sees the latest committed data,
// so after lock it sees every bet the user placed before.
func (b *betCheck) validate(ctx context.Context, q *sqlc.Queries) error {
	// Only a bet that is validated against the points needs them
	available := 0
	if b.config.BettingEnabled && b.betAmount != nil && *b.betAmount >= 0 {
		var err error
		available, err = availableBetPoints(ctx, q, b.userID, b.projectID, b.excludeResponseID)
		if err != nil {
			return err
		}
	}
	return ValidateBet(b.config, available, b.betAmount)
}

// ExtractBetConfigFromQuestion extracts betting configuration from a question row
func ExtractBetConfigFromQuestion(row quizQuestionRow) BetValidationConfig {
	return BetValidationConfig{
		BettingEnabled:       row.GetBettingEnabled(),
		BettingMinPercentage: row.GetBettingMinPercentage(),
		BettingMaxPercentage: row.GetBettingMaxPercentage(),
		BettingMinAbsolute:   row.GetBettingMinAbsolute(),
		BettingMaxAbsolute:   row.GetBettingMaxAbsolute(),
	}
}

// ValidateBettingMultipliers checks the payout multipliers a question will have
// after a create or update: pass the new value if one is given, otherwise the
// stored one (nil = not set, the default applies). Each must be 0..100 with at
// most 2 decimals, and wrong must not exceed correct, defaults included.
func ValidateBettingMultipliers(correct, wrong *float64) error {
	correctFixed, err := multiplierToFixed("bettingMultiplierCorrect", correct)
	if err != nil {
		return err
	}
	wrongFixed, err := multiplierToFixed("bettingMultiplierWrong", wrong)
	if err != nil {
		return err
	}
	if betting.Multiplier(false, correctFixed, wrongFixed) > betting.Multiplier(true, correctFixed, wrongFixed) {
		return &BetValidationError{Field: "bettingMultiplierWrong", Message: "must not be greater than bettingMultiplierCorrect"}
	}
	return nil
}

// multiplierToFixed converts an optional API multiplier to hundredths.
func multiplierToFixed(field string, v *float64) (*int64, error) {
	if v == nil {
		return nil, nil
	}
	fixed, err := betting.MultiplierFromFloat(*v)
	switch {
	case errors.Is(err, betting.ErrMultiplierOutOfRange):
		return nil, &BetValidationError{Field: field, Message: fmt.Sprintf("must be between 0 and %d", betting.MaxMultiplier/betting.MultiplierScale)}
	case errors.Is(err, betting.ErrMultiplierPrecision):
		return nil, &BetValidationError{Field: field, Message: "must have at most 2 decimals"}
	case err != nil:
		return nil, &BetValidationError{Field: field, Message: err.Error()}
	}
	return &fixed, nil
}
