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
	testQuestionID = "QQ01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testUserID     = "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testProjectID  = "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testEventID    = "EV01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testTypeStub   = QuestionType("STUB")
)

func int32Ptr(v int32) *int32 { return &v }
func strPtr(v string) *string { return &v }

// fixedStrategy evaluates every response to outcome and pays multiplier.
func fixedStrategy(outcome Outcome, multiplier float64) Strategy {
	return Strategy{
		Evaluator: EvaluatorFunc(func(Response, []Answer) (Outcome, error) {
			return outcome, nil
		}),
		PayoutRule: PayoutRuleFunc(func(Outcome) float64 { return multiplier }),
	}
}

func failingStrategy(err error) Strategy {
	return Strategy{
		Evaluator: EvaluatorFunc(func(Response, []Answer) (Outcome, error) {
			return Outcome{}, err
		}),
		PayoutRule: PayoutRuleFunc(func(Outcome) float64 { return 2.0 }),
	}
}

func validInput(bet int32) SettleInput {
	eventID := testEventID
	return SettleInput{
		Response: Response{
			ID:             testResponseID,
			QuestionID:     testQuestionID,
			QuestionType:   testTypeStub,
			BettingEnabled: true,
			BetAmount:      int32Ptr(bet),
		},
		UserID:        testUserID,
		ProjectID:     testProjectID,
		EventID:       &eventID,
		Language:      i18n.DefaultLanguage,
		ChallengeName: "Game Night",
	}
}

// journalEntry matches a journal entry with the given points and the shared bet fields.
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

// captureJournalIDs records the IDs of created journal entries in creation order.
func captureJournalIDs(q *mocks.MockQuerier, ids *[]string, points ...int32) {
	for _, pts := range points {
		q.On("CreateScoreJournalEntry", mock.Anything, journalEntry(pts)).
			Run(func(args mock.Arguments) {
				*ids = append(*ids, args.Get(1).(sqlc.CreateScoreJournalEntryParams).ID)
			}).
			Return(&sqlc.ScoreJournal{}, nil).Once()
	}
}

func TestSettle_Skips(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SettleInput)
		want   SkipReason
	}{
		{
			name:   "nil bet",
			mutate: func(in *SettleInput) { in.Response.BetAmount = nil },
			want:   SkipNoBet,
		},
		{
			name:   "zero bet",
			mutate: func(in *SettleInput) { in.Response.BetAmount = int32Ptr(0) },
			want:   SkipNoBet,
		},
		{
			name:   "betting disabled",
			mutate: func(in *SettleInput) { in.Response.BettingEnabled = false },
			want:   SkipBettingDisabled,
		},
		{
			name:   "already settled",
			mutate: func(in *SettleInput) { in.Response.ScoreJournalID = strPtr("SJ01ARZ3NDEKTSV4RRFFQ69G5FAV") },
			want:   SkipAlreadySettled,
		},
		{
			name:   "unsupported question type",
			mutate: func(in *SettleInput) { in.Response.QuestionType = QuestionTypeFreeText },
			want:   SkipUnsupportedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := mocks.NewMockQuerier(t) // no calls expected
			svc := NewService(q, map[QuestionType]Strategy{
				testTypeStub: fixedStrategy(Outcome{Correct: 1, Total: 1}, 2.0),
			})

			in := validInput(100)
			tt.mutate(&in)

			res, err := svc.Settle(context.Background(), in)
			require.NoError(t, err)
			assert.False(t, res.Settled)
			assert.Equal(t, tt.want, res.SkipReason)
		})
	}
}

func TestSettle_EmptyJournalIDStillSettles(t *testing.T) {
	q := mocks.NewMockQuerier(t)
	var ids []string
	captureJournalIDs(q, &ids, -100, 200)
	q.On("UpdateBetResultWithJournal", mock.Anything, mock.Anything).
		Return(&sqlc.UpdateBetResultWithJournalRow{}, nil).Once()

	svc := NewService(q, map[QuestionType]Strategy{
		testTypeStub: fixedStrategy(Outcome{Correct: 1, Total: 1}, 2.0),
	})
	in := validInput(100)
	in.Response.ScoreJournalID = strPtr("")

	res, err := svc.Settle(context.Background(), in)
	require.NoError(t, err)
	assert.True(t, res.Settled)
}

func TestSettle_NotEvaluable(t *testing.T) {
	q := mocks.NewMockQuerier(t) // no calls expected
	svc := NewService(q, map[QuestionType]Strategy{
		testTypeStub: failingStrategy(errors.Join(ErrNotEvaluable, errors.New("bad json"))),
	})

	res, err := svc.Settle(context.Background(), validInput(100))
	require.NoError(t, err)
	assert.False(t, res.Settled)
	assert.Equal(t, SkipNotEvaluable, res.SkipReason)
}

