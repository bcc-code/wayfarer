package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLeaderboardApplyWorker_DrainUntilBelowBatch(t *testing.T) {
	ctx := context.Background()
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	w := NewLeaderboardApplyWorker(q)

	// Two full batches, then a partial one — drain must stop after the partial.
	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Return(int32(500), nil).Twice()
	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Return(int32(120), nil).Once()

	assert.Equal(t, 1120, w.drain(ctx))
}

func TestLeaderboardApplyWorker_DrainEmptyQueue(t *testing.T) {
	ctx := context.Background()
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	w := NewLeaderboardApplyWorker(q)

	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Return(int32(0), nil).Once()

	assert.Equal(t, 0, w.drain(ctx))
}

func TestLeaderboardApplyWorker_DrainStopsOnError(t *testing.T) {
	// Cancelled context: the error path must not sleep for the backoff.
	ctx, cancel := context.WithCancel(context.Background())
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	w := NewLeaderboardApplyWorker(q)

	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Return(int32(500), nil).Once()
	q.On("DrainLeaderboardApplyQueue", ctx, int32(500)).Run(func(_ mock.Arguments) {
		cancel()
	}).Return(int32(0), errors.New("db down")).Once()

	assert.Equal(t, 500, w.drain(ctx))
}

func TestLeaderboardApplyWorker_StartStopFinalDrain(t *testing.T) {
	q := mocks.NewMockLeaderboardApplyQuerier(t)
	w := NewLeaderboardApplyWorker(q)
	w.interval = 5 * time.Millisecond

	// Ticks during the run drain nothing; the final drain on Stop applies
	// what accumulated between the cancel and the shutdown.
	q.On("DrainLeaderboardApplyQueue", mock.Anything, int32(500)).Return(int32(0), nil).Maybe()

	w.Start(context.Background())
	time.Sleep(20 * time.Millisecond)
	w.Stop()
}
