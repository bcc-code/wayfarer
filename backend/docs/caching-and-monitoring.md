# Caching and Performance Monitoring

This document describes how to monitor cache performance and database queries during development and load testing.

## Cache Metrics Endpoint

The server exposes a `/metrics/cache` endpoint that provides real-time cache statistics.

### Accessing Cache Metrics

```bash
# View current cache metrics
curl http://localhost:8080/metrics/cache | jq

# Watch cache metrics in real-time during load test
watch -n 1 'curl -s http://localhost:8080/metrics/cache | jq'
```

### Metrics Explained

```json
{
  "hits": 15234,           // Number of cache hits
  "misses": 982,           // Number of cache misses
  "total": 16216,          // Total cache requests
  "hit_rate": 0.9394,      // Hit rate (0-1)
  "hit_rate_pct": 93.94,   // Hit rate as percentage
  "cost_added": 16216,     // Total cost of items added
  "cost_evicted": 0,       // Cost of items evicted
  "keys_added": 16216,     // Number of keys added
  "keys_updated": 0,       // Number of keys updated
  "keys_evicted": 0,       // Number of keys evicted
  "sets_dropped": 0,       // Sets dropped (buffer full)
  "sets_rejected": 0       // Sets rejected (policy)
}
```

### Target Metrics

For optimal performance during load tests:
- **hit_rate_pct**: Should be > 70% for realistic load (34 users)
- **hit_rate_pct**: Should be > 95% for single-user load
- **keys_evicted**: Should be 0 or very low (indicates cache size is adequate)
- **sets_dropped**: Should be 0 (indicates buffer size is adequate)

---

## Database Query Logging

Enable detailed database query logging to identify performance bottlenecks and verify caching is working.

### Enabling Query Logging

Set the environment variable or add to `.env` file:

```bash
# Enable query logging
DB_LOG_QUERIES=true

# Run server with query logging
DB_LOG_QUERIES=true go run ./cmd/server
```

### Query Log Output

When enabled, all database queries are logged with timing information:

```json
{
  "time": "2025-01-04T12:34:56Z",
  "level": "INFO",
  "msg": "Query",
  "component": "database",
  "sql": "SELECT id, name, email FROM users WHERE id = ANY($1)",
  "args": [["US01K8XV6EN42TPEATDT708X51KE"]],
  "time": "2.5ms"
}
```

### Using Query Logs During Load Tests

```bash
# Terminal 1: Start server with query logging
DB_LOG_QUERIES=true go run ./cmd/server 2>&1 | grep "Query"

# Terminal 2: Run load test
k6 run loadtest/loadtest_me_realistic.js

# Terminal 3: Monitor cache hit rate
watch -n 1 'curl -s http://localhost:8080/metrics/cache | jq ".hit_rate_pct"'
```

### Interpreting Results

**Good caching (after fixes):**
```
[First few requests]
Query: SELECT ... FROM projects ...
Query: SELECT ... FROM users WHERE id = ANY($1) ...
Query: SELECT ... FROM churches WHERE id = ANY($1) ...

[Then mostly silence with occasional queries for cache misses]
```

**Poor caching (before fixes):**
```
[Continuous stream of queries]
Query: SELECT ... FROM projects ...
Query: SELECT ... FROM projects ...
Query: SELECT ... FROM projects ...
Query: SELECT ... FROM users ...
Query: SELECT ... FROM users ...
[Hundreds of repeated queries]
```

---

## Performance Testing Workflow

### 1. Baseline Test (No Cache)

```bash
# Clear cache or restart server
# Run load test and capture metrics
k6 run loadtest/loadtest_me_realistic.js > baseline.txt

# Check database query count (with query logging enabled)
# Count should be very high (3000+ for projects query alone)
```

### 2. Test with Cache

```bash
# Restart server to ensure clean state
go run ./cmd/server

# Run load test
k6 run loadtest/loadtest_me_realistic.js > cached.txt

# Monitor cache metrics during test
curl http://localhost:8080/metrics/cache | jq

# Expected results:
# - hit_rate_pct > 70%
# - Database query count should be minimal
# - p95 latency < 150ms
```

### 3. Verify Specific Queries are Cached

