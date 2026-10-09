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

// fakeAvailableDB answers GetUserAvailableBetPoints (QueryRow) and records
// its arguments; LockUserBettableSubmissions (Query) returns lockErr.
type fakeAvailableDB struct {
	available int64
	err       error
	calls     int
	args      []interface{}
	lockErr   error
	lockArgs  []interface{}
}

func (f *fakeAvailableDB) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (f *fakeAvailableDB) Query(_ context.Context, _ string, args ...interface{}) (pgx.Rows, error) {
	f.lockArgs = args
	return nil, f.lockErr
}

func (f *fakeAvailableDB) QueryRow(_ context.Context, _ string, args ...interface{}) pgx.Row {
	f.calls++
	f.args = args
	return availableRow{f}
}

type availableRow struct{ db *fakeAvailableDB }

func (r availableRow) Scan(dest ...any) error {
	if r.db.err != nil {
		return r.db.err
	}
	*dest[0].(*int64) = r.db.available
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
		available  int
		config     BetValidationConfig
		bet        *int
		wantErr    string // substring; empty = valid unless wantAnyErr
		wantAnyErr bool   // any validation error, message not checked
	}{
		// --- betting enabled / disabled
		{name: "enabled, no bet", config: BetValidationConfig{BettingEnabled: true}, wantErr: "bet is required"},
		{name: "disabled, no bet", config: BetValidationConfig{}},
		{name: "disabled, bet", available: 100, config: BetValidationConfig{}, bet: bet(10), wantErr: "betting is not enabled"},
		{name: "disabled, zero bet means no bet", config: BetValidationConfig{}, bet: bet(0)},
		{name: "negative bet", config: BetValidationConfig{BettingEnabled: true}, bet: bet(-1), wantErr: "cannot be negative"},

		// --- zero bet
		{name: "R2: zero bet is rejected when a positive bet is possible", available: 100,
			config: BetValidationConfig{BettingEnabled: true}, bet: bet(0), wantAnyErr: true},
		{name: "R2: zero bet is accepted with zero score", available: 0,
			config: BetValidationConfig{BettingEnabled: true}, bet: bet(0)},
		{name: "zero bet below absolute minimum", available: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(1)}, bet: bet(0), wantErr: "below minimum (1)"},
		{name: "zero bet below percentage minimum", available: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("10")}, bet: bet(0), wantErr: "below minimum percentage"},

		// --- available points
		{name: "bet equals score", available: 100, config: BetValidationConfig{BettingEnabled: true}, bet: bet(100)},
		{name: "bet one above score", available: 100, config: BetValidationConfig{BettingEnabled: true}, bet: bet(101), wantErr: "exceeds available points (100)"},
		{name: "zero score", available: 0, config: BetValidationConfig{BettingEnabled: true}, bet: bet(1), wantErr: "exceeds available points (0)"},
		{name: "negative score", available: -200, config: BetValidationConfig{BettingEnabled: true}, bet: bet(1), wantErr: "exceeds available points (-200)"},

		// --- absolute limits
		{name: "at absolute minimum", available: 100, config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(20)}, bet: bet(20)},
		{name: "at absolute maximum", available: 100, config: BetValidationConfig{BettingEnabled: true, BettingMaxAbsolute: abs(30)}, bet: bet(30)},
		{name: "above absolute maximum", available: 100, config: BetValidationConfig{BettingEnabled: true, BettingMaxAbsolute: abs(30)}, bet: bet(31), wantErr: "exceeds maximum (30)"},
		{name: "absolute minimum above score: positive bet rejected", available: 50,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(100)}, bet: bet(50), wantErr: "below minimum (100)"},
		{name: "R2: absolute minimum above score: zero bet accepted", available: 50,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(100)}, bet: bet(0)},
		{name: "score is checked before the absolute minimum", available: 50,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(100)}, bet: bet(100), wantErr: "exceeds available points (50)"},
		{name: "absolute maximum above score: score wins", available: 50,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxAbsolute: abs(1000)}, bet: bet(51), wantErr: "exceeds available points (50)"},

		// --- percentage limits (amount = floor(score * pct / 100))
		{name: "min percentage rounds down: 10% of 99 is 9", available: 99,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("10")}, bet: bet(9)},
		{name: "below min percentage of 99", available: 99,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("10")}, bet: bet(8), wantErr: "below minimum percentage"},
		{name: "max percentage rounds down: 33.33% of 100 is 33", available: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("33.33")}, bet: bet(33)},
		{name: "above max percentage 33.33% of 100", available: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("33.33")}, bet: bet(34), wantErr: "exceeds maximum percentage"},
		{name: "max percentage of a small score rounds to 0", available: 5,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("10")}, bet: bet(1), wantErr: "exceeds maximum percentage"},
		{name: "R2: max percentage rounds to 0: zero bet accepted", available: 5,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("10")}, bet: bet(0)},
		{name: "100% max percentage allows the whole score", available: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMaxPercentage: pct("100")}, bet: bet(100)},
		{name: "percentage limits are skipped at zero score", available: 0,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("50")}, bet: bet(0)},
		{name: "R2: negative score: zero bet accepted", available: -10,
			config: BetValidationConfig{BettingEnabled: true}, bet: bet(0)},
		{name: "min percentage above max percentage: positive bet rejected", available: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("60"), BettingMaxPercentage: pct("40")}, bet: bet(50), wantErr: "below minimum percentage"},
		{name: "R2: min percentage above max percentage: zero bet accepted", available: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinPercentage: pct("60"), BettingMaxPercentage: pct("40")}, bet: bet(0)},

		// --- combined
		{name: "absolute and percentage both apply, absolute first", available: 100,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(15), BettingMinPercentage: pct("10")}, bet: bet(12), wantErr: "below minimum (15)"},
		{name: "absolute ok, percentage fails", available: 1000,
			config: BetValidationConfig{BettingEnabled: true, BettingMinAbsolute: abs(15), BettingMinPercentage: pct("10")}, bet: bet(50), wantErr: "below minimum percentage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBet(tt.config, tt.available, tt.bet)
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

func TestNewBetCheck(t *testing.T) {
	enabled := BetValidationConfig{BettingEnabled: true}
	bet := func(v int) *int { return &v }
	responseID := "QR01ARZ3NDEKTSV4RRFFQ69G5FAV"

	t.Run("checks the bet against the available points", func(t *testing.T) {
		db := &fakeAvailableDB{available: 100}
		check := newBetCheck("user1", "project1", nil, enabled, bet(100))
		require.NoError(t, check.validate(context.Background(), sqlc.New(db)))

		db = &fakeAvailableDB{available: 40}
		check = newBetCheck("user1", "project1", nil, enabled, bet(100))
		err := check.validate(context.Background(), sqlc.New(db))
		var betErr *BetValidationError
		require.ErrorAs(t, err, &betErr)
		assert.Contains(t, betErr.Message, "exceeds available points (40)")
	})

	t.Run("passes user, project, excluded response and core-settled types", func(t *testing.T) {
		db := &fakeAvailableDB{available: 100}
		check := newBetCheck("user1", "project1", &responseID, enabled, bet(10))
		require.NoError(t, check.validate(context.Background(), sqlc.New(db)))
		require.Equal(t, 1, db.calls)
		require.Len(t, db.args, 4)
		assert.Equal(t, "user1", db.args[0])
		assert.Equal(t, "project1", db.args[1])
		assert.Equal(t, &responseID, db.args[2])
		assert.Equal(t, sessionBetQuestionTypes, db.args[3])
	})

	t.Run("no lookup when the bet is rejected without it", func(t *testing.T) {
		for name, tc := range map[string]struct {
			config BetValidationConfig
			bet    *int
		}{
			"missing bet":      {config: enabled, bet: nil},
			"negative bet":     {config: enabled, bet: bet(-1)},
			"betting disabled": {config: BetValidationConfig{}, bet: bet(10)},
		} {
			t.Run(name, func(t *testing.T) {
				db := &fakeAvailableDB{}
				err := newBetCheck("user1", "project1", nil, tc.config, tc.bet).validate(context.Background(), sqlc.New(db))
				var betErr *BetValidationError
				require.ErrorAs(t, err, &betErr)
				assert.Zero(t, db.calls)
			})
		}
	})

	t.Run("a database error is not a validation error", func(t *testing.T) {
		dbErr := errors.New("db down")
		err := newBetCheck("user1", "project1", nil, enabled, bet(10)).validate(context.Background(), sqlc.New(&fakeAvailableDB{err: dbErr}))
		require.ErrorIs(t, err, dbErr)
		var betErr *BetValidationError
		assert.False(t, errors.As(err, &betErr))
	})

	t.Run("lock locks the user's submissions in the project", func(t *testing.T) {
		db := &fakeAvailableDB{lockErr: errors.New("lock timeout")}
		err := newBetCheck("user1", "project1", nil, enabled, bet(10)).lock(context.Background(), sqlc.New(db))
		require.ErrorIs(t, err, db.lockErr)
		assert.Contains(t, err.Error(), "failed to lock submissions")
		assert.Equal(t, []interface{}{"user1", "project1"}, db.lockArgs)
	})
}
