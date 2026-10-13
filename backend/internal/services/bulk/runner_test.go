package bulk_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/bulk"
	"github.com/bcc-media/wayfarer/internal/services/bulk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testJobID = "BJ01ARZ3NDEKTSV4RRFFQ69G5FAV"

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func expectFailed(q *mocks.MockJobQuerier, msgContains string) {
	q.EXPECT().MarkBulkJobFailed(mock.Anything, mock.MatchedBy(func(p sqlc.MarkBulkJobFailedParams) bool {
		return p.ID == testJobID && strings.Contains(p.Errormessage, msgContains)
	})).Return(&sqlc.BulkJob{}, nil).Once()
}

func TestRunner_DispatchesEachOperation(t *testing.T) {
	ctx := context.Background()
	userIDs := []string{"US1", "US2"}

	tests := []struct {
		name   string
		params bulk.BulkOperationParams
		expect func(ops *mocks.MockOperationHandler)
	}{
		{"enroll", bulk.BulkEnrollChallengeParams{ChallengeID: "CL1", UserIDs: userIDs}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessBulkEnrollChallenge(mock.Anything, testJobID, bulk.BulkEnrollChallengeParams{ChallengeID: "CL1", UserIDs: userIDs}).Return(nil).Once()
		}},
		{"unenroll", bulk.BulkUnenrollChallengeParams{ChallengeID: "CL1", UserIDs: userIDs}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessBulkUnenrollChallenge(mock.Anything, testJobID, bulk.BulkUnenrollChallengeParams{ChallengeID: "CL1", UserIDs: userIDs}).Return(nil).Once()
		}},
		{"complete", bulk.BulkCompleteChallengeParams{ChallengeID: "CL1", UserIDs: userIDs, CompletedAt: "2026-01-01T00:00:00Z"}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessBulkCompleteChallenge(mock.Anything, testJobID, bulk.BulkCompleteChallengeParams{ChallengeID: "CL1", UserIDs: userIDs, CompletedAt: "2026-01-01T00:00:00Z"}).Return(nil).Once()
		}},
		{"publish", bulk.BulkPublishChallengeParams{ChallengeIDs: []string{"CL1"}, PublishedAt: "2026-01-01T00:00:00Z"}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessBulkPublishChallenge(mock.Anything, testJobID, bulk.BulkPublishChallengeParams{ChallengeIDs: []string{"CL1"}, PublishedAt: "2026-01-01T00:00:00Z"}).Return(nil).Once()
		}},
		{"award", bulk.BulkAwardAchievementParams{AchievementID: "AC1", UserIDs: userIDs}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessBulkAwardAchievement(mock.Anything, testJobID, bulk.BulkAwardAchievementParams{AchievementID: "AC1", UserIDs: userIDs}).Return(nil).Once()
		}},
		{"quiz access", bulk.BulkGrantQuizSessionAccessParams{SessionID: "QS1", UserIDs: userIDs, ProjectID: "PR1"}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessBulkGrantQuizSessionAccess(mock.Anything, testJobID, bulk.BulkGrantQuizSessionAccessParams{SessionID: "QS1", UserIDs: userIDs, ProjectID: "PR1"}).Return(nil).Once()
		}},
		{"score adjustment", bulk.BulkScoreAdjustmentParams{ProjectID: "PR1", Adjustments: []bulk.BulkScoreAdjustmentItem{{UserID: "US1", Points: 5}}}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessBulkScoreAdjustment(mock.Anything, testJobID, bulk.BulkScoreAdjustmentParams{ProjectID: "PR1", Adjustments: []bulk.BulkScoreAdjustmentItem{{UserID: "US1", Points: 5}}}).Return(nil).Once()
		}},
		{"fix content", bulk.FixMissingContentProgressParams{UserIDs: userIDs}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessFixMissingContentProgress(mock.Anything, testJobID, bulk.FixMissingContentProgressParams{UserIDs: userIDs}).Return(nil).Once()
		}},
		{"fix streak", bulk.FixMissingStreakProgressParams{UserIDs: userIDs}, func(ops *mocks.MockOperationHandler) {
			ops.EXPECT().ProcessFixMissingStreakProgress(mock.Anything, testJobID, bulk.FixMissingStreakProgressParams{UserIDs: userIDs}).Return(nil).Once()
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := mocks.NewMockJobQuerier(t)
			ops := mocks.NewMockOperationHandler(t)
			q.EXPECT().MarkBulkJobProcessing(mock.Anything, testJobID).Return(&sqlc.BulkJob{}, nil).Once()
			tt.expect(ops)

			r := bulk.NewRunner(q, ops, true, newTestLogger())
			r.Start(ctx, testJobID, tt.params.OperationType(), mustJSON(t, tt.params))
		})
	}
}

