package betting

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/betting/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// batchInput is testInput with its own response ID, so a batch can hold many bets
func batchInput(n int, amount int32, correct *bool) SettleInput {
	in := testInput(amount, correct)
	in.Bet.ResponseID = fmt.Sprintf("QR%026d", n)
	return in
}

// claimAll makes ClaimUnsettledBets return every requested ID
func claimAll(q *mocks.MockBatchQuerier) {
	q.On("ClaimUnsettledBets", mock.Anything, mock.Anything).
		Return(func(_ context.Context, ids []string) ([]string, error) { return ids, nil }).Once()
}

// entry is one written journal entry, without its random ID
type entry struct {
	userID, eventID, sourceID, reason string
	points                            int32
}

// settleOne runs Settle on a recording mock and returns its entries and result
func settleOne(t *testing.T, in SettleInput) ([]entry, string, Result) {
	t.Helper()
	q := mocks.NewMockQuerier(t)
	var entries []entry
	ids := map[string]int32{}
	q.On("CreateScoreJournalEntry", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			p := args.Get(1).(sqlc.CreateScoreJournalEntryParams)
			entries = append(entries, entry{p.UserID, derefString(p.EventID), derefString(p.SourceID), derefString(p.Reason), p.Points})
			ids[p.ID] = p.Points
		}).Return(&sqlc.ScoreJournal{}, nil).Maybe()
	var storedKind string
	q.On("SettleBetResult", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			storedKind = journalKind(ids[args.Get(1).(sqlc.SettleBetResultParams).Scorejournalid])
		}).Return(int64(1), nil).Maybe()
	res, err := Settle(context.Background(), q, in)
	require.NoError(t, err)
	return entries, storedKind, res
}

// journalKind tells the stake entry (negative) from the winnings entry
func journalKind(points int32) string {
	if points < 0 {
		return "stake"
	}
	return "winnings"
}

func TestSettleBatch_MatchesSettle(t *testing.T) {
	wrongMult := testInput(100, boolPtr(false))
	wrongMult.Bet.MultiplierWrong = fixedPtr(50)
	correctMult := testInput(77, boolPtr(true))
	correctMult.Bet.MultiplierCorrect = fixedPtr(150)
	noEvent := testInput(30, boolPtr(true))
	noEvent.EventID = nil

	cases := []SettleInput{
		testInput(100, boolPtr(true)),  // default x2
		testInput(100, boolPtr(false)), // default x0: winnings entry of 0
		testInput(40, nil),             // void: stake returned
		testInput(0, boolPtr(true)),    // no bet: skipped
		wrongMult,
		correctMult,
		noEvent,
	}
	ins := make([]SettleInput, len(cases))
	for i, c := range cases {
		c.Bet.ResponseID = fmt.Sprintf("QR%026d", i)
		ins[i] = c
	}

	q := mocks.NewMockBatchQuerier(t)
	claimAll(q)
	var written sqlc.CreateScoreJournalEntriesParams
	q.On("CreateScoreJournalEntries", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { written = args.Get(1).(sqlc.CreateScoreJournalEntriesParams) }).
		Return(nil).Once()
	var stored sqlc.SettleBetResultsParams
	q.On("SettleBetResults", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { stored = args.Get(1).(sqlc.SettleBetResultsParams) }).
		Return(nil).Once()

	results, err := SettleBatch(context.Background(), q, ins)
	require.NoError(t, err)
	require.Len(t, results, len(ins))

	// Journal entries and stored results of the batch, by response
	batchEntries := map[string][]entry{}
	pointsByID := map[string]int32{}
	for i := range written.Ids {
		e := entry{written.UserIds[i], written.EventIds[i], written.SourceIds[i], written.Reasons[i], written.Points[i]}
		batchEntries[e.sourceID] = append(batchEntries[e.sourceID], e)
		pointsByID[written.Ids[i]] = written.Points[i]
		assert.Equal(t, testProjectID, written.ProjectIds[i])
		assert.Equal(t, SourceTypeBet, written.SourceTypes[i])
		assert.Len(t, written.Ids[i], 28)
	}
	storedByID := map[string]int{}
	for i, id := range stored.Ids {
		storedByID[id] = i
	}

	for i, in := range ins {
		wantEntries, wantKind, want := settleOne(t, in)
		got := results[i]
		require.NoError(t, got.Err, "case %d", i)
		assert.Equal(t, want.Settled, got.Settled, "case %d", i)
		assert.Equal(t, want.SkipReason, got.SkipReason, "case %d", i)
		assert.Equal(t, want.NetPoints, got.NetPoints, "case %d", i)
		assert.Equal(t, want.Winnings, got.Winnings, "case %d", i)
		assert.Equal(t, want.Void, got.Void, "case %d", i)
		assert.Equal(t, wantEntries, batchEntries[in.Bet.ResponseID], "case %d entries", i)

		j, ok := storedByID[in.Bet.ResponseID]
		if !want.Settled {
			assert.False(t, ok, "case %d: skipped bet must not be stored", i)
			continue
		}
		require.True(t, ok, "case %d stored", i)
		assert.Equal(t, int32(want.NetPoints), stored.PointsEarned[j], "case %d", i)
		assert.Equal(t, got.JournalID, stored.ScoreJournalIds[j], "case %d", i)
		assert.Equal(t, wantKind, journalKind(pointsByID[got.JournalID]), "case %d: stored journal entry kind", i)
	}
}

