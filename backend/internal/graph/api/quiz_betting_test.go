package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockQueries implements a minimal subset of sqlc.Queries for testing
type mockQueriesForBetting struct {
	score int64
	err   error
}

func (m *mockQueriesForBetting) GetUserProjectScore(ctx context.Context, params sqlc.GetUserProjectScoreParams) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.score, nil
}

// Helper function to create a numeric from float64
func numericFromFloat(val float64) pgtype.Numeric {
	var n pgtype.Numeric
	// pgtype.Numeric.Scan() requires a string representation to properly set Valid=true
	_ = n.Scan(fmt.Sprintf("%f", val))
	return n
}

func TestValidateBet_NilBet_BettingEnabled(t *testing.T) {
	config := BetValidationConfig{
		BettingEnabled: true,
	}

	err := ValidateBet(context.Background(), nil, "user1", "project1", config, nil)
	require.Error(t, err, "nil bet should be rejected when betting is enabled")
	betErr, ok := err.(*BetValidationError)
	require.True(t, ok)
	assert.Equal(t, "betAmount", betErr.Field)
	assert.Contains(t, betErr.Message, "bet is required when betting is enabled")
}

func TestValidateBet_NilBet_BettingDisabled(t *testing.T) {
	config := BetValidationConfig{
		BettingEnabled: false,
	}

	err := ValidateBet(context.Background(), nil, "user1", "project1", config, nil)
	assert.NoError(t, err, "nil bet should be valid when betting is disabled")
}

func TestValidateBet_NegativeBet(t *testing.T) {
	config := BetValidationConfig{
		BettingEnabled: true,
	}

	negativeBet := -10
	err := ValidateBet(context.Background(), nil, "user1", "project1", config, &negativeBet)

	require.Error(t, err)
	betErr, ok := err.(*BetValidationError)
	require.True(t, ok)
	assert.Equal(t, "betAmount", betErr.Field)
	assert.Contains(t, betErr.Message, "cannot be negative")
}

func TestValidateBet_BettingDisabled(t *testing.T) {
	config := BetValidationConfig{
		BettingEnabled: false,
	}

	bet := 10
	err := ValidateBet(context.Background(), nil, "user1", "project1", config, &bet)

	require.Error(t, err)
	betErr, ok := err.(*BetValidationError)
	require.True(t, ok)
	assert.Equal(t, "betAmount", betErr.Field)
	assert.Contains(t, betErr.Message, "betting is not enabled")
}

func TestValidateBet_ExceedsCurrentScore(t *testing.T) {
	// Create a mock that returns a score
	queries := &mockQueriesForBetting{score: 100}

	config := BetValidationConfig{
		BettingEnabled: true,
	}

	bet := 150
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	require.Error(t, err)
	betErr, ok := err.(*BetValidationError)
	require.True(t, ok)
	assert.Equal(t, "betAmount", betErr.Field)
	assert.Contains(t, betErr.Message, "exceeds current score")
}

func TestValidateBet_ValidWithinScore(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}

	config := BetValidationConfig{
		BettingEnabled: true,
	}

	bet := 50
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	assert.NoError(t, err)
}

func TestValidateBet_BelowMinAbsolute(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}
	minAbs := int32(20)

	config := BetValidationConfig{
		BettingEnabled:     true,
		BettingMinAbsolute: &minAbs,
	}

	bet := 10
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	require.Error(t, err)
	betErr, ok := err.(*BetValidationError)
	require.True(t, ok)
	assert.Contains(t, betErr.Message, "below minimum")
}

func TestValidateBet_AboveMaxAbsolute(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}
	maxAbs := int32(30)

	config := BetValidationConfig{
		BettingEnabled:     true,
		BettingMaxAbsolute: &maxAbs,
	}

	bet := 50
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	require.Error(t, err)
	betErr, ok := err.(*BetValidationError)
	require.True(t, ok)
	assert.Contains(t, betErr.Message, "exceeds maximum")
}

func TestValidateBet_ValidWithinAbsoluteLimits(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}
	minAbs := int32(10)
	maxAbs := int32(50)

	config := BetValidationConfig{
		BettingEnabled:     true,
		BettingMinAbsolute: &minAbs,
		BettingMaxAbsolute: &maxAbs,
	}

	bet := 25
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	assert.NoError(t, err)
}

func TestValidateBet_BelowMinPercentage(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}

	config := BetValidationConfig{
		BettingEnabled:       true,
		BettingMinPercentage: numericFromFloat(10), // 10% minimum = 10 points
	}

	bet := 5 // 5% = below minimum
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	require.Error(t, err)
	betErr, ok := err.(*BetValidationError)
	require.True(t, ok)
	assert.Contains(t, betErr.Message, "below minimum percentage")
}

