//go:build bugrepro

// Reproduces R7 from notes/multi-session-loadtest-plan.md: the dataloaders are
// built once and shared by every request (cmd/server/main.go), and
// dataloader v7 runs a batch under the context of the first caller in the
// batch window. When that caller disconnects, every other request in the batch
// fails with its error.
//
// Opt-in until the fix lands: go test -race -tags bugrepro ./internal/loaders/
// Then drop the build tag so this runs in CI as a regression test.
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