func TestSettle_EvaluatorError(t *testing.T) {
	q := mocks.NewMockQuerier(t) // no calls expected
	svc := NewService(q, map[QuestionType]Strategy{
		testTypeStub: failingStrategy(errors.New("boom")),
	})

	_, err := svc.Settle(context.Background(), validInput(100))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestSettle_Payouts(t *testing.T) {
	tests := []struct {
		name          string
		bet           int32
		multiplier    float64
		wantWinnings  int
		wantNet       int
		wantWinningID bool // true: response stores winnings entry, false: stake entry
	}{
		{name: "double", bet: 100, multiplier: 2.0, wantWinnings: 200, wantNet: 100, wantWinningID: true},
		{name: "partial", bet: 100, multiplier: 1.25, wantWinnings: 125, wantNet: 25, wantWinningID: true},
		{name: "lost", bet: 100, multiplier: 0, wantWinnings: 0, wantNet: -100, wantWinningID: false},
		{name: "partial loss", bet: 100, multiplier: 0.5, wantWinnings: 50, wantNet: -50, wantWinningID: true},
		{name: "winnings truncate", bet: 3, multiplier: 1.25, wantWinnings: 3, wantNet: 0, wantWinningID: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := mocks.NewMockQuerier(t)
			var ids []string
			captureJournalIDs(q, &ids, -tt.bet, int32(tt.wantWinnings))

			var stored sqlc.UpdateBetResultWithJournalParams
			q.On("UpdateBetResultWithJournal", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					stored = args.Get(1).(sqlc.UpdateBetResultWithJournalParams)
				}).
				Return(&sqlc.UpdateBetResultWithJournalRow{}, nil).Once()

			outcome := Outcome{Correct: 2, Total: 4}
			svc := NewService(q, map[QuestionType]Strategy{
				testTypeStub: fixedStrategy(outcome, tt.multiplier),
			})

			res, err := svc.Settle(context.Background(), validInput(tt.bet))
			require.NoError(t, err)
			require.Len(t, ids, 2)
			stakeID, winningsID := ids[0], ids[1]

			assert.True(t, res.Settled)
			assert.Equal(t, SkipNone, res.SkipReason)
			assert.Equal(t, outcome, res.Outcome)
			assert.Equal(t, tt.multiplier, res.Multiplier)
			assert.Equal(t, int(tt.bet), res.Stake)
			assert.Equal(t, tt.wantWinnings, res.Winnings)
			assert.Equal(t, tt.wantNet, res.NetPoints)

			wantJournalID := stakeID
			if tt.wantWinningID {
				wantJournalID = winningsID
			}
			assert.Equal(t, wantJournalID, res.JournalID)
			assert.Equal(t, sqlc.UpdateBetResultWithJournalParams{
				ID:             testResponseID,
				Pointsearned:   int32(tt.wantNet),
				Scorejournalid: wantJournalID,
			}, stored)
		})
	}
}

func TestSettle_PassesResponseAndAnswersToEvaluator(t *testing.T) {
	q := mocks.NewMockQuerier(t)
	var ids []string
	captureJournalIDs(q, &ids, -10, 20)
	q.On("UpdateBetResultWithJournal", mock.Anything, mock.Anything).
		Return(&sqlc.UpdateBetResultWithJournalRow{}, nil).Once()

	answers := []Answer{{ID: "A1", IsCorrect: true}, {ID: "A2"}}
	var gotResp Response
	var gotAnswers []Answer
	var gotOutcome Outcome
	svc := NewService(q, map[QuestionType]Strategy{
		testTypeStub: {
			Evaluator: EvaluatorFunc(func(r Response, a []Answer) (Outcome, error) {
				gotResp, gotAnswers = r, a
				return Outcome{Correct: 1, Total: 1}, nil
			}),
			PayoutRule: PayoutRuleFunc(func(o Outcome) float64 {
				gotOutcome = o
				return 2.0
			}),
		},
	})

	in := validInput(10)
	in.Answers = answers
	_, err := svc.Settle(context.Background(), in)
	require.NoError(t, err)

	assert.Equal(t, in.Response, gotResp)
	assert.Equal(t, answers, gotAnswers)
	assert.Equal(t, Outcome{Correct: 1, Total: 1}, gotOutcome)
}

func TestSettle_DatabaseErrors(t *testing.T) {
	dbErr := errors.New("db down")

	t.Run("stake entry fails", func(t *testing.T) {
		q := mocks.NewMockQuerier(t)
		q.On("CreateScoreJournalEntry", mock.Anything, journalEntry(-100)).Return(nil, dbErr).Once()

		svc := NewService(q, map[QuestionType]Strategy{testTypeStub: fixedStrategy(Outcome{1, 1}, 2.0)})
		_, err := svc.Settle(context.Background(), validInput(100))
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("winnings entry fails", func(t *testing.T) {
		q := mocks.NewMockQuerier(t)
		q.On("CreateScoreJournalEntry", mock.Anything, journalEntry(-100)).Return(&sqlc.ScoreJournal{}, nil).Once()
		q.On("CreateScoreJournalEntry", mock.Anything, journalEntry(200)).Return(nil, dbErr).Once()

		svc := NewService(q, map[QuestionType]Strategy{testTypeStub: fixedStrategy(Outcome{1, 1}, 2.0)})
		_, err := svc.Settle(context.Background(), validInput(100))
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("storing result fails", func(t *testing.T) {
		q := mocks.NewMockQuerier(t)
		var ids []string
		captureJournalIDs(q, &ids, -100, 200)
		q.On("UpdateBetResultWithJournal", mock.Anything, mock.Anything).Return(nil, dbErr).Once()

		svc := NewService(q, map[QuestionType]Strategy{testTypeStub: fixedStrategy(Outcome{1, 1}, 2.0)})
		_, err := svc.Settle(context.Background(), validInput(100))
		require.ErrorIs(t, err, dbErr)
	})
}

func TestSupports(t *testing.T) {
	svc := NewService(nil, map[QuestionType]Strategy{
		QuestionTypeOrdering: fixedStrategy(Outcome{}, 0),
	})
	assert.True(t, svc.Supports(QuestionTypeOrdering))
	assert.False(t, svc.Supports(QuestionTypeNumber))
}

func TestOutcome_AllCorrect(t *testing.T) {
	assert.True(t, Outcome{Correct: 4, Total: 4}.AllCorrect())
	assert.True(t, Outcome{Correct: 1, Total: 1}.AllCorrect())
	assert.False(t, Outcome{Correct: 3, Total: 4}.AllCorrect())
	assert.False(t, Outcome{Correct: 0, Total: 0}.AllCorrect())
}
