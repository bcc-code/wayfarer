# Load Testing Quick Start

Quick commands to run load tests with monitoring.

## Prerequisites

```bash
# Install k6
brew install k6

# Generate JWT tokens for test users
./generate_tokens.sh
```

---

## Basic Load Test

```bash
# Terminal 1: Start server
go run ./cmd/server

# Terminal 2: Run realistic load test (3000 concurrent users)
k6 run loadtest_me_realistic.js
```

---

## Load Test with Cache Monitoring

```bash
# Terminal 1: Start server
go run ./cmd/server

# Terminal 2: Watch cache hit rate in real-time
watch -n 1 'curl -s http://localhost:8080/metrics/cache | jq "{hit_rate_pct, hits, misses, keys_evicted}"'

# Terminal 3: Run load test
k6 run loadtest_me_realistic.js
```

**Expected Results:**
- Cache hit rate: > 70%
- p95 latency: < 150ms
- Keys evicted: 0 (or very low)

---

## Load Test with Database Query Logging

```bash
# Terminal 1: Start server with query logging
DB_LOG_QUERIES=true go run ./cmd/server 2>&1 | grep "Query"

# Terminal 2: Run load test
k6 run loadtest_me_realistic.js
```

**What to look for:**
- First few requests: You should see queries (cache misses)
- After cache warms up: Very few queries (cache hits)
- The `projects` query should appear only ONCE

---

## Full Monitoring Setup

```bash
# Terminal 1: Server with query logging
DB_LOG_QUERIES=true go run ./cmd/server 2>&1 | tee server.log

# Terminal 2: Cache metrics
watch -n 1 'curl -s http://localhost:8080/metrics/cache | jq'

# Terminal 3: Run load test
k6 run loadtest_me_realistic.js

# After test, analyze:
grep "FROM projects" server.log | wc -l  # Should be 1
grep "Query" server.log | wc -l          # Should be < 100
```

---

## Available Load Tests

### loadtest_me_realistic.js
- **Users**: 3000 concurrent (34 unique user IDs)
- **Duration**: 70 seconds
- **Use case**: Realistic event start scenario
- **Expected p95**: < 150ms

```bash
k6 run loadtest_me_realistic.js
```

### loadtest_cache_analysis.js
- **Users**: 50 concurrent
- **Scenarios**: Low diversity (5 users) vs high diversity (34 users)
- **Use case**: Compare cache performance
- **Expected**: High hit rate in low diversity, lower in high diversity

```bash
k6 run loadtest_cache_analysis.js
```

### loadtest_extreme.js
- **Users**: 5000 concurrent
- **Duration**: 50 seconds
- **Use case**: Stress test / breaking point
- **Expected**: System should handle without errors

```bash
k6 run loadtest_extreme.js
```

---

## Interpreting Results

### Good Performance (Cache Working)
```
✓ http_req_duration: p95=89ms (< 150ms target)
✓ Cache hit rate: 87.3% (> 70% target)
✓ Database queries: 42 total
✓ Projects query: 1 occurrence
✓ http_req_failed: 0%
```

### Poor Performance (Cache Not Working)
```
✗ http_req_duration: p95=450ms (>> 150ms target)
✗ Cache hit rate: 15% (<< 70% target)
✗ Database queries: 3000+ total
✗ Projects query: 3000+ occurrences
✗ http_req_failed: 0-5%
```

---

## Endpoints for Monitoring

### Cache Metrics
```bash
curl http://localhost:8080/metrics/cache | jq
```

Returns:
```json
{
  "hits": 15234,
  "misses": 982,
  "hit_rate_pct": 93.94,
  "keys_evicted": 0
}
```

### Health Check
```bash
curl http://localhost:8080/health
```

---

## Troubleshooting

### "Connection refused" errors
**Solution:** Start the server first: `go run ./cmd/server`

### "JWT authentication errors"
**Solution:** Regenerate tokens: `./generate_tokens.sh`

### Low cache hit rate (< 50%)
**Check:**
1. Are you running multiple server instances? (Each has separate cache)
2. Is the server restarting between requests? (Cache is in-memory)
3. Are different users being tested? (Lower hit rate expected)

### High p95 latency (> 200ms)
**Check:**
1. Database query logging: Are same queries repeating?
2. Cache metrics: Is hit rate above 70%?
3. Server resources: Is CPU/memory maxed out?

**Debug:**
```bash
# Enable query logging to identify problematic queries
DB_LOG_QUERIES=true go run ./cmd/server 2>&1 | grep "Query" | sort | uniq -c | sort -rn
```

---

## Complete Testing Workflow

```bash
# 1. Clean start
pkill -f "go run ./cmd/server" # Kill any running servers
rm -f tokens.js                 # Remove old tokens
./generate_tokens.sh            # Generate fresh tokens

# 2. Start monitored server
DB_LOG_QUERIES=true go run ./cmd/server 2>&1 | tee server.log &
sleep 2  # Wait for server to start

# 3. Run test and monitor
watch -n 1 'curl -s http://localhost:8080/metrics/cache | jq' &
k6 run loadtest_me_realistic.js

# 4. Analyze results
echo "=== Cache Performance ==="
curl -s http://localhost:8080/metrics/cache | jq

echo "=== Database Query Analysis ==="
echo "Total queries: $(grep -c "Query" server.log)"
echo "Projects queries: $(grep -c "FROM projects" server.log)"
echo "User queries: $(grep -c "FROM users" server.log)"

# 5. Cleanup
pkill -f "watch"
pkill -f "go run"
```

---

## Documentation

For detailed information, see:
- **`docs/caching-and-monitoring.md`** - Complete monitoring guide
- **`loadtest/analyze_db_queries.md`** - Issues found and fixed
- **`loadtest/README.md`** - Load test descriptions