func TestRunner_MarksFailed(t *testing.T) {
	ctx := context.Background()

	t.Run("unknown operation type", func(t *testing.T) {
		q := mocks.NewMockJobQuerier(t)
		ops := mocks.NewMockOperationHandler(t)
		q.EXPECT().MarkBulkJobProcessing(mock.Anything, testJobID).Return(&sqlc.BulkJob{}, nil).Once()
		expectFailed(q, "unknown operation type")

		bulk.NewRunner(q, ops, true, newTestLogger()).Start(ctx, testJobID, "NOPE", json.RawMessage(`{}`))
	})

	t.Run("invalid params", func(t *testing.T) {
		q := mocks.NewMockJobQuerier(t)
		ops := mocks.NewMockOperationHandler(t)
		q.EXPECT().MarkBulkJobProcessing(mock.Anything, testJobID).Return(&sqlc.BulkJob{}, nil).Once()
		expectFailed(q, "failed to unmarshal params")

		bulk.NewRunner(q, ops, true, newTestLogger()).Start(ctx, testJobID, bulk.OperationBulkEnrollChallenge, json.RawMessage(`not json`))
	})

	t.Run("handler error", func(t *testing.T) {
		q := mocks.NewMockJobQuerier(t)
		ops := mocks.NewMockOperationHandler(t)
		q.EXPECT().MarkBulkJobProcessing(mock.Anything, testJobID).Return(&sqlc.BulkJob{}, nil).Once()
		ops.EXPECT().ProcessFixMissingStreakProgress(mock.Anything, testJobID, mock.Anything).Return(errors.New("boom")).Once()
		expectFailed(q, "boom")

		bulk.NewRunner(q, ops, true, newTestLogger()).Start(ctx, testJobID, bulk.OperationFixMissingStreakProgress, json.RawMessage(`{}`))
	})

	t.Run("handler panic", func(t *testing.T) {
		q := mocks.NewMockJobQuerier(t)
		ops := mocks.NewMockOperationHandler(t)
		q.EXPECT().MarkBulkJobProcessing(mock.Anything, testJobID).Return(&sqlc.BulkJob{}, nil).Once()
		ops.EXPECT().ProcessFixMissingStreakProgress(mock.Anything, testJobID, mock.Anything).Run(
			func(context.Context, string, bulk.FixMissingStreakProgressParams) { panic("kaboom") },
		).Return(nil).Once()
		expectFailed(q, "panic: kaboom")

		bulk.NewRunner(q, ops, true, newTestLogger()).Start(ctx, testJobID, bulk.OperationFixMissingStreakProgress, json.RawMessage(`{}`))
	})

	t.Run("mark processing fails skips run", func(t *testing.T) {
		q := mocks.NewMockJobQuerier(t)
		ops := mocks.NewMockOperationHandler(t)
		q.EXPECT().MarkBulkJobProcessing(mock.Anything, testJobID).Return(nil, errors.New("db down")).Once()

		bulk.NewRunner(q, ops, true, newTestLogger()).Start(ctx, testJobID, bulk.OperationFixMissingStreakProgress, json.RawMessage(`{}`))
	})
}

func TestRunner_Async(t *testing.T) {
	q := mocks.NewMockJobQuerier(t)
	ops := mocks.NewMockOperationHandler(t)

	release := make(chan struct{})
	var jobCtxErr error
	q.EXPECT().MarkBulkJobProcessing(mock.Anything, testJobID).Return(&sqlc.BulkJob{}, nil).Once()
	ops.EXPECT().ProcessFixMissingStreakProgress(mock.Anything, testJobID, mock.Anything).Run(
		func(ctx context.Context, _ string, _ bulk.FixMissingStreakProgressParams) {
			<-release
			jobCtxErr = ctx.Err()
		},
	).Return(nil).Once()

	r := bulk.NewRunner(q, ops, false, newTestLogger())
	assert.False(t, r.IsSync())

	reqCtx, cancelReq := context.WithCancel(context.Background())
	r.Start(reqCtx, testJobID, bulk.OperationFixMissingStreakProgress, json.RawMessage(`{}`))

	// Start returned while the job is still blocked; the request ending must not cancel it.
	cancelReq()

	shortCtx, cancelShort := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelShort()
	assert.ErrorIs(t, r.Wait(shortCtx), context.DeadlineExceeded)

	close(release)
	waitCtx, cancelWait := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelWait()
	require.NoError(t, r.Wait(waitCtx))
	assert.NoError(t, jobCtxErr)
}

func TestRunner_SyncRunsBeforeReturning(t *testing.T) {
	q := mocks.NewMockJobQuerier(t)
	ops := mocks.NewMockOperationHandler(t)

	ran := false
	q.EXPECT().MarkBulkJobProcessing(mock.Anything, testJobID).Return(&sqlc.BulkJob{}, nil).Once()
	ops.EXPECT().ProcessFixMissingStreakProgress(mock.Anything, testJobID, mock.Anything).Run(
		func(context.Context, string, bulk.FixMissingStreakProgressParams) { ran = true },
	).Return(nil).Once()

	r := bulk.NewRunner(q, ops, true, newTestLogger())
	assert.True(t, r.IsSync())
	r.Start(context.Background(), testJobID, bulk.OperationFixMissingStreakProgress, json.RawMessage(`{}`))
	assert.True(t, ran)
}
