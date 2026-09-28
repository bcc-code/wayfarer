# A/B functional-equivalence verification: `worktree-ram-first` vs `main` (2026-08-25)

## Verdict

**The branch is functionally identical to main in all resolver logic, write paths,
scoring, and plugin behavior — but NOT in serving behavior: the always-on
whole-response cache introduces three demonstrated correctness bugs (one of them a
cross-user data leak), and the narrowed enrollment invalidation causes one
demonstrated stale read.** Every divergence found was reproduced deterministically
and is attributable to those two changes; nothing else differs.

| Area | Result |
|---|---|
| Resolver logic: filters (all 9 `LeaderboardFilter` fields + combined), leaderboards (PERSONS/TEAMS/SUPER_TEAMS/CHURCHES), pagination round-trips, profile, challenges, quiz reads | **identical** (61-step cold battery ×2 runs, plus A-vs-A control, 0 diffs) |
| Writes: enroll, full quiz lifecycle (free-text/multi/number), ordering + betting quiz (authored via admin mutations), score adjustments | **identical** (all steps matched; DB end-state identical) |
| Plugin (Ladder to Heaven): team-rename webhook (9 members × 300 pts), content-event webhook (deadline points), quiz-finalized webhook (challenge completion), superteam preview | **identical** |
| Database end state after all writes | **identical** — 16/16 deterministic projections (score_journal count/sum/digest, enrollments, completions, quiz submissions/responses, session access, all 4 leaderboard aggregate tables, teams, achievements) |
| Response serving under the branch's 30s whole-response cache | **NOT identical** — bugs B1–B3 below |
| `GET /metrics/http` | new unauthenticated endpoint on the branch (404 on main, 200 on branch) |

## Demonstrated divergences (each with pinned repro in the results archive)

### B1 — Cross-user data leak (severity: critical)
`response_cache.go` deems a query cacheable when every **top-level** field is
`currentProject`/`myCurrentProject`, but those types have user-scoped children.
Key is `sha256(rawQuery)+language` — no user identity.

Reproduced (probe P1): user 2 requests `myCurrentProject { leaderboard { me … } }`;
within 30s user 3 sends the byte-identical query text and **receives user 2's `me`**
(`me.id = US…02`, caller `US…03`). Also observed organically across the warm read
battery: user 3's `ActiveChallengesPage` returned **user 2's team joinCode**
(`JC000003` instead of `JC000004`), and users 100/500 likewise received foreign
`myTeam`, challenge lists, and standings. Affected shipped frontend queries:
`ActiveChallengesPage`, `CompletedChallengesPage`, `StandingsPage`,
`StandingsUnitPage`, `StandingsGlobalPage`, `CurrentProject`.

### B2 — GraphQL variables excluded from the cache key (severity: high)
Reproduced (probe P2): same query text with `{"first": 5}` then `{"first": 50}`
returns **5 edges both times** on the branch. Organically: `StandingsGlobalPage`
for `TEAMS`/`SUPER_TEAMS`/`CHURCHES` and **every filtered variant** returned the
first caller's PERSONS/unfiltered result for 30s (verified: TEAMS request answered
with US-prefixed person nodes).

### B3 — Stale read after challenge enrollment (severity: medium)
`EnrollInChallenge` now invalidates only 5 exact keys (`InvalidateUserChallengeEnrollment`)
instead of `InvalidateUser` + `InvalidateChallenge`. Reproduced (probe P3): enroll,
then immediately re-read `ActiveChallengesPage` — the branch still returns
`userEnrolledAt: null` while main shows the enrollment. `ChallengePage` and
`ProfilePage` stayed fresh (their keys are among the 5 cleared). Note: in this
single-instance test the stale response is (at least) served by the whole-response
cache; **the removed `InvalidateUser` broadcast additionally means multi-replica
deployments would keep stale user caches on other pods until TTL** — that failure
mode cannot manifest on a single instance and is asserted from code
(`invalidation.go:381-390` documents that callers must pair it with the broadcast).

### B4 — 30s staleness window for admin edits (severity: low, inherent to design)
Reproduced (probe P4): admin updates the project info message; a previously-warmed
`myCurrentProject` read stays stale on the branch for up to 30s and is fresh after
TTL expiry. Nothing invalidates `gqlresponse:*` keys.

### B5 — Unauthenticated `GET /metrics/http` (severity: low)
404 on main, 200 on the branch with per-route counts/latency percentiles. Same
posture as the pre-existing `/metrics/cache`, but new surface.

Not bugs, verified equivalent or intentional: Firebase token warmer (pre-warm only;
`firebaseToken` redacted-but-nonempty on both sides), gin access log removal in
production (log-only), HTTP stats middleware (passive), quiz-lookup-via-dataloader
on enroll (no observable difference), `bench/optimize_award_path.sql` (was applied
manually to the old bench DB only; **not** part of the branch build — this test ran
both builds against the unmodified schema, and the leaderboard aggregates came out
identical).

## Method

