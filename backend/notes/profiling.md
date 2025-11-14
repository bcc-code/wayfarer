# Performance Profiling Guide

This document describes how to profile the Wayfarer backend to identify performance bottlenecks.

## Overview

The server exposes pprof endpoints at `/debug/pprof` for runtime profiling. These endpoints allow you to capture CPU profiles, memory profiles, and other runtime metrics while the server is running.

## Quick Start

1. Start your server:
   ```bash
   go run cmd/server/main.go
   ```

2. Generate load on the server (make some GraphQL requests)

3. Capture profiles using the helper script:
   ```bash
   ./scripts/profile.sh all 30
   ```

4. Analyze the results:
   ```bash
   go tool pprof -http=:8081 profiles/<timestamp>/cpu.prof
   ```

## Profile Types

### CPU Profile
Identifies which functions consume the most CPU time. Use this to find computational bottlenecks.

```bash
# Capture 30 seconds of CPU profile
./scripts/profile.sh cpu 30

# Or manually:
curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof
```

### Heap Profile
Shows current memory allocations. Use this to identify memory leaks or excessive allocations.

```bash
# Capture heap snapshot
./scripts/profile.sh heap

# Or manually:
curl http://localhost:8080/debug/pprof/heap -o heap.prof
```

### Allocations Profile
Tracks all memory allocations (not just what's currently allocated). Use this to find allocation hotspots.

```bash
./scripts/profile.sh allocs
```

### Goroutine Profile
Shows all currently running goroutines. Use this to identify goroutine leaks or deadlocks.

```bash
./scripts/profile.sh goroutine
```

## Analyzing Profiles

### Interactive Web UI (Recommended)
```bash
go tool pprof -http=:8081 profiles/<timestamp>/cpu.prof
```

This opens a browser with:
- Flame graph visualization
- Top functions by time/memory
- Call graph
- Source code view

### Command Line Analysis

#### Top functions:
```bash
go tool pprof -top profiles/<timestamp>/cpu.prof
```

#### Text report:
```bash
go tool pprof -text profiles/<timestamp>/cpu.prof
```

#### List specific function:
```bash
go tool pprof -list=FunctionName profiles/<timestamp>/cpu.prof
```

### Generate Static Reports

#### Flame graph (SVG):
```bash
go tool pprof -svg profiles/<timestamp>/cpu.prof > cpu_flamegraph.svg
```

#### PDF report:
```bash
go tool pprof -pdf profiles/<timestamp>/cpu.prof > cpu_report.pdf
```

## Profiling Workflow

1. **Identify slow operation**: Use application logs, metrics, or user reports

2. **Reproduce under load**:
   - Use actual traffic patterns if possible
   - Or create a load test that simulates the slow operation
   - Example: Run GraphQL queries in a loop while profiling

3. **Capture profile during load**:
   ```bash
   # Start load in one terminal
   for i in {1..1000}; do
     curl -X POST http://localhost:8080/graphql \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer $TOKEN" \
       -d '{"query":"{ your slow query here }"}' &
   done

   # In another terminal, capture profile
   ./scripts/profile.sh all 30
   ```

4. **Analyze results**:
   ```bash
   go tool pprof -http=:8081 profiles/<timestamp>/cpu.prof
   ```

5. **Look for**:
   - Functions with high "flat" time (time spent in the function itself)
   - Functions with high "cum" time (time spent in function + callees)
   - Unexpected allocations in hot paths
   - Database query patterns (look for sqlc generated functions)

6. **Common bottlenecks in this codebase**:
   - N+1 queries (check DataLoader usage)
   - Missing database indices
   - Inefficient GraphQL resolvers
   - Large result sets loaded into memory
   - Cache misses (check `/metrics/cache` endpoint)

## Database Query Profiling

For SQL-specific profiling:

1. Enable PostgreSQL query logging:
   ```sql
   ALTER DATABASE wayfarer SET log_statement = 'all';
   ALTER DATABASE wayfarer SET log_duration = on;
   ALTER DATABASE wayfarer SET log_min_duration_statement = 100; -- log queries > 100ms
   ```

2. Check slow queries in logs:
   ```bash
   docker logs wayfarer-db 2>&1 | grep "duration:"
   ```

3. Use EXPLAIN ANALYZE in psql:
   ```sql
   EXPLAIN ANALYZE SELECT ...;
   ```

## Tips

- Profile during realistic load, not idle server
- CPU profiles need at least 10-30 seconds to be meaningful
- Compare profiles before and after optimization
- Focus on the hot path first (what's called most often)
- Check cache hit rate at `/metrics/cache` before profiling
- Remember: premature optimization is the root of all evil - profile first, optimize second

## Troubleshooting

### "No samples" or empty profile
- Server needs to be under load while profiling
- Increase profiling duration
- Make sure you're hitting the slow code path

### Profile looks normal but still slow
- Check database query performance separately
- Look at network latency
- Check external API calls (Members API, Auth0)
- Review cache configuration

### High memory usage
- Compare heap and allocs profiles
- Look for leaked goroutines
- Check for large objects held in memory
- Review cache eviction policy
