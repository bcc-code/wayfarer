// dataloaders are built once and shared by every request
// (cmd/server/main.go), and dataloader v7 runs a batch under the context of
// the first caller in the batch window. A caller disconnecting must not fail
// the other requests in its batch.
package loaders

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedLoader_CancelledCallerDoesNotFailOthersInBatch(t *testing.T) {
	var (
		mu      sync.Mutex
		batches [][]string
	)
	// Behaves like the real batch functions: the query runs with the batch
	// ctx, and pgx returns ctx.Err() once that ctx is cancelled.
	batchFn := func(ctx context.Context, keys []string) []*dataloader.Result[string] {
		mu.Lock()
		batches = append(batches, append([]string(nil), keys...))
		mu.Unlock()

		results := make([]*dataloader.Result[string], len(keys))
		select {
		case <-ctx.Done():
			for i := range keys {
				results[i] = &dataloader.Result[string]{Error: ctx.Err()}
			}
		case <-time.After(5 * time.Millisecond):
			for i, k := range keys {
				results[i] = &dataloader.Result[string]{Data: "value-" + k}
			}
		}
		return results
	}
	loader := newBatchedLoader(batchFn)

	// Request A arrives first and then disconnects; request B is still live.
	ctxA, cancelA := context.WithCancel(context.Background())
	thunkA := loader.Load(ctxA, "a")
	thunkB := loader.Load(context.Background(), "b")
	cancelA()

	_, _ = thunkA() // A may fail: it was cancelled.
	valB, errB := thunkB()

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, [][]string{{"a", "b"}}, batches,
		"precondition: both loads must share one batch")
	assert.NoError(t, errB, "a live request must not fail because another request in its batch was cancelled")
	assert.Equal(t, "value-b", valB)
}

func TestDetachedBatch_TimeoutBoundsTheBatch(t *testing.T) {
	// A batch whose query never returns ends at the loader's own timeout,
	// now that no caller's cancellation can end it
	batchFn := func(ctx context.Context, keys []string) []*dataloader.Result[string] {
		<-ctx.Done()
		results := make([]*dataloader.Result[string], len(keys))
		for i := range keys {
			results[i] = &dataloader.Result[string]{Error: ctx.Err()}
		}
		return results
	}
	loader := newBatchedLoaderWithTimeout(batchFn, 50*time.Millisecond)

	start := time.Now()
	_, err := loader.Load(context.Background(), "a")()

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second, "the batch must end at its timeout, not hang")
}

func TestDetachedBatch_KeepsContextValues(t *testing.T) {
	type ctxKey struct{}
	var (
		mu  sync.Mutex
		got any
	)
	batchFn := func(ctx context.Context, keys []string) []*dataloader.Result[string] {
		mu.Lock()
		got = ctx.Value(ctxKey{})
		mu.Unlock()
		results := make([]*dataloader.Result[string], len(keys))
		for i, k := range keys {
			results[i] = &dataloader.Result[string]{Data: k}
		}
		return results
	}
	loader := newBatchedLoader(batchFn)

	ctx := context.WithValue(context.Background(), ctxKey{}, "trace-and-request-values")
	_, err := loader.Load(ctx, "a")()

	require.NoError(t, err)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "trace-and-request-values", got, "values such as the trace span reach the batch")
}

func TestDetachedBatch_ConfiguredTimeoutIsUsed(t *testing.T) {
	previous := BatchTimeout
	BatchTimeout = 40 * time.Millisecond
	t.Cleanup(func() { BatchTimeout = previous })

	var deadline time.Duration
	batchFn := func(ctx context.Context, keys []string) []*dataloader.Result[string] {
		if d, ok := ctx.Deadline(); ok {
			deadline = time.Until(d)
		}
		results := make([]*dataloader.Result[string], len(keys))
		for i, k := range keys {
			results[i] = &dataloader.Result[string]{Data: k}
		}
		return results
	}
	loader := newBatchedLoader(batchFn)

	_, err := loader.Load(context.Background(), "a")()

	require.NoError(t, err)
	assert.Greater(t, deadline, time.Duration(0), "the batch context has a deadline")
	assert.LessOrEqual(t, deadline, 40*time.Millisecond, "the deadline comes from BatchTimeout")
}
