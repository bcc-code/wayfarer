# Load Test Performance — Round 3 (freetext quiz spike: batching + quiz write path)

**Date:** 2026-07-13
**Source:** Jaeger traces from the k6 freetext-quiz spike test (`make loadtest-quiz-freetext-spike`), `wayfarer-backend` service
**Follow-up to:** `loadtest-performance-round2.md`
**Branch:** `perf/backend-p0-fixes`

---

## 1. What the traces showed

The spike test sends N **distinct** users through cold load → ChallengePage →
startQuizSession → submitQuizAnswer → finalizeQuiz within seconds. All latency
was in the connection layer, not SQL (normal Neon statement RTT: 30–90ms):

- **Connection storm**: the pool opened ~22 new Neon connections at spike
  onset; each `connect` span took 5–9s (TLS/auth queueing). Everything else
  blocked in `pool.acquire` — `GetMe` averaged 8.5s.
- **Per-user cache misses fan out un-batched**: the round-2 per-user Ristretto
  caches never hit for distinct cold users, and each miss was a direct pool
  query. `GetUserTeamByProjectID` query spans averaged 1.7s during the spike
  (server-side queueing), one per user.
- **finalizeQuiz held a pool connection + `FOR UPDATE` row lock across ~10+
  sequential round-trips**, including a byte-identical duplicate aggregate
  (`CalculateSubmissionPointsFromResponses` == `CalculateSubmissionScore`) and
  static reads inside the tx.

Decision: pool warmup (P0) was considered and **deliberately dropped** — fix
by cutting query volume, not by tuning the pool (see
`db-pool-tuned-for-neon` constraint: never raise `DB_MAX_OPEN_CONNS`).

## 2. Changes made

### P1 — per-user lookups now batch through dataloaders

Three new loaders keyed by `UserProjectKey`, registered in
`internal/loaders/loaders.go`, all built on a shared
`batchUserProjectLookup` helper (`internal/loaders/user_project_lookups.go`):
check Ristretto per key inside the batch fn, group misses by project, one
grouped SQL per project, cache every result (negatives included), return in
key order. 100 cold users in a 2ms window now cost 1 query instead of 100.

| Loader | SQL (new bulk variant) | Cache key (unchanged) |
|---|---|---|
| `UserTeamIDInProjectLoader` | `GetUserTeamsByProjectIDBulk` (`DISTINCT ON`) | `UserTeamInProjectKey` |
| `UserEnrolledChallengeIDsLoader` | `GetUsersEnrolledChallengeIDsInProject` | `UserEnrolledChallengesKey` |
| `UserAccessibleQuizIDsLoader` | `GetBulkUsersSessionAccessQuizIDsByProject` | `UserQuizSessionAccessKey` (45s TTL) |

- Quiz access is now computed **project-wide** (the old `Covered`-map subset
  logic is gone); old-shape cached entries fail the type assertion and are
  treated as misses.
- `getUserTeamInProject` / `getUserEnrolledChallengeIDs` /
  `getUserAccessibleQuizIDs` in `internal/graph/api/` now just call the
  loaders; all invalidation semantics (`InvalidateUser`,
  `InvalidateChallenge`, `InvalidateQuizSessionAccess`) are unchanged.
- `UserHasAccessToVisibleSession` in `LoadChallengeWithVisibility`
  (`resolver.go`) was replaced by the batched/cached quiz-access lookup
  (same state filter + access join).

### P2 — quiz path

**finalizeQuiz** (`quizzes.resolvers.go`):
- Deleted `CalculateSubmissionPointsFromResponses` (byte-identical to
  `CalculateSubmissionScore`); total points derive from the score. The M2M
  `CreateQuizSubmission` path reuses its locally accumulated score.
- Static reads moved **before** `Begin`: non-locking submission pre-read
  (fail-fast validation, re-validated under the lock), quiz + challenge via
  loaders, quiz achievement criteria via a new cached helper. The tx is now
  FOR UPDATE select → score aggregate → UPDATE → conditional journal insert
  (~5 round-trips, and no loader awaits while holding the row lock).
