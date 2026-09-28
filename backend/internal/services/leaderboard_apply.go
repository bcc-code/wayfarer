package services

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// LeaderboardApplyQuerier is the database access the apply worker needs.
type LeaderboardApplyQuerier interface {
	DrainLeaderboardApplyQueue(ctx context.Context, batchSize int32) (int32, error)
	CountLeaderboardApplyQueue(ctx context.Context) (int64, error)
}

const (
	leaderboardApplyInterval  = 200 * time.Millisecond
	leaderboardApplyBatchSize = 500
	// Backlog above this logs a warning once per report interval — it means
	// the applier is falling behind the award rate.
	leaderboardApplyBacklogWarn   = 5000
	leaderboardApplyReportEvery   = time.Minute
	leaderboardApplyErrorBackoff  = 2 * time.Second
	leaderboardApplyDrainMaxLoops = 100 // per tick; bounds one tick's work
)

// LeaderboardApplyWorker drains the leaderboard_apply_queue outbox filled by
// the score_journal INSERT trigger (migration 00101), applying score deltas
// to the leaderboard tables outside the request transaction. Deltas commute,
// so batching and concurrent instances (SKIP LOCKED) are safe.
type LeaderboardApplyWorker struct {
	queries LeaderboardApplyQuerier

	interval  time.Duration
	batchSize int32

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewLeaderboardApplyWorker constructs the worker with production defaults.
func NewLeaderboardApplyWorker(queries LeaderboardApplyQuerier) *LeaderboardApplyWorker {
	return &LeaderboardApplyWorker{
		queries:   queries,
		interval:  leaderboardApplyInterval,
		batchSize: leaderboardApplyBatchSize,
	}
}

// Start launches the background drain loop. Call Stop to shut it down.
func (w *LeaderboardApplyWorker) Start(ctx context.Context) {
	ctx, w.cancel = context.WithCancel(ctx)
	w.wg.Add(1)
	go w.run(ctx)
}

// Stop cancels the loop and waits for the in-flight batch to finish, then
// performs one final drain so a clean shutdown leaves no backlog behind.
func (w *LeaderboardApplyWorker) Stop() {
	if w.cancel == nil {
		return
	}
	w.cancel()
	w.wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if n := w.drain(ctx); n > 0 {
		slog.Info("leaderboard apply: drained on shutdown", "applied", n)
	}
}

func (w *LeaderboardApplyWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	report := time.NewTicker(leaderboardApplyReportEvery)
	defer report.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.drain(ctx)
		case <-report.C:
			backlog, err := w.queries.CountLeaderboardApplyQueue(ctx)
			if err == nil && backlog > leaderboardApplyBacklogWarn {
				slog.Warn("leaderboard apply: backlog high", "backlog", backlog)
			}
		}
	}
}

// drain applies queued deltas until the queue yields less than a full batch
// (or the loop bound is hit), returning the total applied.
func (w *LeaderboardApplyWorker) drain(ctx context.Context) int {
	total := 0
	for i := 0; i < leaderboardApplyDrainMaxLoops; i++ {
		applied, err := w.queries.DrainLeaderboardApplyQueue(ctx, w.batchSize)
		if err != nil {
			if ctx.Err() == nil {
				slog.Error("leaderboard apply: drain failed", "error", err)
				select {
				case <-ctx.Done():
				case <-time.After(leaderboardApplyErrorBackoff):
				}
			}
			return total
		}
		total += int(applied)
		if applied < w.batchSize {
			return total
		}
	}
	return total
}
