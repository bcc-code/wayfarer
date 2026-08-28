# Wayfarer backend — optimization history, step by step

How the backend went from a distributed cloud deployment that buckled under a
10,000-user spike to a single box that serves the same spike with 100% success
and TLS terminated locally. Each step lists the bottleneck, the change, and
the measured effect.

Sources: `notes/ram-first-architecture.html` (§17–§19 measurements, 2026-08-14,
in the `worktree-ram-first` worktree), `notes/loadtest-offbox.md`,
`notes/loadtest-performance-round{2,3}.md`, and the limit-finding campaign of
2026-08-27 (documented at the end of `loadtest-offbox.md`).

---

## Phase 0 — the starting point, and its taxes (distributed era)

Several stateless Cloud Run instances against remote Neon Postgres. Most of
the pain was not application logic but the price of being distributed:

| Tax | Mechanism | Cost |
|---|---|---|
| Remote DB connections | Each new pool connection did TLS + pooler auth to Neon | **5–9 s connect spans** at spike onset; `GetMe` averaged 8.5 s while queries queued behind `pool.acquire` |
| Connection ceiling | Neon limits forced `DB_MAX_OPEN_CONNS=25` | Pool starvation under any burst |
| Statement latency | Network round-trip per statement | 30–90 ms baseline vs sub-ms on loopback |
| Cross-instance cache invalidation | `internal/cache/sync.go`: a dedicated non-pooled `LISTEN cache_invalidate` connection per instance, `pg_notify` fan-out, full `cache.Clear()` on any reconnect | ~300 lines of failure-prone machinery; a connection blip cold-flushed the whole cache |
| Per-instance caches | Ristretto with guessed TTLs (45 s…1 h), hashed filter keys that can't be selectively invalidated | Stampedes and stale windows; instances could disagree |

## Phase 1 — the RAM-first decision: scale up, not out (2026-08-14)

Premises: **one process, one bigger machine; Postgres on loopback.** The
original proposal went further — a fully resident in-memory "World" read
model — but measurement on the real hardware (Ryzen 5 3600, 62 GB, NVMe,
PG 17 local) settled it:

- The whole production dataset is **294 MB** — Postgres caches all of it in RAM.
- Local Postgres alone dissolved the worst latency source (the 5–9 s connect
  spans) and retired the 25-connection rule; the pool is now sized to the box.
- The existing architecture passed 10k users / 10 s **flat** arrival at
  p95 303 ms / p99 379 ms, 0 failures. The resident World was shelved
  (§18: "don't build it for performance") — kept as a design for a future
  capacity problem, not a latency one.

**The move to one box was itself the biggest optimization.** Everything after
is targeted fixes found by profiling the front-loaded arrival curve, which
degraded p95 by 7× vs flat arrival (identical volume — arrival *shape* is
what breaks things).

## Phase 2 — the measured burst fixes (2026-08-14, `worktree-ram-first`, merged)

Profiling the front-loaded 10k burst put ~25% of server CPU in RSA signing,
~20% in GC, and none in scoring. Cumulative effect on the ramped-arrival run:

| Step | p95 | p99 | p50 |
|---|---|---|---|
| Baseline (local PG, ramped 10k) | 2,157 ms | 2,812 ms | 242 ms |
| + Firebase **token warmer**, remove DEBUG print, pool 100 | 1,430 ms | 1,870 ms | 14 ms |
| + narrowed **enrollment invalidation** | 947 ms | 1,210 ms | 13 ms |
| + **GOGC=600, GOMEMLIMIT=12GiB** | **734 ms** | **897 ms** | 14 ms |

1. **Firebase token warmer** (`internal/services/firebase_token_warmer.go`).
   Custom-token minting is a local RSA-2048 signature (~1 ms = 1 core per
   1,000 users/s). Cold burst minting ate up to 4 cores. Background rotation
   (40 min interval, 55 min TTL) costs ~5 signs/s; boot pass mints 11k in 3 s.
   `GetFirebaseToken` p95: 701 → 118 ms.
2. **Enrollment invalidation blast radius.** Self-enrollment invalidated five
   *global* per-user cache prefixes — at 4,000 enrolls/s the cache never
   survived to serve the ChallengePage that follows. Replaced with five exact
   per-user deletes.
3. **A leftover `fmt.Printf("DEBUG: …")`** on every enrollment, writing to
   stdout at burst rate.
