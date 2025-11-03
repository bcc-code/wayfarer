# Load Testing

This directory contains k6 load tests for the Wayfarer GraphQL API.

## Prerequisites

Install k6:
```bash
# macOS
brew install k6

# Or download from https://k6.io/docs/getting-started/installation/
```

## Load Test Files

### loadtest_me.js
Original simple load test using a single user ID. Useful for establishing baseline performance.

**Usage:**
```bash
k6 run loadtest_me.js
```

**Characteristics:**
- 1000 VUs for 10 seconds
- Single user ID (high cache hit rate)
- Tests maximum throughput with optimal caching

### loadtest_me_realistic.js
Realistic load test simulating event start with thousands of concurrent users.

**Usage:**
```bash
k6 run loadtest_me_realistic.js
```

**Characteristics:**
- 34 different user IDs (simulates real user diversity)
- Rapid spike: 0 → 3000 users in 10 seconds
- Realistic think time (0.5-2 seconds between requests)
- Models youth camp event start scenario

**Load profile:**
1. Rapid ramp: 0 → 3000 users (10s)
2. Sustained load at 3000 users (30s)
3. Drop to 1000 users (20s)
4. Ramp down (10s)

**Total duration:** 70 seconds

### loadtest_cache_analysis.js
Specialized test for analyzing DataLoader cache behavior.

**Usage:**
```bash
k6 run loadtest_cache_analysis.js
```

**Characteristics:**
- Two scenarios running sequentially
- **Low diversity** (0-30s): 50 VUs using only 5 users → High cache hit rate
- **High diversity** (35-65s): 50 VUs using all 34 users → Lower cache hit rate

Use this to compare performance between high and low cache hit scenarios.

### loadtest_extreme.js
Extreme load test pushing system to maximum capacity.

**Usage:**
```bash
k6 run loadtest_extreme.js
```

**Characteristics:**
- Tests system breaking point
- 0 → 5000 users in 10 seconds
- Minimal sleep between requests (50-150ms)
- Simplified query for maximum throughput

**Load profile:**
1. Quick ramp to 1000 users (5s)
2. Extreme spike to 5000 users (5s)
3. Sustain 5000 users (20s)
4. Drop to 2000 users (10s)
5. Ramp down (10s)

## Authentication Setup

All load tests require valid JWT tokens for authentication.

### Generating JWT Tokens

You need to generate JWT tokens for each test user. The tokens should contain the user ID in the payload.

Example token generation (adjust based on your JWT configuration):
```go
// In a separate tool or test file
import (
	"github.com/golang-jwt/jwt/v5"
)

func generateTestToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key"))
}
```

### Using Tokens in Load Tests

Update the load test files with actual tokens:

```javascript
// In loadtest_me_realistic.js
const tokensByUserID = {
  'US01K8XV6EN42TPEATDT708X51KE': 'eyJhbGc...',
  'US01K8XV6EPTGRWZPZKV3SZ3MTRP': 'eyJhbGc...',
  // ... add tokens for all user IDs
};
```

## Profiling During Load Tests

To profile the server during load testing, use Go's built-in pprof HTTP endpoint:

```bash
# Terminal 1: Start server (pprof automatically exposed on port 6060 in development)
go run ./cmd/server

# Terminal 2: Run load test
k6 run loadtest_me_realistic.js

# Terminal 3: Capture profile while test is running
# CPU profile (30 seconds):
curl "http://localhost:6060/debug/pprof/profile?seconds=30" > cpu.prof

# Memory profile:
curl http://localhost:6060/debug/pprof/heap > mem.prof

# Goroutine profile:
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# Analyze profiles:
go tool pprof -top cpu.prof
go tool pprof -web cpu.prof  # Opens visualization in browser
```

## Analyzing Results

k6 provides metrics after each run:

- **http_req_duration**: Request latency (p50, p95, p99)
- **http_req_failed**: Error rate
- **http_reqs**: Total requests and requests/second
- **vus**: Number of virtual users

### Key Metrics to Watch

- **p95 latency < 200ms**: 95% of requests complete under 200ms
- **p99 latency < 500ms**: 99% of requests complete under 500ms
- **Error rate < 1%**: Less than 1% of requests fail

### Comparing Cache Performance

Run the cache analysis test and compare:
- Low diversity scenario latency (high cache hits)
- High diversity scenario latency (lower cache hits)

The difference shows the impact of cache hit rate on performance.

## Expected Performance

With DataLoader implementation:

- **Single user (100% cache hit)**: p95 < 50ms
- **Multiple users (20-40% cache hit)**: p95 < 150ms
- **Cold start (0% cache hit)**: p95 < 200ms

## Troubleshooting

### Connection Refused
Server is not running. Start it with:
```bash
go run ./cmd/server
```

### JWT Authentication Errors
Either disable JWT auth for testing or provide valid tokens.

### High Latency
Check:
1. Database connection pooling configuration
2. Database query performance
3. DataLoader batching behavior
4. Server resource utilization
