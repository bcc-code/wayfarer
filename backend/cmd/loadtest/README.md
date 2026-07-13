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
| `make loadtest-challenges` | Challenges page test: 50 VUs for 5 minutes |
| `make loadtest-challenges-quick` | Quick challenges page test: 10 VUs for 1 minute |
| `make loadtest-quiz` | Quiz load test: 20 VUs for 5 minutes |
| `make loadtest-quiz-quick` | Quick quiz test: 5 VUs for 1 minute |
| `make loadtest-quiz-stress` | Quiz stress test: 10 RPS for 5 minutes |
| `make loadtest-gen` | Generate tokens only (1000 users) |
| `make loadtest-gen-all` | Generate tokens for all users (up to 10k) |

## Test Scenarios

### 1. Steady Load (`steady_load`)
Each iteration simulates one realistic user session, mirroring the queries the
SPA actually fires (see `k6/lib/journey.js`):

1. **Cold app load** on the home page — the bootstrap queries every session
   fires once (`GetMe`, `CurrentProject`, `GetFirebaseToken`) plus
   `ProfilePage`.
2. **Challenges and standings pages** visited in random order (SPA navigation,
   so only page queries fire):
   - Challenges: `ActiveChallengesPage`; ~40% of users open the completed tab
     (`CompletedChallengesPage`); ~50% click a random challenge — external
     challenges link straight out of the app (no request), all other types
     fire `ChallengePage`.
   - Standings: `StandingsPage` wrapper + a random tab (40% global, 30% local,
     30% unit).

### 2. Spike Test (`spike_test`)
Simulates sudden traffic surges using the same user journey:
- Ramps from 0 to 100 VUs in 30s
- Peaks at 500 VUs for 1 minute
- Ramps down to 0 over 1 minute

### 3. Leaderboard Stress (`leaderboard_stress`)
Focused testing of database-intensive leaderboard queries:
- Constant 100 requests/second
- 50% global standings, 30% local, 20% team

### 4. Challenges Page (`challenges-scenario.js`)
Focused test of the challenges page flow. Each iteration:
1. Cold load of `/challenges`: `GetMe`, `CurrentProject`,
   `ActiveChallengesPage`, `GetFirebaseToken`
2. User scans the list (2–5s), then clicks one random challenge:
   - `ExternalChallenge` → leaves the app, no further requests
   - any other type → `ChallengePage`, then leaves the app

### 5. Quiz Load Test (`quiz-scenario.js`)
Tests the complete quiz flow: get quiz details, start quiz, answer all questions, finalize.
- `quiz_completion`: Steady load of users completing quizzes
- `quiz_spike`: Spike test with many concurrent quiz takers
- `quiz_stress`: High-frequency quiz completions

**Prerequisites for Quiz Tests:**
1. Insert the test quiz into the database:
   ```bash
   # Set your project and challenge IDs, then run:
   psql $DATABASE_URL -v project_id="'YOUR_PROJECT_ID'" -v challenge_id="'YOUR_CHALLENGE_ID'" \
     -f ./cmd/loadtest/scripts/insert_quiz.sql
   ```
2. The quiz ID defaults to `QZ01ARQN6LOADTEST00000QUIZ`, or set via `QUIZ_ID` env var

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
| `CHALLENGES_VUS` | 50 | Virtual users for challenges page test |
| `SKIP_FIREBASE_TOKEN` | _(unset)_ | Set to skip the `GetFirebaseToken` bootstrap query (for environments without Firebase credentials) |
| `QUIZ_ID` | QZ01LOADTESTQUIZ000000000000 | Quiz ID to test |
| `QUIZ_VUS` | 20 | Virtual users for quiz test |
| `QUIZ_RPS` | 10 | Requests per second for quiz stress test |
| `STRESS_DURATION` | 5m | Duration of quiz stress test |

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
        scenarios.js     # Combined test script (steady + spike + leaderboard)
        steady.js        # Steady load scenario
        spike.js         # Spike scenario
        leaderboard.js   # Leaderboard stress scenario
        challenges-scenario.js # Challenges page scenario
        quiz-scenario.js # Quiz-specific test scenarios
        lib/
            graphql.js   # GraphQL HTTP helpers
            journey.js   # Realistic user session journey
        queries/
            bootstrap.js # Cold-load queries (GetMe, CurrentProject, GetFirebaseToken)
            challenges.js
            profile.js
            standings-page.js
            standings-global.js
            standings-local.js
            standings-unit.js
            challenge.js
            quiz.js      # Quiz query/mutation helpers
    scripts/
        insert_quiz.sql  # SQL script to insert test quiz
    config.json          # Generated tokens (gitignored)
    README.md
```
