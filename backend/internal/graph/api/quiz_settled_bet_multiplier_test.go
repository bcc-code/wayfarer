package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettledBetMultiplier(t *testing.T) {
	journalID := "SJ01ARZ3NDEKTSV4RRFFQ69G5FAV"
	ptr := func(v int) *int { return &v }

	tests := []struct {
		name           string
		scoreJournalID *string
		betAmount      *int
		pointsEarned   *int
		want           *float64
	}{
		{
			name:           "doubled stake is 2x",
			scoreJournalID: &journalID,
			betAmount:      ptr(50),
			pointsEarned:   ptr(50),
			want:           f(2),
		},
		{
			name:           "lost stake is 0x",
			scoreJournalID: &journalID,
			betAmount:      ptr(50),
			pointsEarned:   ptr(-50),
			want:           f(0),
		},
		{
			name:           "a caller-supplied recordBetResult payout is reported as-is",
			scoreJournalID: &journalID,
			betAmount:      ptr(40),
			pointsEarned:   ptr(10),
			want:           f(1.25),
		},
		{
			name:           "breaking even is 1x, not a loss",
			scoreJournalID: &journalID,
			betAmount:      ptr(30),
			pointsEarned:   ptr(0),
			want:           f(1),
		},
		{
			name:           "a floored payout reports what the ledger actually paid",
			scoreJournalID: &journalID,
			betAmount:      ptr(3),
			pointsEarned:   ptr(4), // floor(3 * 2.5) = 7 winnings, net 4
			want:           f(7.0 / 3.0),
		},
		{
			name:           "an unsettled bet has no multiplier",
			scoreJournalID: nil,
			betAmount:      ptr(50),
			pointsEarned:   ptr(50),
			want:           nil,
		},
		{
			name:           "an empty journal id counts as unsettled",
			scoreJournalID: strPtr(""),
			betAmount:      ptr(50),
			pointsEarned:   ptr(50),
			want:           nil,
		},
		{
			name:           "a settled response with no points yet has no multiplier",
			scoreJournalID: &journalID,
			betAmount:      ptr(50),
			pointsEarned:   nil,
			want:           nil,
		},
		{
			name:           "no bet means no multiplier",
			scoreJournalID: &journalID,
			betAmount:      nil,
			pointsEarned:   ptr(10),
			want:           nil,
		},
		{
			name:           "a zero bet is not a bet, so it never divides by zero",
			scoreJournalID: &journalID,
			betAmount:      ptr(0),
			pointsEarned:   ptr(0),
			want:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SettledBetMultiplier(tt.scoreJournalID, tt.betAmount, tt.pointsEarned)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.InDelta(t, *tt.want, *got, 1e-9)
		})
	}
}

func f(v float64) *float64 { return &v }
