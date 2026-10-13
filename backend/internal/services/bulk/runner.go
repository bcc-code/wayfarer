package bulk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// JobQuerier is the subset of sqlc queries the runner needs to track job status.
type JobQuerier interface {
	MarkBulkJobProcessing(ctx context.Context, id string) (*sqlc.BulkJob, error)
	MarkBulkJobFailed(ctx context.Context, arg sqlc.MarkBulkJobFailedParams) (*sqlc.BulkJob, error)
}

// OperationHandler executes a single bulk operation. Implementations are
// responsible for marking the job completed on success.
type OperationHandler interface {
	ProcessBulkEnrollChallenge(ctx context.Context, jobID string, params BulkEnrollChallengeParams) error
	ProcessBulkUnenrollChallenge(ctx context.Context, jobID string, params BulkUnenrollChallengeParams) error
	ProcessBulkCompleteChallenge(ctx context.Context, jobID string, params BulkCompleteChallengeParams) error
	ProcessBulkPublishChallenge(ctx context.Context, jobID string, params BulkPublishChallengeParams) error
	ProcessBulkAwardAchievement(ctx context.Context, jobID string, params BulkAwardAchievementParams) error
	ProcessBulkGrantQuizSessionAccess(ctx context.Context, jobID string, params BulkGrantQuizSessionAccessParams) error
	ProcessBulkScoreAdjustment(ctx context.Context, jobID string, params BulkScoreAdjustmentParams) error
	ProcessFixMissingContentProgress(ctx context.Context, jobID string, params FixMissingContentProgressParams) error
	ProcessFixMissingStreakProgress(ctx context.Context, jobID string, params FixMissingStreakProgressParams) error
}

// Runner executes bulk jobs on the instance that created them.
//
// In sync mode the job runs inside the caller's request. This is required on
// Cloud Run, which throttles CPU once the response has been sent, so background
// goroutines can't be relied on to finish. Otherwise the job runs in a
// goroutine detached from the request and Start returns immediately.
type Runner struct {
	queries JobQuerier
	ops     OperationHandler
	sync    bool
	logger  *slog.Logger
	wg      sync.WaitGroup
}

// NewRunner creates a bulk job runner
func NewRunner(queries JobQuerier, ops OperationHandler, sync bool, logger *slog.Logger) *Runner {
	return &Runner{
		queries: queries,
		ops:     ops,
		sync:    sync,
		logger:  logger,
	}
}

// IsSync reports whether jobs run inside the caller's request
func (r *Runner) IsSync() bool {
	return r.sync
}

// Start runs the job, either inline (sync mode) or in the background.
// Failures are recorded on the job row rather than returned.
func (r *Runner) Start(ctx context.Context, jobID string, opType OperationType, params json.RawMessage) {
	if r.sync {
		r.runAndLog(ctx, jobID, opType, params)
		return
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.runAndLog(context.WithoutCancel(ctx), jobID, opType, params)
	}()
}

// Wait blocks until all background jobs have finished or ctx is done.
func (r *Runner) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Runner) runAndLog(ctx context.Context, jobID string, opType OperationType, params json.RawMessage) {
	r.logger.Info("Running bulk job", "job_id", jobID, "operation_type", opType, "sync", r.sync)
	if err := r.run(ctx, jobID, opType, params); err != nil {
		r.logger.Error("Bulk job failed", "job_id", jobID, "operation_type", opType, "error", err)
		return
	}
	r.logger.Info("Bulk job finished", "job_id", jobID, "operation_type", opType)
}

// run marks the job as processing, dispatches it and marks it failed on error or panic.
func (r *Runner) run(ctx context.Context, jobID string, opType OperationType, params json.RawMessage) (err error) {
	if _, err := r.queries.MarkBulkJobProcessing(ctx, jobID); err != nil {
		return fmt.Errorf("failed to mark job as processing: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic: %v", p)
		}
		if err != nil {
			if _, markErr := r.queries.MarkBulkJobFailed(ctx, sqlc.MarkBulkJobFailedParams{
				ID:           jobID,
				Errormessage: err.Error(),
			}); markErr != nil {
				r.logger.Error("Failed to mark job as failed",
					"job_id", jobID,
					"mark_error", markErr,
					"process_error", err,
				)
			}
		}
	}()

	return r.dispatch(ctx, jobID, opType, params)
}

func (r *Runner) dispatch(ctx context.Context, jobID string, opType OperationType, raw json.RawMessage) error {
	switch opType {
	case OperationBulkEnrollChallenge:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessBulkEnrollChallenge)
	case OperationBulkUnenrollChallenge:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessBulkUnenrollChallenge)
	case OperationBulkCompleteChallenge:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessBulkCompleteChallenge)
	case OperationBulkPublishChallenge:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessBulkPublishChallenge)
	case OperationBulkAwardAchievement:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessBulkAwardAchievement)
	case OperationBulkGrantQuizSessionAccess:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessBulkGrantQuizSessionAccess)
	case OperationBulkScoreAdjustment:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessBulkScoreAdjustment)
	case OperationFixMissingContentProgress:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessFixMissingContentProgress)
	case OperationFixMissingStreakProgress:
		return decodeAndRun(ctx, jobID, raw, r.ops.ProcessFixMissingStreakProgress)
	default:
		return fmt.Errorf("unknown operation type: %s", opType)
	}
}

func decodeAndRun[P any](ctx context.Context, jobID string, raw json.RawMessage, fn func(context.Context, string, P) error) error {
	var params P
	if err := json.Unmarshal(raw, &params); err != nil {
		return fmt.Errorf("failed to unmarshal params: %w", err)
	}
	return fn(ctx, jobID, params)
}
