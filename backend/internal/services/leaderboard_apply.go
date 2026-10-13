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
	// Applied deltas refresh the cached leaderboards at most this often, so
	// a steady trickle of scores doesn't flush the caches on every tick
	leaderboardApplyRefreshEvery = time.Second
)

// LeaderboardApplyWorker drains the leaderboard_apply_queue outbox filled by
// the score_journal INSERT trigger (migration 00101), applying score deltas
// to the leaderboard tables outside the request transaction. Deltas commute,
// so batching and concurrent instances (SKIP LOCKED) are safe.
type LeaderboardApplyWorker struct {
	queries LeaderboardApplyQuerier
	// onApplied runs after deltas were applied (throttled to refreshEvery),
	// e.g. to drop cached leaderboards that are now stale. Optional.
	onApplied    func()
	refreshEvery time.Duration

	interval  time.Duration
	batchSize int32

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewLeaderboardApplyWorker constructs the worker with production defaults.
// onApplied (may be nil) runs at most once per second while deltas are being
// applied, and once more after the last ones, so leaderboard caches never
// outlive the table state they were read from by more than about a second.
func NewLeaderboardApplyWorker(queries LeaderboardApplyQuerier, onApplied func()) *LeaderboardApplyWorker {
	return &LeaderboardApplyWorker{
		queries:      queries,
		onApplied:    onApplied,
		refreshEvery: leaderboardApplyRefreshEvery,
		interval:     leaderboardApplyInterval,
		batchSize:    leaderboardApplyBatchSize,
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
		if w.onApplied != nil {
			w.onApplied()
		}
	}
}

func (w *LeaderboardApplyWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	report := time.NewTicker(leaderboardApplyReportEvery)
	defer report.Stop()

	var pending bool // deltas applied since the last onApplied
	var lastRefresh time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.drain(ctx) > 0 {
				pending = true
			}
			if pending && w.onApplied != nil && time.Since(lastRefresh) >= w.refreshEvery {
				w.onApplied()
				pending = false
				lastRefresh = time.Now()
			}
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