```bash
# Start server with query logging
DB_LOG_QUERIES=true go run ./cmd/server 2>&1 | tee server.log &

# Run load test
k6 run loadtest/loadtest_me_realistic.js

# Count occurrences of the projects query
grep "FROM projects" server.log | wc -l
# Should be 1 (or very low number, not 3000+)

# Count occurrences of user queries
grep "FROM users" server.log | wc -l
# Should be ~34 (one per unique user, then cached)
```

---

## Cache Configuration

The cache is configured with sensible defaults in `internal/cache/cache.go`:

```go
NumCounters: 100_000      // Track 100k items
MaxCost:     100_000_000  // 100MB max cache size
BufferItems: 64           // 64 keys per buffer
DefaultTTL:  15 minutes   // Default expiration time
```

### Adjusting Cache Settings

To modify cache behavior, edit `internal/cache/cache.go` in the `DefaultConfig()` function:

```go
func DefaultConfig() Config {
    return Config{
        NumCounters: 200_000,     // Increase for more items
        MaxCost:     200_000_000, // Increase cache size
        BufferItems: 128,         // Increase for higher concurrency
        DefaultTTL:  30 * time.Minute, // Longer cache duration
    }
}
```

Or add environment variables (future enhancement).

---

## Cached Entities

The following entities are currently cached:

### Entity-Level Caching
- **Users** (15 min TTL) - `UserByIDLoader`
- **Churches** (30 min TTL) - `ChurchLoader`
- **Projects by User** (15 min TTL) - `ProjectsByUserLoader`
- **Projects by ID** (15 min TTL) - `ProjectByIDLoader`
- **User Roles** (15 min TTL) - `RolesByUserLoader`

### Query-Level Caching
- **All Projects** (15 min TTL) - Root `projects` query

---

## Troubleshooting

### Cache Hit Rate is Low (< 50%)

**Possible causes:**
1. Cache TTL too short - increase `DefaultTTL`
2. Cache size too small - increase `MaxCost`
3. User diversity too high - normal for tests with many different users
4. Keys being evicted - check `keys_evicted` metric

**Solution:**
```bash
# Check eviction stats
curl http://localhost:8080/metrics/cache | jq '.keys_evicted'

# If evictions are high, increase cache size in DefaultConfig()
```

### Database Still Shows High Query Count

**Check:**
1. Is query logging enabled? (`DB_LOG_QUERIES=true`)
2. Are you seeing the same query repeated multiple times?
3. Check which query is repeating

**Identify the query:**
```bash
# Filter and count queries
DB_LOG_QUERIES=true go run ./cmd/server 2>&1 | grep "Query" | grep "SELECT" | sort | uniq -c | sort -rn

# This shows most frequent queries
```

### Sets Dropped or Rejected

If `sets_dropped` or `sets_rejected` is high:

**Causes:**
- `sets_dropped`: Write buffer is full (increase `BufferItems`)
- `sets_rejected`: Items rejected by admission policy (rare, usually OK)

**Solution:**
```go
// Increase buffer size in DefaultConfig()
BufferItems: 128, // Was 64
```

---

## Example: Complete Load Test Session

```bash
# Terminal 1: Start server with query logging
DB_LOG_QUERIES=true go run ./cmd/server 2>&1 | tee server.log

# Terminal 2: Watch cache metrics
watch -n 1 'curl -s http://localhost:8080/metrics/cache | jq "{hits, misses, hit_rate_pct, keys_evicted}"'

# Terminal 3: Run load test
cd loadtest
./generate_tokens.sh
k6 run loadtest_me_realistic.js

# After test completes, analyze results:
# 1. Check k6 output for p95 latency (should be < 150ms)
# 2. Check cache hit rate (should be > 70%)
# 3. Count database queries
grep "Query" server.log | wc -l  # Should be low (< 100)
grep "FROM projects" server.log | wc -l  # Should be 1
```

Expected outcome:
```
✓ p95 latency: 89ms (target: < 150ms)
✓ Cache hit rate: 87.3% (target: > 70%)
✓ Total DB queries: 42 (vs 3000+ without cache)
✓ Projects queries: 1 (vs 3000+ without cache)
```