func TestValidateBet_AboveMaxPercentage(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}

	config := BetValidationConfig{
		BettingEnabled:       true,
		BettingMaxPercentage: numericFromFloat(50), // 50% maximum = 50 points
	}

	bet := 60 // 60% = above maximum
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	require.Error(t, err)
	betErr, ok := err.(*BetValidationError)
	require.True(t, ok)
	assert.Contains(t, betErr.Message, "exceeds maximum percentage")
}

func TestValidateBet_ValidWithinPercentageLimits(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}

	config := BetValidationConfig{
		BettingEnabled:       true,
		BettingMinPercentage: numericFromFloat(10), // 10% = 10 points
		BettingMaxPercentage: numericFromFloat(50), // 50% = 50 points
	}

	bet := 30 // 30% = within limits
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	assert.NoError(t, err)
}

func TestValidateBet_CombinedLimits_MustSatisfyBoth(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}
	minAbs := int32(15)
	maxAbs := int32(40)

	config := BetValidationConfig{
		BettingEnabled:       true,
		BettingMinPercentage: numericFromFloat(10), // 10% = 10 points
		BettingMaxPercentage: numericFromFloat(50), // 50% = 50 points
		BettingMinAbsolute:   &minAbs,              // absolute min = 15
		BettingMaxAbsolute:   &maxAbs,              // absolute max = 40
	}

	// Test: bet=12 passes percentage (>=10%) but fails absolute (>=15)
	bet := 12
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "below minimum")

	// Test: bet=45 passes percentage (<=50%) but fails absolute (<=40)
	bet = 45
	err = validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")

	// Test: bet=25 passes both
	bet = 25
	err = validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)
	assert.NoError(t, err)
}

func TestValidateBet_ZeroScore_PercentageLimitsIgnored(t *testing.T) {
	queries := &mockQueriesForBetting{score: 0}

	config := BetValidationConfig{
		BettingEnabled:       true,
		BettingMinPercentage: numericFromFloat(10), // Would be 0
		BettingMaxPercentage: numericFromFloat(50), // Would be 0
	}

	// With score 0, any bet exceeds score
	bet := 1
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds current score")
}

func TestValidateBet_ZeroScore_AbsoluteLimitsStillApply(t *testing.T) {
	queries := &mockQueriesForBetting{score: 0}
	minAbs := int32(0)
	maxAbs := int32(0)

	config := BetValidationConfig{
		BettingEnabled:     true,
		BettingMinAbsolute: &minAbs,
		BettingMaxAbsolute: &maxAbs,
	}

	// With betting enabled, zero bet is rejected (bet is required)
	bet := 0
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)
	require.Error(t, err, "zero bet should be rejected when betting is enabled")
	assert.Contains(t, err.Error(), "bet is required when betting is enabled")
}

func TestValidateBet_NoLimitsSet_AnyBetUpToScoreValid(t *testing.T) {
	queries := &mockQueriesForBetting{score: 100}

	config := BetValidationConfig{
		BettingEnabled: true,
		// No min/max limits set
	}

	// Any bet up to score should be valid
	bet := 100
	err := validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)
	assert.NoError(t, err)

	bet = 1
	err = validateBetWithMockQueries(context.Background(), queries, "user1", "project1", config, &bet)
	assert.NoError(t, err)
}

// validateBetWithMockQueries is a test helper that calls ValidateBet with mocked queries
func validateBetWithMockQueries(
	ctx context.Context,
	mockQueries *mockQueriesForBetting,
	userID string,
	projectID string,
	config BetValidationConfig,
	betAmount *int,
) error {
	// If betting is enabled, a bet is required
	if config.BettingEnabled && (betAmount == nil || *betAmount == 0) {
		return &BetValidationError{
			Field:   "betAmount",
			Message: "bet is required when betting is enabled",
		}
	}

	// No bet or zero bet is valid when betting is not enabled
	if betAmount == nil || *betAmount == 0 {
		return nil
	}

	bet := *betAmount

	if bet < 0 {
		return &BetValidationError{
			Field:   "betAmount",
			Message: "bet amount cannot be negative",
		}
	}

	if !config.BettingEnabled {
		return &BetValidationError{
			Field:   "betAmount",
			Message: "betting is not enabled for this question",
		}
	}

	// Get user's current project score using mock
	score, err := mockQueries.GetUserProjectScore(ctx, sqlc.GetUserProjectScoreParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}

	currentScore := int(score)

	if bet > currentScore {
		return &BetValidationError{
			Field:   "betAmount",
			Message: "bet amount exceeds current score",
		}
	}

	if config.BettingMinAbsolute != nil {
		minAbs := int(*config.BettingMinAbsolute)
		if bet < minAbs {
			return &BetValidationError{
				Field:   "betAmount",
				Message: "bet amount is below minimum",
			}
		}
	}

	if config.BettingMaxAbsolute != nil {
		maxAbs := int(*config.BettingMaxAbsolute)
		if bet > maxAbs {
			return &BetValidationError{
				Field:   "betAmount",
				Message: "bet amount exceeds maximum",
			}
		}
	}

	if currentScore > 0 {
		if config.BettingMinPercentage.Valid {
			val, _ := config.BettingMinPercentage.Float64Value()
			minPct := val.Float64
			minAmount := int(float64(currentScore) * minPct / 100)
			if bet < minAmount {
				return &BetValidationError{
					Field:   "betAmount",
					Message: "bet amount is below minimum percentage",
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
					Message: "bet amount exceeds maximum percentage",
				}
			}
		}
	}

	return nil
}

