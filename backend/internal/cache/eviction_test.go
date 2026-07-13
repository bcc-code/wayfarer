package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKeyRegistry_PrunesEvictedKeys verifies that keys ristretto evicts or rejects
// on its own (cost-based eviction here) are removed from the KeyRegistry. Without the
// OnEvict/OnReject hook the registry would retain every key ever inserted, leaking
// memory and slowing DeletePrefix over time.
func TestKeyRegistry_PrunesEvictedKeys(t *testing.T) {
	c, err := NewCacheWithRegistry(Config{
		NumCounters: 1000,
		MaxCost:     50, // cost-1 items => ~50 entries live at most
		BufferItems: 64,
		DefaultTTL:  time.Hour,
	})
	require.NoError(t, err)

	const n = 5000
	for i := 0; i < n; i++ {
		c.Set(UserKey(fmt.Sprintf("US%026d", i)), i)
	}
	c.Wait() // flush the async set buffer, draining eviction/rejection callbacks

	// All distinct user keys register under the shared "user:" prefix; that slice is
	// where an unpruned registry would accumulate every inserted key. Poll until the
	// tracked count stabilizes so the async eviction callbacks (slower under -race)
	// have fully drained.
	tracked := len(c.registry.GetKeys(PrefixUser))
	for attempt := 0; attempt < 20; attempt++ {
		time.Sleep(25 * time.Millisecond)
		next := len(c.registry.GetKeys(PrefixUser))
		if next == tracked {
			break
		}
		tracked = next
	}
	t.Logf("registry tracked %d keys after inserting %d (live cache budget ~50)", tracked, n)

	// The strong claim: the registry no longer retains every inserted key. Without the
	// eviction hook this would be ~n. We use a generous bound (half of n) because
	// ristretto evicts from a sample, so the tracked set converges toward the live size
	// but not exactly to it within a single flush.
	assert.Less(t, tracked, n/2, "registry must prune keys ristretto evicted")
}

// TestKeyRegistry_ExplicitDeleteUnregisters is a sanity check that the value wrapping
// does not break the existing explicit Delete path.
func TestKeyRegistry_ExplicitDeleteUnregisters(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)

	key := UserKey("US00000000000000000000000001")
	c.Set(key, "v")
	c.Wait()

	got, ok := c.Get(key)
	require.True(t, ok)
	assert.Equal(t, "v", got, "value must be unwrapped on Get")

	c.Delete(key)
	assert.Empty(t, c.registry.GetKeys(PrefixUser), "explicit delete should unregister the key")
}
