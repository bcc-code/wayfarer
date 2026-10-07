package betting

import (
	"context"
	"errors"
	"testing"

	"github.com/bcc-media/wayfarer/i18n"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/betting/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testResponseID = "QR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testUserID     = "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testProjectID  = "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testEventID    = "EV01ARZ3NDEKTSV4RRFFQ69G5FAV"
)

func boolPtr(v bool) *bool        { return &v }
func floatPtr(v float64) *float64 { return &v }
func testInput(amount int32, correct *bool) SettleInput {
	eventID := testEventID
	return SettleInput{
		Bet: Bet{
			ResponseID: testResponseID,
			UserID:     testUserID,
			Amount:     amount,
			IsCorrect:  correct,
		},
		ProjectID:     testProjectID,
		EventID:       &eventID,
		Language:      i18n.DefaultLanguage,
		ChallengeName: "Bible quiz",
	}
}

// journalEntry matches a bet journal entry with the given points.
func journalEntry(points int32) interface{} {
	return mock.MatchedBy(func(p sqlc.CreateScoreJournalEntryParams) bool {
		return p.Points == points &&
			len(p.ID) == 28 &&
			p.ProjectID == testProjectID &&
			p.UserID == testUserID &&
			p.EventID != nil && *p.EventID == testEventID &&
			p.ChallengeID == nil &&
			p.SourceType == SourceTypeBet &&
			p.SourceID != nil && *p.SourceID == testResponseID &&
			p.Reason != nil && *p.Reason != ""
	})
}

// expectJournalEntries expects the stake and winnings entries and records their IDs in order.
func expectJournalEntries(q *mocks.MockQuerier, ids *[]string, stake, winnings int32) {
	for _, pts := range []int32{-stake, winnings} {
		q.On("CreateScoreJournalEntry", mock.Anything, journalEntry(pts)).
			Run(func(args mock.Arguments) {
				*ids = append(*ids, args.Get(1).(sqlc.CreateScoreJournalEntryParams).ID)
			}).
			Return(&sqlc.ScoreJournal{}, nil).Once()
	}
}

func TestMultiplier(t *testing.T) {
	assert.Equal(t, DefaultMultiplierCorrect, Multiplier(true, nil, nil))
	assert.Equal(t, DefaultMultiplierWrong, Multiplier(false, nil, nil))
	assert.Equal(t, 3.0, Multiplier(true, floatPtr(3.0), floatPtr(0.5)))
	assert.Equal(t, 0.5, Multiplier(false, floatPtr(3.0), floatPtr(0.5)))
	assert.Equal(t, DefaultMultiplierCorrect, Multiplier(true, nil, floatPtr(0.5)), "only wrong set")
	assert.Equal(t, DefaultMultiplierWrong, Multiplier(false, floatPtr(3.0), nil), "only correct set")
}

func TestSettle_Skips(t *testing.T) {
	tests := []struct {
		name  string
		input SettleInput
		want  SkipReason
	}{
		{name: "zero bet", input: testInput(0, boolPtr(true)), want: SkipNoBet},
		{name: "negative bet", input: testInput(-5, boolPtr(true)), want: SkipNoBet},
		{name: "not graded", input: testInput(100, nil), want: SkipNotGraded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := mocks.NewMockQuerier(t) // no calls expected

			res, err := Settle(context.Background(), q, tt.input)
			require.NoError(t, err)
			assert.False(t, res.Settled)
			assert.Equal(t, tt.want, res.SkipReason)
		})
	}
}