- Baseline A = `main` @ `411fc319`; candidate B = `worktree-ram-first` @ `3057e17b`,
  frozen as pushed tag **`ram-first-2026-08-25`**.
- Both built on the test box (49.12.121.62) as `wayfarer-a` (:8080) / `wayfarer-b`
  (:8081), `ENVIRONMENT=production`, against `wayfarer_diff_a` / `wayfarer_diff_b` —
  the latter a byte-identical `CREATE DATABASE … TEMPLATE` clone of the seeded
  former (13,162 users, 790 teams, 179 churches, 484,492 score_journal rows from
  `bench/gendata*.sql`, loadtest quiz fixture at 25 completion points, plus plugin
  fixtures: content achievement + external content `task-diff-1`, rename challenge).
- Harness: `backend/cmd/loadtest/diffcmp/` — lockstep A→B replay of the k6 query
  corpus and scripted write scenarios, with structural JSON diffing (runtime-ULID
  and timestamp normalization, per-side captured variables), per-side assertions,
  and three cache modes (nonce'd cold / repeated warm / verbatim). Services were
  restarted between phases and probes so the branch cache was provably empty at
  each probe start. An A-vs-A control run of the read battery matched 100%,
  validating the normalization. DB oracle: `oracle.sh` (16 projections, runtime
  ULIDs collapsed, timestamps excluded, ordered md5 digests).
- Scenario matrix: 61-step read battery (cold + warm + post-write rerun), pagination
  round-trips, W1–W8 write scenarios, P1–P5 pinned probes.
- Full raw evidence (every A/B response body, per-step diffs, verdicts):
  `results-final.tar.gz` on the box under `/opt/wayfarer-diff/` and locally under
  `backend/cmd/loadtest/results/` (not committed).

## Environment gotchas found on the way (recorded for reruns)

- The committed `bench/gendata.sql` references consent `mkid('CO',1)`, but
  `consents.id` requires the `CN` prefix — the original bench DB had a pre-inserted
  `CN…01` consent. `setup_box.sh` now pre-inserts it and remaps.
- A fresh DB's `settings.current_project_id` (seeded by migration) points at a dev
  project; the server panics at boot until it is updated to `PR…0001`.
- `@requireRole` resolves roles from `user_roles`, not the JWT — admin mutations
  need an `ADMIN` row for the harness admin user.
- The e2e suite on main (`go test ./e2e/`) is intermittently flaky when run as a
  full package (`TestScoring` / `TestAchievements` fail occasionally, pass in
  isolation and on rerun) — pre-existing, unrelated to this work.

## Recommendation

Before this branch (or a successor) ships, B1/B2 require the response-cache
eligibility check to walk the full selection set (or an allowlist of provably
user-independent operations) and the key to include the variables hash + user
scope; B3 requires restoring the `InvalidateUser` broadcast (or an equivalent
cross-replica invalidation) on enroll; B5 wants auth or a bind to localhost. The
perf mechanisms themselves (token warmer, member_count reuse, stats middleware)
introduced no functional drift.

## Fix verification (2026-08-25, later the same day)

All five divergences were fixed on `worktree-ram-first` in commit `9bdbb917`
("fix(cache): correct response-cache keying, enrollment broadcast, metrics
exposure") and re-verified on the same box setup with the fixed build deployed
as side B (results in `/opt/wayfarer-diff/results-fixpass/`):

| Check | Result |
|---|---|
| Warm read battery (verbatim repeats across users, all filters/variables) | **61/61 MATCH** (was 25 divergences) |
| P1 cross-user leak | no longer reproduces — each user gets their own `me`/joinCode |
| P2 variable collision | no longer reproduces — `first: 5` then `first: 50` returns 5 then 50 |
| P3 stale-after-enroll (fresh user) | no longer reproduces — all post-enroll rereads fresh |
| P4 TTL staleness on admin edit | no longer reproduces — project invalidation clears response entries |
| P5 `/metrics/http` | externally 404 on both sides (loopback-only on the branch); on-box access retained |
| Perf smoke (`freetext-quiz-rampspike`, RAMP_SCALE=0.05, THINK_SCALE=0.2, k6 on-box, fresh caches, disjoint user slices) | main: avg 9.35ms / p95 15.56ms; fixed branch: avg 9.25ms / p95 15.83ms — within noise; 502/502 iterations, 0 failures, 100% checks on both; sub-ms minimums on the branch confirm the (now per-user) response cache still serves hits |

Fix mechanics: response-cache keys now include a variables hash and — whenever
the selection (walked recursively, fragments included) touches any field not on
a conservative user-independent allowlist — the calling user's ID; successful
mutations flush the caller's per-user entries (read-your-own-writes); user,
enrollment, and project invalidation clear the relevant response entries; the
narrow enrollment invalidation now broadcasts a `userenroll` sync message to
other replicas; `/metrics/http` answers only on loopback. Unit tests cover the
key shapes, shareability classification and all three invalidation paths; the
branch's full suite including e2e passes.

The harness probes were re-pointed for the second pass (fresh enroll user,
new P4 message) since the databases carry the first run's writes.