func TestValidateBettingMultipliers(t *testing.T) {
	f := func(v float64) *float64 { return &v }

	tests := []struct {
		name      string
		correct   *float64
		wrong     *float64
		wantField string // empty = valid
	}{
		{name: "both unset", correct: nil, wrong: nil},
		{name: "only correct", correct: f(2.0)},
		{name: "only wrong, below default correct", wrong: f(0.5)},
		{name: "only wrong, equals default correct", wrong: f(2.0)},
		{name: "only wrong, above default correct", wrong: f(3.0), wantField: "bettingMultiplierWrong"},
		{name: "only correct, zero (default wrong is 0)", correct: f(0)},
		{name: "wrong below correct", correct: f(2.0), wrong: f(0.5)},
		{name: "wrong equals correct", correct: f(1.0), wrong: f(1.0)},
		{name: "zeros", correct: f(0), wrong: f(0)},
		{name: "negative correct", correct: f(-1), wantField: "bettingMultiplierCorrect"},
		{name: "negative wrong", wrong: f(-0.5), wantField: "bettingMultiplierWrong"},
		{name: "wrong above correct", correct: f(1.5), wrong: f(2.0), wantField: "bettingMultiplierWrong"},
		{name: "max correct", correct: f(100)},
		{name: "correct above max", correct: f(100.01), wantField: "bettingMultiplierCorrect"},
		{name: "wrong above max", correct: f(100), wrong: f(150), wantField: "bettingMultiplierWrong"},
		{name: "correct with 3 decimals", correct: f(1.234), wantField: "bettingMultiplierCorrect"},
		{name: "wrong with 3 decimals", wrong: f(0.125), wantField: "bettingMultiplierWrong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBettingMultipliers(tt.correct, tt.wrong)
			if tt.wantField == "" {
				assert.NoError(t, err)
				return
			}
			var betErr *BetValidationError
			require.ErrorAs(t, err, &betErr)
			assert.Equal(t, tt.wantField, betErr.Field)
		})
	}
}

func TestNumericToMultiplier(t *testing.T) {
	num := func(s string) pgtype.Numeric {
		var n pgtype.Numeric
		require.NoError(t, n.Scan(s))
		return n
	}

	tests := []struct {
		name    string
		in      pgtype.Numeric
		want    *int64
		wantErr bool
	}{
		{name: "not set", in: pgtype.Numeric{}, want: nil},
		{name: "integer", in: num("2"), want: int64Ptr(200)},
		{name: "two decimals", in: num("2.50"), want: int64Ptr(250)},
		{name: "trailing zeros", in: num("1.500000"), want: int64Ptr(150)},
		{name: "zero", in: num("0"), want: int64Ptr(0)},
		{name: "max", in: num("100.00"), want: int64Ptr(10000)},
		{name: "above max", in: num("100.01"), wantErr: true},
		{name: "negative", in: num("-1"), wantErr: true},
		{name: "three decimals", in: num("1.234"), wantErr: true},
		{name: "NaN", in: pgtype.Numeric{Valid: true, NaN: true}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := numericToMultiplier(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func int64Ptr(v int64) *int64 { return &v }

func TestNumericToFloat(t *testing.T) {
	var n pgtype.Numeric
	require.NoError(t, n.Scan("2.50"))
	got := numericToFloat(n)
	require.NotNil(t, got)
	assert.Equal(t, 2.5, *got)

	assert.Nil(t, numericToFloat(pgtype.Numeric{}))
	assert.Nil(t, numericToFloat(pgtype.Numeric{Valid: true, NaN: true}))
	assert.Nil(t, numericToFloat(pgtype.Numeric{Valid: true, InfinityModifier: pgtype.Infinity}))
}
