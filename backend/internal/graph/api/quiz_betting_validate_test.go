package api

import (
	"context"
	"errors"
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeScoreDB answers GetUserProjectScore, the only query ValidateBet runs,
// so the real ValidateBet can be tested without a database.
type fakeScoreDB struct {
	score int64
	err   error
}

func (f *fakeScoreDB) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (f *fakeScoreDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (f *fakeScoreDB) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return scoreRow{f}
}

type scoreRow struct{ db *fakeScoreDB }

func (r scoreRow) Scan(dest ...any) error {
	if r.db.err != nil {
		return r.db.err
	}
	*dest[0].(*int64) = r.db.score
	return nil
}

// TestValidateBet_Real runs the real ValidateBet (the older tests above use a
// copy of its logic) and asserts the intended rules, so failures point at
// bugs. Rule R2 (see e2e/quiz_betting_edge_cases_test.go): a bet is required
// and must be at least 1, unless no positive bet is allowed, then 0 is
// accepted so the question can still be answered. See notes/betting-edge-cases.md.
func TestValidateBet_Real(t *testing.T) {
	pct := func(s string) pgtype.Numeric {
		var n pgtype.Numeric
		require.NoError(t, n.Scan(s))
		return n
	}
	abs := func(v int32) *int32 { return &v }
	bet := func(v int) *int { return &v }

	tests := []struct {
		name       string
		score      int64
		config     BetValidationConfig
		bet        *int
		wantErr    string // substring; empty = valid unless wantAnyErr
		wantAnyErr bool   // any validation error, message not checked
	}{
		// --- betting enabled / disabled
		{name: "enabled, no bet", config: BetValidationConfig{BettingEnabled: true}, wantErr: "bet is required"},
		{name: "disabled, no bet", config: BetValidationConfig{}},
		{name: "disabled, bet", score: 100, config: BetValidationConfig{}, bet: bet(10), wantErr: "betting is not enabled"},
		{name: "disabled, zero bet means no bet", config: BetValidationConfig{}, bet: bet(0)},
		{name: "negative bet", config: BetValidationConfig{BettingEnabled: true}, bet: bet(-1), wantErr: "cannot be negative"},

		// --- zero bet
		{name: "R2: zero bet is rejected when a positive bet is possible", score: 100,
			config: BetValidationConfig{BettingEnabled: true}, bet: bet(0), wantAnyErr: true},
		{name: "R2: zero bet is accepted with zero score", score: 0,
			config: BetValidationConfig{BettingEnabled: true}, bet: bet(0)},
		{name: "zero bet below absolute minimum", score: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(1)}, bet: bet(0), wantErr: "below minimum (1)"},
		{name: "zero bet below percentage minimum", score: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("10")}, bet: bet(0), wantErr: "below minimum percentage"},

		// --- current score
		{name: "bet equals score", score: 100, config: BetValidationConfig{BettingEnabled: true}, bet: bet(100)},
		{name: "bet one above score", score: 100, config: BetValidationConfig{BettingEnabled: true}, bet: bet(101), wantErr: "exceeds current score (100)"},
		{name: "zero score", score: 0, config: BetValidationConfig{BettingEnabled: true}, bet: bet(1), wantErr: "exceeds current score (0)"},
		{name: "negative score", score: -200, config: BetValidationConfig{BettingEnabled: true}, bet: bet(1), wantErr: "exceeds current score (-200)"},

		// --- absolute limits
		{name: "at absolute minimum", score: 100, config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(20)}, bet: bet(20)},
		{name: "at absolute maximum", score: 100, config: BetValidationConfig{BettingEnabled: true, BettingMaxAbsolute: abs(30)}, bet: bet(30)},
		{name: "above absolute maximum", score: 100, config: BetValidationConfig{BettingEnabled: true, BettingMaxAbsolute: abs(30)}, bet: bet(31), wantErr: "exceeds maximum (30)"},
		{name: "absolute minimum above score: positive bet rejected", score: 50,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(100)}, bet: bet(50), wantErr: "below minimum (100)"},
		{name: "R2: absolute minimum above score: zero bet accepted", score: 50,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(100)}, bet: bet(0)},
		{name: "score is checked before the absolute minimum", score: 50,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(100)}, bet: bet(100), wantErr: "exceeds current score (50)"},
		{name: "absolute maximum above score: score wins", score: 50,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxAbsolute: abs(1000)}, bet: bet(51), wantErr: "exceeds current score (50)"},

		// --- percentage limits (amount = floor(score * pct / 100))
		{name: "min percentage rounds down: 10% of 99 is 9", score: 99,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("10")}, bet: bet(9)},
		{name: "below min percentage of 99", score: 99,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("10")}, bet: bet(8), wantErr: "below minimum percentage"},
		{name: "max percentage rounds down: 33.33% of 100 is 33", score: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("33.33")}, bet: bet(33)},
		{name: "above max percentage 33.33% of 100", score: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("33.33")}, bet: bet(34), wantErr: "exceeds maximum percentage"},
		{name: "max percentage of a small score rounds to 0", score: 5,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("10")}, bet: bet(1), wantErr: "exceeds maximum percentage"},
		{name: "R2: max percentage rounds to 0: zero bet accepted", score: 5,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("10")}, bet: bet(0)},
		{name: "100% max percentage allows the whole score", score: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("100")}, bet: bet(100)},
		{name: "percentage limits are skipped at zero score", score: 0,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("50")}, bet: bet(0)},
		{name: "R2: negative score: zero bet accepted", score: -10,
			config: BetValidationConfig{BettingEnabled: true}, bet: bet(0)},
		{name: "min percentage above max percentage: positive bet rejected", score: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("60"), BettingMaxPercentage: pct("40")}, bet: bet(50), wantErr: "below minimum percentage"},
		{name: "R2: min percentage above max percentage: zero bet accepted", score: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("60"), BettingMaxPercentage: pct("40")}, bet: bet(0)},

		// --- combined
		{name: "absolute and percentage both apply, absolute first", score: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(15), BettingMinPercentage: pct("10")}, bet: bet(12), wantErr: "below minimum (15)"},
		{name: "absolute ok, percentage fails", score: 1000,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(15), BettingMinPercentage: pct("10")}, bet: bet(50), wantErr: "below minimum percentage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeScoreDB{score: tt.score}
			err := ValidateBet(context.Background(), sqlc.New(db), "user1", "project1", tt.config, tt.bet)
			if tt.wantErr == "" && !tt.wantAnyErr {
				assert.NoError(t, err)
				return
			}
			var betErr *BetValidationError
			require.ErrorAs(t, err, &betErr)
			assert.Equal(t, "betAmount", betErr.Field)
			assert.Contains(t, betErr.Message, tt.wantErr)
		})
	}
}

func TestValidateBet_Real_ScoreLookupFails(t *testing.T) {
	dbErr := errors.New("db down")
	bet := 10
	err := ValidateBet(context.Background(), sqlc.New(&fakeScoreDB{err: dbErr}), "user1", "project1",
		BetValidationConfig{BettingEnabled: true}, &bet)
	require.ErrorIs(t, err, dbErr)
	var betErr *BetValidationError
	assert.False(t, errors.As(err, &betErr), "a database error is not a validation error")
}
