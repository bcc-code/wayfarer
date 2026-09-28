# Load Test Performance — Round 2 (per-user query caching)

**Date:** 2026-07-13
**Source:** Jaeger traces from k6 load test (steady 50 VUs + 500-VU spike), `wayfarer-backend` service
**Follow-up to:** `challenges-performance-report.md`, `front-page-performance-report.md`

---

## 1. What the traces showed

Outside the spike the backend was healthy (POST /graphql p50 ~88ms). During two ~30s
bursts p50 jumped to 1.5–4s, throughput collapsed, `pool.acquire` hit 2.4–2.7s, and
even trivial indexed queries took 300ms–2s.

**Three uncached per-user queries were ~99% of all DB traffic** (~18k queries in a
4-minute window, essentially all returning 0 rows):

| Query | Count | Caller |
|-------|-------|--------|
| `GetUserTeamByProjectID` | ~9,000 | `Project.myTeam` resolver (every page) |
| `GetUserEnrolledChallengeIDsInProject` | ~4,600 | `getFilteredChallenges` |
| `GetBulkUserSessionAccessQuizIDs` | ~4,200 | `getFilteredChallenges` |

Everything else (challenge lists, projects, quizzes) already cache-hit. Under the
spike these queries flooded the 25-connection pool (sized for Neon — do not raise);
queuing cascaded into every other query.

Secondary findings:

- The shared dataloaders used the graph-gophers default **16ms batch window**;
  the challenges pipeline awaits 4–5 loaders sequentially → up to ~80ms of pure
  wait per request even when fully cached. Batch SQL also runs on the *first*
  batching caller's trace, so blocked requests show untraced self-time gaps.
- Leaderboards were cached as **JSON bytes**: every cache *hit* re-unmarshaled the
  full ~9k-entry board and linear-scanned for "me". No singleflight anywhere, so a
  hot key expiring under 500 VUs stampeded the DB.

## 2. Changes made (branch `perf/backend-p0-fixes`)

### Per-user lookup caches (negative results included)

New keys in `internal/cache/keys.go`, all registered under the `user:{userID}`
invalidation tag so the existing `InvalidateUser` calls in every membership /
enrollment mutation cover them:

- `UserTeamInProjectKey(userID, projectID)` → cached **team ID only** ("" = no team);
  team details still come from `TeamByIDLoader` (invalidated by `InvalidateTeam`),
  so team renames etc. stay visible without touching the membership cache.
  Used by `Resolver.getUserTeamInProject` (`internal/graph/api/teams.go`).
- `UserEnrolledChallengesKey(userID, projectID)` → enrolled challenge ID set.
  Also dropped globally by `InvalidateChallenge`.
- `UserQuizSessionAccessKey(userID, projectID)` → accessible quiz ID set with the
  covered quiz set, **45s TTL** (session states change over time). Dropped by
  `InvalidateChallenge`, `InvalidateQuiz`, and the new
  `InvalidateQuizSessionAccess()` (called from OpenQuizSession, Revoke*, and the
  bulk grant path; broadcast cross-instance via the new
  `InvalidationTypeQuizSessionAccess` message).

Helpers live in `internal/graph/api/challenges_user_cache.go`;
`getVisibleEventChallenges` shares the same caches (enrollment is project-wide).

### Pipeline latency

- Quiz-access and enrollment lookups in `getFilteredChallenges` now run
  **concurrently** (they are independent).
- `newBatchedLoader` sets `WithWait(2ms)` instead of the 16ms library default.

### Leaderboards (`internal/services/leaderboard.go`)

- All 8 board types now go through `getFullLeaderboardCached`: the cache stores the
  **decoded** `[]LeaderboardEntry` + a `userID→index` map (no JSON round-trip on
  hits; "me" lookup is O(1)).
- Cache-miss refills are guarded by **singleflight** keyed on the cache key; the
  fetch runs with `context.WithoutCancel` so a cancelled leading request cannot
  poison the shared result.
- The `teamByIDBatchFunc` loader conversion was missing `SuperTeamID` (latent bug:
  any team loaded through it lost its superTeam) — fixed.

## 3. Expected impact

- The three hot queries drop from ~18k per test window to roughly one per active
  user per TTL/invalidation — pool pressure during spikes disappears with them.
- Fully-cached challenge pages lose up to ~80ms of dataloader batch-window wait.
- Leaderboard cache hits are pure slice/map operations; expiry no longer stampedes.

## 4. Verification

- Unit tests: `internal/cache/user_lookup_cache_test.go`,
  `internal/graph/api/challenges_user_cache_test.go`,
  `internal/services/leaderboard_cache_test.go` (incl. singleflight concurrency
  and cancellation tests). `MockLeaderboardQuerier` added to `.mockery.yml`.
- `make fmt` + `make test` (with `-race`) pass. The e2e suite has **pre-existing
  flakiness** (TestAchievements/TestTeams/TestScoring fail intermittently on the
  base commit too, ~1 in 2 full runs) — unrelated to these changes.
- Re-run the k6 scenario and compare in Jaeger: query counts for the three hot
  queries, `pool.acquire` p95 during the spike, and POST /graphql p95.
