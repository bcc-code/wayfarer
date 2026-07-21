# Load Test Performance — Round 4 (JoinProject / church-leaderboard triggers)

**Date:** 2026-07-14
**Source:** Neon query insights from the freetext-quiz spike test + EXPLAIN ANALYZE on the Neon `development` branch
**Follow-up to:** `loadtest-performance-round3.md`

---

## 1. Symptom

`JoinProject` (`INSERT INTO user_projects ... ON CONFLICT DO NOTHING`) averaged
**67.9ms server-side over 9,999 calls (679s total)** during the spike test. In
the QR-code entry flow every `enrollInChallenge` is a first-time join, so every
call was a real insert.

## 2. Root cause

The bare INSERT (PK + 2 secondary indexes + 2 FK checks) is sub-millisecond.
The cost was the `AFTER INSERT FOR EACH ROW` trigger
`trigger_user_projects_recalculate_church_average` (created in migration 00080,
body from 00081), which per join executed:

1. `users` PK lookup for church_id — cheap.
2. `SUM(points)` over `score_journal` for (user, project) — indexed but rows
   are scattered across heap pages; measured **64.9ms cold** on Neon (40
   pageserver reads), sub-ms warm.
3. `COUNT(DISTINCT u.id)` join recomputing the church's member count —
   **seq-scanned all 10,588 `user_projects` rows** + hash join + sort; 5.7ms
   warm, growing linearly with project size.
4. UPSERT into `leaderboard_project_churches` — the per-(project, church) row
   lock serialized all concurrent joiners from the same church (largest test
   church: 640 users); lock waits count toward statement exec time.

The same `COUNT(DISTINCT)` recompute pattern existed in
`trigger_recalculate_averages_on_event_membership_change` (event joins) and in
`update_church_leaderboard`, which runs on **every `score_journal` insert**
(every point award) via `trigger_update_leaderboard_from_score_journal`.

## 3. Fix — migration `00101_optimize_church_leaderboard_triggers.sql`

New invariant: **membership triggers own `member_count` (±1 incrementally);
the score path owns `total_points`; both recompute `score` from stored
values.** The `COUNT(DISTINCT)` now runs only when a (project|event, church)
leaderboard row is first created (UPDATE-first, INSERT-fallback pattern — the
count must not sit in a plain upsert's VALUES because `ON CONFLICT` evaluates
VALUES before conflict detection).

Changes:

- `trigger_recalculate_church_average_on_project_membership_change` — hot path
  is one UPDATE (`member_count ± 1`, `total_points ±` the user's journal sum).
  The old `points > 0` guard is gone: rows are created/maintained even at 0
  points so incremental counts stay correct. Reads filter
  `score >= COALESCE(minscore, 1)` (`queries/leaderboards.sql`), so 0-point
  rows are never returned to clients.
- `trigger_recalculate_averages_on_event_membership_change` — same treatment
  for event churches, event teams, and event superteams. (Safe because a user
  has at most one team per project — `joinTeam` removes prior memberships.)
- `update_church_leaderboard` — reuses the stored `member_count`; only the
  row-missing fallback computes a count.
- `regenerate_leaderboards()` — church sections now source rows from
  membership (LEFT JOIN score_journal) so every church with members gets a row
  with a correct `member_count`; members' points only (a user who left no
  longer contributes). Called at the end of the Up section as the one-time
  backfill. Down restores the 00080/00081 bodies and re-runs regenerate.

Known trade-offs:

- A concurrent create-race on the same new (project, church) row can overcount
  by 1; `regenerate_leaderboards()` self-heals. The same class of drift already
  existed (e.g. a user changing church never re-triggered a recount).
- The `score_journal` SUM on join stays (correctness-required, indexed); its
  cold-page cost on Neon is inherent.
- Row-lock contention on the church row remains, but hold time drops from
  ~70ms to ~1ms, collapsing the queue.

## 4. Verification

- E2E: `membership_leaderboard_test.go` — existing points tests plus new
  `TestChurchLeaderboardMemberCount` (first-join row creation, increment /
  decrement, score-award count stability, event membership, consistency with
  `regenerate_leaderboards()`). New testutil getters
  `GetLeaderboardProjectChurchMemberCount` / `GetLeaderboardEventChurchMemberCount`.
- Real validation: after deploying + running the migration, re-run
  `make loadtest-quiz-freetext-spike` and compare Neon query insights —
  `JoinProject` mean should drop from ~68ms to low single-digit ms;
  `score_journal` INSERT mean drops similarly.

## 5. Pre-existing e2e flake found along the way (not fixed here)

`cmd/seed/seeders/projects.go:107-110` generates project start/end dates with
the **global, unseeded** `math/rand` while the rest of the seeder uses the
seeded faker. Seeded project end-date offsets land in [-150, +305] days, so
~30% of `go test ./e2e/` processes draw an already-ended `data.ProjectIDs[0]`
and any award-related test fails with "project has ended and is no longer
accepting achievement awards" (`internal/services/project_helpers.go:20`).
Reproduced on clean `main`: 3 of 5 full-suite runs failed (TestEvents,
TestAchievements, TestScoring variants). Fix idea: thread the seeded rng into
`SeedProjects` (note this changes `make seed` output too, and a fixed seed
could then deterministically pick an ended project — pick offsets so the
first projects are active, or have e2e select an active project).