4. **GC tuning** — after RSA, allocation was the top cost; the box was using
   1 GB of 62. `GOGC=600` + `GOMEMLIMIT=12GiB` traded idle RAM for GC time.
5. Housekeeping that had poisoned earlier measurements: CPU governor was
   `powersave` (clocks capped at 71%), OTel left on, RAID resync running.

Also from this pass, correctness findings kept as guardrails: the scoring
write (`FinalizeQuiz` — journal insert + 38-statement trigger fan-out) was
**never** the bottleneck (14 ms p95, unmoved under load); zero-point test
quizzes had made three earlier perf rounds blind to the scoring path; and the
rank query rewrite (`DENSE_RANK()` over the full board → `COUNT(DISTINCT
score)+1`: 362 → 4.5 ms) remains the known fix if standings queries ever
matter in a spike.

## Phase 3 — honest measurement: off-box load generation (2026-08-21)

k6 on the same box peaked at 705% CPU vs the server's 422% — the generator
was starving the system under test. `scripts/loadtest-remote.sh` +
`cmd/loadtestui` moved generation to a separate machine (same-DC VM, ~1 ms
RTT), with windowed JSONL server stats (`HTTP_STATS_FILE`) and per-route
percentiles replacing guesswork. Related server work: whole-response cache
(30 s TTL) for user-independent queries, response-cache keying fixes,
aggregated HTTP stats, gin release mode in production (per-request access
logging is measurable at spike rates).

## Phase 4 — production topology on one box: the limit ladder (2026-08-27)

The dokploy/traefik deployment (interact-test.bcc.media, 12 cores / 62 GB,
TLS on-box) was pushed with the full 10k rampspike (REALISM=1: Auth0 dance,
Firebase token refresh, real Firestore writes) until something broke —
four times. Each limit fixed before the next appeared:

| # | Limit | Symptom | Fix |
|---|---|---|---|
| 1 | Postgres `max_connections=100` vs app pool 100 | SQLSTATE 53300 at burst peak | `max_connections=400` (ansible), pool 250/250 kept warm |
| 2 | Container accept backlog (somaxconn is per-netns; host tuning doesn't reach it) | 502s under >4k conn/s burst | `sysctls: net.core.somaxconn=8192` in the compose |
| 3 | **RSA-4096 TLS handshakes** (traefik ACME default) | traefik at ~11 of 12 cores, app starved, 502s | `keyType: EC256` — handshakes ~25× cheaper; read p95 dropped 4× |
| 4 | Traefik proxy throughput (structural) | ~5k req/s over ~11k TLS conns ≈ 11 cores, key type irrelevant | Unfixed; only reachable by the 30×-compressed stressor. Escalations if ever needed: CDN in front, proxy swap, or Go-terminated TLS (costs ACME plumbing + zero-downtime deploys) |

Note the rhyme between phases 2 and 4: RSA signature cost was the dominant
burst CPU both times — first Firebase custom tokens (fixed by the warmer,
ES256 unavailable there), then TLS handshakes (fixed by EC256, available).

## Where it landed (2026-08-27, tuned stack, realistic 10k event, TLS on-box)

- **100.00% of 396,346 checks, 0 failed requests, 9,780 quiz completions**
- Client-side p95 1.75 s *during the 10-second stampede window*; server-side
  p99 ≈ 1 s in that window and **10–25 ms for the rest of the run**
- `FinalizeQuiz` (scoring + Firestore fan-out) p95 15.75 ms throughout
- App peak ~3.5 cores; the box has headroom on the realistic profile

## The distilled lessons

1. **Locality beats machinery.** Local Postgres removed more latency than any
   code change; most of the "distributed" code (cross-instance invalidation,
   connection ceilings, per-instance cache skew) existed to compensate for
   distance the system didn't need.
2. **RAM is the budget, not the architecture.** The dataset fits in cache
   many times over; the resident in-memory World was measured (builds in
   178 ms) and *rejected* because Postgres-in-RAM already delivers.
3. **Arrival shape > arrival volume.** Same 10k users, 7× p95 difference
   between flat and front-loaded curves. Test the stampede.
4. **Asymmetric crypto is the recurring burst tax.** Profile for it first.
5. **Never co-locate the generator; check the governor; make tests assert the
   work happened** (a zero-point quiz hid the scoring path for three rounds).
