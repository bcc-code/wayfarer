# Load Testing

k6-based load testing for the Wayfarer GraphQL API.

## Prerequisites

1. Install k6:
   ```bash
   brew install k6  # macOS
   # or see https://k6.io/docs/getting-started/installation/
   ```

2. Seed the database with test users:
   ```bash
   make seed-large  # Creates 10k users
   ```

3. Start the server:
   ```bash
   make dev
   ```

## Quick Start

```bash
# Run the default load test (50 VUs, 5 minutes)
make loadtest

# Run a quick sanity check (10 VUs, 1 minute)
make loadtest-quick
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make loadtest` | Default load test: 50 VUs for 5 minutes |
| `make loadtest-quick` | Quick test: 10 VUs for 1 minute |
| `make loadtest-stress` | Stress test: 200 VUs for 10 minutes |
| `make loadtest-spike` | Spike test only: ramp 0 -> 500 -> 0 VUs |
| `make loadtest-leaderboard` | Leaderboard stress: 100 requests/second |
| `make loadtest-gen` | Generate tokens only (1000 users) |
| `make loadtest-gen-all` | Generate tokens for all users (up to 10k) |

## Test Scenarios

### 1. Steady Load (`steady_load`)
Simulates typical user traffic with weighted distribution:
- 30% ChallengesPage
- 20% ProfilePage
- 20% StandingsGlobalPage
- 15% StandingsLocalPage
- 15% StandingsUnitPage

### 2. Spike Test (`spike_test`)
Simulates sudden traffic surges:
- Ramps from 0 to 100 VUs in 30s
- Peaks at 500 VUs for 1 minute
- Ramps down to 0 over 1 minute

### 3. Leaderboard Stress (`leaderboard_stress`)
Focused testing of database-intensive leaderboard queries:
- Constant 100 requests/second
- 50% global standings, 30% local, 20% team

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `STEADY_VUS` | 50 | Virtual users for steady-state test |
| `DURATION` | 5m | Duration of steady-state test |
| `SPIKE_START` | 6m | When spike test starts |
| `SPIKE_PEAK` | 500 | Peak VUs during spike |
| `LEADERBOARD_START` | 0s | When leaderboard test starts |
| `LEADERBOARD_DURATION` | 5m | Duration of leaderboard test |
| `LEADERBOARD_RPS` | 100 | Requests per second |

### Token Generator Flags

```bash
go run ./cmd/loadtest/tokengen \
  -limit 1000 \           # Number of users
  -output config.json \   # Output file
  -base-url http://localhost:8080 \  # API URL
  -valid-days 7 \         # Token validity
  -secret "your-jwt-secret"  # Override JWT_SECRET
```

## Thresholds

The test will fail if:
- p95 response time > 500ms
- p99 response time > 1000ms
- Error rate > 1%
- GraphQL error rate > 1%

## Custom Test Runs

```bash
# Run with custom parameters
k6 run --env STEADY_VUS=100 --env DURATION=10m ./cmd/loadtest/k6/scenarios.js

# Run against a different target
go run ./cmd/loadtest/tokengen -base-url https://staging.api.example.com -output ./cmd/loadtest/config.json
k6 run ./cmd/loadtest/k6/scenarios.js

# Output JSON results
k6 run --out json=results.json ./cmd/loadtest/k6/scenarios.js
```

## Output Metrics

Key metrics to watch:
- `http_req_duration` - Request latency (p50, p95, p99)
- `http_req_failed` - Failed request percentage
- `graphql_errors` - GraphQL-specific errors
- `http_reqs` - Total requests per second
- `vus` - Active virtual users

## Files

```
cmd/loadtest/
    tokengen/
        main.go          # Go tool to generate JWT tokens
    k6/
        scenarios.js     # Main test script
        lib/
            graphql.js   # GraphQL HTTP helpers
        queries/
            challenges.js
            profile.js
            standings-global.js
            standings-local.js
            standings-unit.js
            challenge.js
    config.json          # Generated tokens (gitignored)
    README.md
```
