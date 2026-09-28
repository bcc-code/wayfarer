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

// TestKeyRegistry_DuplicateSetKeepsRegistration pins a subtle ristretto interaction:
// when the same new key is Set twice before the set buffer is processed, ristretto
// stores the first item and fires OnReject for the second — for a key whose value is
// alive in the cache. Unregistering on that callback would strand the live entry
// outside the registry, so DeletePrefix (i.e. every prefix-based invalidation) would
// silently stop covering it until TTL expiry.
func TestKeyRegistry_DuplicateSetKeepsRegistration(t *testing.T) {
	c, err := NewCacheWithRegistry(Config{
		NumCounters: 1000,
		MaxCost:     10_000,
		BufferItems: 64,
		DefaultTTL:  time.Hour,
	})
	require.NoError(t, err)
	defer c.Close()

	key := UserQuizSessionAccessKey("US01ARZ3NDEKTSV4RRFFQ69G5FAV", "PR01ARZ3NDEKTSV4RRFFQ69G5FAV")

	// Two back-to-back sets of the same fresh key: both enter the set buffer as
	// itemNew, the second is rejected by the admission policy (key already added).
	c.SetWithTTL(key, map[string]bool{"QZA": true}, time.Hour)
	c.SetWithTTL(key, map[string]bool{"QZA": true}, time.Hour)
	c.Wait()

	// The value must still be cached, and — after the async pruner has had time to
	// process the rejection — the key must still be registered so prefix
	// invalidation can find it.
	_, cached := c.Get(key)
	require.True(t, cached, "value should be cached after duplicate set")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(c.registry.GetKeys(PrefixUserQuizSessionAccess)) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.NotEmpty(t, c.registry.GetKeys(PrefixUserQuizSessionAccess),
		"registry must keep a live key registered after a duplicate-set rejection")

	// And prefix invalidation must actually clear it.
	c.DeletePrefix(PrefixUserQuizSessionAccess)
	_, cached = c.Get(key)
	assert.False(t, cached, "DeletePrefix must delete the live entry")
}

// TestKeyRegistry_UnregisterIfGen verifies the generation guard: an unregister
// based on a stale generation (a Register happened in between) must be a no-op,
// so the async pruner cannot unregister a concurrently re-inserted key.
func TestKeyRegistry_UnregisterIfGen(t *testing.T) {
	kr := NewKeyRegistry()
	key := UserKey("US01ARZ3NDEKTSV4RRFFQ69G5FAV")

	kr.Register(key)
	gen, ok := kr.Gen(key)
	require.True(t, ok)

	// Concurrent re-insert bumps the generation; the stale unregister is ignored.
	kr.Register(key)
	kr.UnregisterIfGen(key, gen)
	_, stillRegistered := kr.Gen(key)
	assert.True(t, stillRegistered, "stale-generation unregister must not remove a re-registered key")

	// With the current generation it removes the key.
	gen, _ = kr.Gen(key)
	kr.UnregisterIfGen(key, gen)
	_, stillRegistered = kr.Gen(key)
	assert.False(t, stillRegistered)
	assert.Empty(t, kr.GetKeys(PrefixUser))
}

// TestCacheWithRegistry_SweepPrunesDeadKeys verifies the overflow fallback: a
// full registry sweep unregisters keys whose cache entries are gone but keeps
// live ones. (In production the sweep runs when the eviction queue overflowed
// and notifications were dropped.)
func TestCacheWithRegistry_SweepPrunesDeadKeys(t *testing.T) {
	c, err := NewCacheWithRegistry(Config{
		NumCounters: 1000,
		MaxCost:     10_000,
		BufferItems: 64,
		DefaultTTL:  time.Hour,
	})
	require.NoError(t, err)
	defer c.Close()

	liveKey := UserKey("US01ARZ3NDEKTSV4RRFFQ69G5FAV")
	deadKey := UserKey("US01ARZ3NDEKTSV4RRFFQ69G5FAX")
	require.True(t, c.SetWithTTL(liveKey, "live", time.Hour))
	require.True(t, c.SetWithTTL(deadKey, "dead", time.Hour))
	c.Wait()

	// Delete the entry behind the registry's back (embedded Cache, so no
	// Unregister happens) — this is the state a dropped eviction leaves.
	c.Cache.Delete(deadKey)
	c.Wait()

	c.pruneKeys(c.registry.AllKeys())

	keys := c.registry.GetKeys(PrefixUser)
	assert.Contains(t, keys, liveKey, "sweep must keep keys whose entries are still cached")
	assert.NotContains(t, keys, deadKey, "sweep must unregister keys whose entries are gone")
}