func TestSettleBatch_SkipsBetsSettledElsewhere(t *testing.T) {
	ins := []SettleInput{
		batchInput(1, 100, boolPtr(true)),
		batchInput(2, 50, boolPtr(false)),
		batchInput(3, 10, nil),
	}

	q := mocks.NewMockBatchQuerier(t)
	// A concurrent settlement already paid out bet 2
	q.On("ClaimUnsettledBets", mock.Anything,
		[]string{ins[0].Bet.ResponseID, ins[1].Bet.ResponseID, ins[2].Bet.ResponseID}).
		Return([]string{ins[0].Bet.ResponseID, ins[2].Bet.ResponseID}, nil).Once()
	q.On("CreateScoreJournalEntries", mock.Anything, mock.MatchedBy(func(p sqlc.CreateScoreJournalEntriesParams) bool {
		for _, src := range p.SourceIds {
			if src == ins[1].Bet.ResponseID {
				return false
			}
		}
		return len(p.Ids) == 4
	})).Return(nil).Once()
	q.On("SettleBetResults", mock.Anything, mock.MatchedBy(func(p sqlc.SettleBetResultsParams) bool {
		return assert.ObjectsAreEqual([]string{ins[0].Bet.ResponseID, ins[2].Bet.ResponseID}, p.Ids)
	})).Return(nil).Once()

	results, err := SettleBatch(context.Background(), q, ins)
	require.NoError(t, err)
	assert.True(t, results[0].Settled)
	assert.ErrorIs(t, results[1].Err, ErrAlreadySettled)
	assert.False(t, results[1].Settled)
	assert.True(t, results[2].Settled)
	assert.True(t, results[2].Void)
}

func TestSettleBatch_NothingClaimedWritesNothing(t *testing.T) {
	ins := []SettleInput{batchInput(1, 100, boolPtr(true))}
	q := mocks.NewMockBatchQuerier(t)
	q.On("ClaimUnsettledBets", mock.Anything, mock.Anything).Return([]string{}, nil).Once()

	results, err := SettleBatch(context.Background(), q, ins)
	require.NoError(t, err)
	assert.ErrorIs(t, results[0].Err, ErrAlreadySettled)
}

func TestSettleBatch_NoBetsNoQueries(t *testing.T) {
	q := mocks.NewMockBatchQuerier(t) // no calls expected

	results, err := SettleBatch(context.Background(), q, []SettleInput{
		batchInput(1, 0, boolPtr(true)),
		batchInput(2, 0, nil),
	})
	require.NoError(t, err)
	for _, r := range results {
		assert.False(t, r.Settled)
		assert.Equal(t, SkipNoBet, r.SkipReason)
	}

	results, err = SettleBatch(context.Background(), q, nil)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSettleBatch_PayoutOutOfRangeOnlyFailsThatBet(t *testing.T) {
	ins := []SettleInput{
		batchInput(1, math.MaxInt32, boolPtr(true)), // 2x does not fit in int32
		batchInput(2, 100, boolPtr(true)),
	}
	q := mocks.NewMockBatchQuerier(t)
	q.On("ClaimUnsettledBets", mock.Anything, []string{ins[1].Bet.ResponseID}).
		Return([]string{ins[1].Bet.ResponseID}, nil).Once()
	q.On("CreateScoreJournalEntries", mock.Anything, mock.Anything).Return(nil).Once()
	q.On("SettleBetResults", mock.Anything, mock.Anything).Return(nil).Once()

	results, err := SettleBatch(context.Background(), q, ins)
	require.NoError(t, err)
	assert.ErrorIs(t, results[0].Err, ErrPayoutOutOfRange)
	assert.False(t, results[0].Settled)
	assert.True(t, results[1].Settled)
	assert.Equal(t, 100, results[1].NetPoints)
}

func TestSettleBatch_DatabaseErrors(t *testing.T) {
	dbErr := errors.New("db down")
	ins := []SettleInput{batchInput(1, 100, boolPtr(true))}

	t.Run("claim fails", func(t *testing.T) {
		q := mocks.NewMockBatchQuerier(t)
		q.On("ClaimUnsettledBets", mock.Anything, mock.Anything).Return(nil, dbErr).Once()
		_, err := SettleBatch(context.Background(), q, ins)
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("journal entries fail", func(t *testing.T) {
		q := mocks.NewMockBatchQuerier(t)
		claimAll(q)
		q.On("CreateScoreJournalEntries", mock.Anything, mock.Anything).Return(dbErr).Once()
		_, err := SettleBatch(context.Background(), q, ins)
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("storing results fails", func(t *testing.T) {
		q := mocks.NewMockBatchQuerier(t)
		claimAll(q)
		q.On("CreateScoreJournalEntries", mock.Anything, mock.Anything).Return(nil).Once()
		q.On("SettleBetResults", mock.Anything, mock.Anything).Return(dbErr).Once()
		_, err := SettleBatch(context.Background(), q, ins)
		require.ErrorIs(t, err, dbErr)
	})
}