- Achievement awarding stays **inside the transaction** (atomic with
  finalization — a failed award rolls back the completion so a retried
  finalize awards again) but uses `AwardUserAchievementIdempotent` (now
  `:execresult`; `RowsAffected()==0` means already-awarded, so push
  notifications still only fire for new awards). The old Check+Award pair
  per achievement is gone, so it's 1 round-trip per eligible achievement
  and the criteria rows are loaded (cached) before `Begin`.

**submitQuizAnswer**: question validation and answer correctness now come
from `QuizQuestionsByQuizLoader` / `QuizAnswersByQuestionLoader` (cached
static data) instead of per-request `GetQuizQuestionByID` /
`GetPredefinedAnswersByQuestionID`. The rare betting path still fetches the
sqlc question row for exact `pgtype.Numeric` bet limits. Free-text answer =
1 submission read + 1 INSERT.

**startQuizSession**: session row via `QuizSessionByIDLoader` (now stored
with 45s TTL) and questions via `QuizQuestionsByQuizLoader`. Access check +
existing-submission check + INSERT stay direct (correctness-critical).

**Quiz.userActiveSession** (every ChallengePage, previously uncached): new
`UserActiveQuizSessionLoader` keyed `UserQuizKey{UserID, QuizID}`, batch
grouped by quiz (`GetUsersActiveSessionForQuiz`, `DISTINCT ON (user_id)`),
cached 45s under the new `useractivesession:` prefix (nil = no session is
cached too).

### New invalidation plumbing (`internal/cache/`)

- `InvalidateQuizSession(sessionID)` — drops `QuizSessionKey`, broadcast via
  new `InvalidationTypeQuizSession`. Called from all six session mutations
  (update/delete/open/lock/finish/reopen) and the three scheduler transition
  loops in `internal/handlers/quiz_scheduler.go` (which now also call
  `InvalidateQuizSessionAccess`).
- `invalidateQuizSessionAccessLocal` and `invalidateQuizLocal` also drop the
  `useractivesession:` prefix; `invalidateQuizLocal` drops the new
  `QuizAchievementsByQuizKey`. Quiz-achievement create/update/delete
  mutations call `InvalidateQuiz` for affected quiz IDs.
- New prefixes are registered in `ExtractUserTag` **and** in
  `extractPrefixes` (invalidation.go) — a prefix missing from
  `extractPrefixes` is silently unregistered and `DeletePrefix` won't find
  it (this bit us in tests; remember for future keys).

## 3. Cold-user cost per request after this round

| Operation | Before | After |
|---|---|---|
| GetMe (cold) | 2 direct statements | unchanged (batched via user loader) |
| ActiveChallengesPage (cold) | 3 un-batched per-user queries | 3 batched (≈3 queries per 100 users) |
| ChallengePage | 3–4 per-user, `userActiveSession` uncached | batched + 45s cached |
| startQuizSession | 5 statements | 3 (access check, submission check, INSERT) |
| submitQuizAnswer (free text) | 3 statements | 2 (submission read, INSERT) |
| finalizeQuiz | tx with ~10+2N round-trips under FOR UPDATE | 1 pre-read + tx of ~5+N (N = eligible achievements, usually 0) |

The connection storm at spike onset (5–9s `connect`) is **still present** —
P0 (pool warmup / MinConns) was explicitly dropped. If a future spike test
still breaches latency targets purely on cold-start connects, that's the
remaining lever.

## 4. Verification

- `make fmt` + `make test` green; quiz e2e (`go test ./e2e/ -run TestQuiz`)
  green.
- New unit tests: `internal/loaders/user_project_lookups_test.go`,
  `internal/loaders/user_active_quiz_session_test.go`, extended
  `internal/cache/user_lookup_cache_test.go`, rewritten
  `internal/graph/api/challenges_user_cache_test.go`.
- Real validation: re-run the spike against a test env and compare Jaeger —
  `GetUserTeamByProjectID` should collapse into a few batched queries and
  per-request statement counts drop as above.
