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
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: cfg.BufferItems,
		Metrics:     true, // Enable metrics for monitoring
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
	return c.cache.Get(key)
}

// Set stores a value in the cache with the default TTL
func (c *Cache) Set(key string, value interface{}) bool {
	return c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) bool {
	// Cost of 1 per item (can be customized based on actual size if needed)
	return c.cache.SetWithTTL(key, value, 1, ttl)
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
