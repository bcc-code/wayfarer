package cache

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
)

// Cache wraps ristretto cache with our custom configuration and methods
type Cache struct {
	cache      *ristretto.Cache
	defaultTTL time.Duration
}

// Config holds cache configuration
type Config struct {
	// NumCounters is the number of keys to track frequency (10x max items recommended)
	NumCounters int64
	// MaxCost is the maximum memory size in bytes (items have cost 1 by default)
	MaxCost int64
	// BufferItems is the number of keys per Get buffer
	BufferItems int64
	// DefaultTTL is the default expiration time for cache entries
	DefaultTTL time.Duration

	// onEvictKey, if set, is called with the original string key whenever ristretto
	// evicts or rejects an entry. It lets a KeyRegistry prune keys that ristretto
	// removed on its own (TTL expiry / cost eviction), which it otherwise never
	// learns about. Set internally by NewCacheWithRegistry.
	onEvictKey func(key string)
}

// registryEntry wraps every cached value so the original string key can be
// recovered in ristretto's eviction callbacks (which only expose the hashed key).
type registryEntry struct {
	key   string
	value interface{}
}

// DefaultConfig returns sensible defaults for cache configuration
func DefaultConfig() Config {
	return Config{
		NumCounters: 100_000,     // Track 100k items
		MaxCost:     100_000_000, // 100MB max cache size
		BufferItems: 64,          // 64 keys per buffer
		DefaultTTL:  15 * time.Minute,
	}
}

// New creates a new cache instance with the given configuration
func New(cfg Config) (*Cache, error) {
	// When a key-eviction hook is configured, recover the original string key from
	// the wrapped value and forward it. OnEvict covers TTL/cost eviction; OnReject
	// covers admission rejection — both remove the value ristretto held for us.
	var onEvict, onReject func(*ristretto.Item)
	if cfg.onEvictKey != nil {
		hook := func(item *ristretto.Item) {
			if entry, ok := item.Value.(*registryEntry); ok {
				cfg.onEvictKey(entry.key)
			}
		}
		onEvict = hook
		onReject = hook
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: cfg.BufferItems,
		Metrics:     true, // Enable metrics for monitoring
		OnEvict:     onEvict,
		OnReject:    onReject,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	return &Cache{
		cache:      cache,
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	raw, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	if entry, ok := raw.(*registryEntry); ok {
		return entry.value, true
	}
	return raw, true
}

// Set stores a value in the cache with the default TTL
func (c *Cache) Set(key string, value interface{}) bool {
	return c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) bool {
	// Cost of 1 per item (can be customized based on actual size if needed).
	// The value is wrapped so eviction callbacks can recover the string key.
	return c.cache.SetWithTTL(key, &registryEntry{key: key, value: value}, 1, ttl)
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.cache.Del(key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.cache.Clear()
}

// Metrics returns cache statistics
func (c *Cache) Metrics() *ristretto.Metrics {
	return c.cache.Metrics
}

// Wait blocks until all pending writes to the cache have been processed
func (c *Cache) Wait() {
	c.cache.Wait()
}

// Close waits for all goroutines to finish and closes the cache
func (c *Cache) Close() {
	c.cache.Close()
}
