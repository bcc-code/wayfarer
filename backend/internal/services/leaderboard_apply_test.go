package services

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLeaderboardApplyWorker_DrainUntilBelowBatch(t *testing.T) {
	ctx := context.Background()
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	w := NewLeaderboardApplyWorker(q, nil)

	// Two full batches, then a partial one — drain must stop after the partial.
	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Return(int32(500), nil).Twice()
	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Return(int32(120), nil).Once()

	assert.Equal(t, 1120, w.drain(ctx))
}

func TestLeaderboardApplyWorker_DrainEmptyQueue(t *testing.T) {
	ctx := context.Background()
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	w := NewLeaderboardApplyWorker(q, nil)

	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Return(int32(0), nil).Once()

	assert.Equal(t, 0, w.drain(ctx))
}

func TestLeaderboardApplyWorker_DrainStopsOnError(t *testing.T) {
	// Cancelled context: the error path must not sleep for the backoff.
	ctx, cancel := context.WithCancel(context.Background())
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	w := NewLeaderboardApplyWorker(q, nil)

	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Return(int32(500), nil).Once()
	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Run(func(_ mock.Arguments) {
		cancel()
	}).Return(int32(0), errors.New("db down")).Once()

	assert.Equal(t, 500, w.drain(ctx))
}

func TestLeaderboardApplyWorker_StartStopFinalDrain(t *testing.T) {
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	w := NewLeaderboardApplyWorker(q, nil)
	w.interval = 5 * time.Millisecond

	// Ticks during the run drain nothing; the final drain on Stop applies
	// what accumulated between the cancel and the shutdown.
	q.On("DrainLeaderboardApplyQueue", mock.Anything, int32(500)).Return(int32(0), nil).Maybe()

	w.Start(context.Background())
	time.Sleep(20 * time.Millisecond)
	w.Stop()
}

// appliesFor makes the queue yield `ticks` small batches (one per drain
// call), then stay empty
func appliesFor(q *mocks.MockLeaderboardApplyQuerier, ticks int32) {
	var left atomic.Int32
	left.Store(ticks)
	q.On("DrainLeaderboardApplyQueue", mock.Anything, int32(500)).
		Return(func(context.Context, int32) (int32, error) {
			if left.Add(-1) >= 0 {
				return 10, nil
			}
			return 0, nil
		}).Maybe()
}

func TestLeaderboardApplyWorker_OnAppliedAfterDeltas(t *testing.T) {
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	var calls atomic.Int32
	w := NewLeaderboardApplyWorker(q, func() { calls.Add(1) })
	w.interval = 2 * time.Millisecond
	appliesFor(q, 1)

	w.Start(context.Background())
	defer w.Stop()
	assert.Eventually(t, func() bool { return calls.Load() == 1 }, time.Second, time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(1), calls.Load(), "an empty queue must not refresh again")
}

func TestLeaderboardApplyWorker_OnAppliedNotWhenNothingApplied(t *testing.T) {
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	var calls atomic.Int32
	w := NewLeaderboardApplyWorker(q, func() { calls.Add(1) })
	w.interval = 2 * time.Millisecond
	appliesFor(q, 0)

	w.Start(context.Background())
	time.Sleep(30 * time.Millisecond)
	w.Stop()
	assert.Zero(t, calls.Load())
}

func TestLeaderboardApplyWorker_OnAppliedThrottledThenFlushed(t *testing.T) {
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	var calls atomic.Int32
	var last atomic.Int64
	w := NewLeaderboardApplyWorker(q, func() {
		calls.Add(1)
		last.Store(time.Now().UnixNano())
	})
	w.interval = 2 * time.Millisecond
	w.refreshEvery = 40 * time.Millisecond
	// Deltas on ~50 consecutive ticks (~100ms): far fewer refreshes than ticks
	appliesFor(q, 50)

	w.Start(context.Background())
	defer w.Stop()
	time.Sleep(250 * time.Millisecond)

	n := calls.Load()
	assert.GreaterOrEqual(t, n, int32(2), "refreshes while deltas keep coming")
	assert.LessOrEqual(t, n, int32(6), "throttled to about one per refreshEvery")
	// The deltas of the last ticks were refreshed too, after they stopped
	assert.Eventually(t, func() bool {
		return time.Since(time.Unix(0, last.Load())) > 50*time.Millisecond && calls.Load() == n
	}, time.Second, 5*time.Millisecond)
}

func TestLeaderboardApplyWorker_StopFinalDrainRefreshes(t *testing.T) {
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	var calls atomic.Int32
	w := NewLeaderboardApplyWorker(q, func() { calls.Add(1) })
	w.interval = time.Hour // only the final drain on Stop applies anything

	q.On("DrainLeaderboardApplyQueue", mock.Anything, int32(500)).Return(int32(7), nil).Once()

	w.Start(context.Background())
	w.Stop()
	assert.Equal(t, int32(1), calls.Load())
}
