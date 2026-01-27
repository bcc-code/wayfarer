package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
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
// - If bet is nil or 0, it's always valid (no bet placed)
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
	// No bet or zero bet is always valid
	if betAmount == nil || *betAmount == 0 {
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

	// Betting must be enabled on the question
	if !config.BettingEnabled {
		return &BetValidationError{
			Field:   "betAmount",
			Message: "betting is not enabled for this question",
		}
	}

	// Get user's current project score
	score, err := queries.GetUserScore(ctx, sqlc.GetUserScoreParams{
		UserID:    userID,
		ProjectID: projectID,
		EventID:   "", // Project-level score, not event-specific
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