func TestSettle_Payouts(t *testing.T) {
	tests := []struct {
		name          string
		amount        int32
		correct       bool
		multCorrect   *float64
		multWrong     *float64
		wantMult      float64
		wantWinnings  int
		wantNet       int
		wantWinningID bool // true: response stores the winnings entry, false: the stake entry
	}{
		{name: "correct, default", amount: 100, correct: true, wantMult: 2.0, wantWinnings: 200, wantNet: 100, wantWinningID: true},
		{name: "wrong, default", amount: 100, correct: false, wantMult: 0, wantWinnings: 0, wantNet: -100},
		{name: "correct, custom", amount: 100, correct: true, multCorrect: floatPtr(3.0), wantMult: 3.0, wantWinnings: 300, wantNet: 200, wantWinningID: true},
		{name: "wrong, half back", amount: 100, correct: false, multWrong: floatPtr(0.5), wantMult: 0.5, wantWinnings: 50, wantNet: -50, wantWinningID: true},
		{name: "correct, money back", amount: 100, correct: true, multCorrect: floatPtr(1.0), wantMult: 1.0, wantWinnings: 100, wantNet: 0, wantWinningID: true},
		{name: "winnings round down", amount: 3, correct: true, multCorrect: floatPtr(1.5), wantMult: 1.5, wantWinnings: 4, wantNet: 1, wantWinningID: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := mocks.NewMockQuerier(t)
			var ids []string
			expectJournalEntries(q, &ids, tt.amount, int32(tt.wantWinnings))

			var stored sqlc.SettleBetResultParams
			q.On("SettleBetResult", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					stored = args.Get(1).(sqlc.SettleBetResultParams)
				}).
				Return(int64(1), nil).Once()

			in := testInput(tt.amount, boolPtr(tt.correct))
			in.Bet.MultiplierCorrect = tt.multCorrect
			in.Bet.MultiplierWrong = tt.multWrong

			res, err := Settle(context.Background(), q, in)
			require.NoError(t, err)
			require.Len(t, ids, 2)
			stakeID, winningsID := ids[0], ids[1]

			assert.True(t, res.Settled)
			assert.Equal(t, SkipNone, res.SkipReason)
			assert.Equal(t, tt.correct, res.Correct)
			assert.Equal(t, tt.wantMult, res.Multiplier)
			assert.Equal(t, int(tt.amount), res.Stake)
			assert.Equal(t, tt.wantWinnings, res.Winnings)
			assert.Equal(t, tt.wantNet, res.NetPoints)

			wantJournalID := stakeID
			if tt.wantWinningID {
				wantJournalID = winningsID
			}
			assert.Equal(t, wantJournalID, res.JournalID)
			assert.Equal(t, sqlc.SettleBetResultParams{
				ID:             testResponseID,
				Pointsearned:   int32(tt.wantNet),
				Scorejournalid: wantJournalID,
			}, stored)
		})
	}
}

func TestSettle_AlreadySettled(t *testing.T) {
	q := mocks.NewMockQuerier(t)
	var ids []string
	expectJournalEntries(q, &ids, 100, 200)
	// Another settlement stored its result first: the caller must roll back
	q.On("SettleBetResult", mock.Anything, mock.Anything).Return(int64(0), nil).Once()

	res, err := Settle(context.Background(), q, testInput(100, boolPtr(true)))
	require.ErrorIs(t, err, ErrAlreadySettled)
	assert.False(t, res.Settled)
}

func TestSettle_DatabaseErrors(t *testing.T) {
	dbErr := errors.New("db down")

	t.Run("stake entry fails", func(t *testing.T) {
		q := mocks.NewMockQuerier(t)
		q.On("CreateScoreJournalEntry", mock.Anything, journalEntry(-100)).Return(nil, dbErr).Once()

		_, err := Settle(context.Background(), q, testInput(100, boolPtr(true)))
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("winnings entry fails", func(t *testing.T) {
		q := mocks.NewMockQuerier(t)
		q.On("CreateScoreJournalEntry", mock.Anything, journalEntry(-100)).Return(&sqlc.ScoreJournal{}, nil).Once()
		q.On("CreateScoreJournalEntry", mock.Anything, journalEntry(200)).Return(nil, dbErr).Once()

		_, err := Settle(context.Background(), q, testInput(100, boolPtr(true)))
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("storing result fails", func(t *testing.T) {
		q := mocks.NewMockQuerier(t)
		var ids []string
		expectJournalEntries(q, &ids, 100, 200)
		q.On("SettleBetResult", mock.Anything, mock.Anything).Return(int64(0), dbErr).Once()

		_, err := Settle(context.Background(), q, testInput(100, boolPtr(true)))
		require.ErrorIs(t, err, dbErr)
	})
}
