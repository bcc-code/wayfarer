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

// ValidateBet validates a bet amount against the question's betting configuration
// and the user's current project score.
//
// Returns nil if the bet is valid, or a BetValidationError if invalid.
//
// Validation rules:
// - If betting is enabled, a bet must be provided (nil or 0 is rejected)
// - If betting is disabled and bet is nil or 0, it's valid (no bet placed)
// - If betting is disabled on the question, any non-zero bet is rejected
// - Bet must be >= 0
// - Bet must not exceed user's current score
// - If bettingMinAbsolute is set, bet must be >= that value
// - If bettingMaxAbsolute is set, bet must be <= that value
// - If bettingMinPercentage is set and score > 0, bet must be >= (score * minPercentage / 100)
// - If bettingMaxPercentage is set and score > 0, bet must be <= (score * maxPercentage / 100)
func ValidateBet(
	ctx context.Context,
	queries *sqlc.Queries,
	userID string,
	projectID string,
	config BetValidationConfig,
	betAmount *int,
) error {
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

	// Get user's current project score
	score, err := queries.GetUserProjectScore(ctx, sqlc.GetUserProjectScoreParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("failed to get user score: %w", err)
	}

	currentScore := int(score)

	// Bet cannot exceed current score
	if bet > currentScore {
		return &BetValidationError{
			Field:   "betAmount",
			Message: fmt.Sprintf("bet amount (%d) exceeds current score (%d)", bet, currentScore),
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

	// Validate against percentage limits (only if score > 0)
	if currentScore > 0 {
		if config.BettingMinPercentage.Valid {
			val, _ := config.BettingMinPercentage.Float64Value()
			minPct := val.Float64
			minAmount := int(float64(currentScore) * minPct / 100)
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
			maxAmount := int(float64(currentScore) * maxPct / 100)
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
