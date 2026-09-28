# Off-box load testing (k6 on a dev machine, server on the bench box)

**Date:** 2026-08-21
**Follow-up to:** `ram-first-architecture.html` §18/§19 — on-box runs had k6
peaking at 705% CPU (1.7x the server itself), so every §17 latency figure is
pessimistic. "Re-measure with an off-box load generator" was the top-priority
outstanding work. This is that setup.

## The setup

- **Bench box** (`root@49.12.121.62`, Ryzen 5 3600, Debian 13): runs
  `wayfarer.service` (env in `/opt/wayfarer/wayfarer.env`, binary
  `/opt/wayfarer/bin/server`, listening on `*:8080`) and PostgreSQL 17 on
  loopback (`bench:bench@127.0.0.1:5432/wayfarer_bench`). The loadtest tree
  incl. tokengen is at `/opt/wayfarer/backend/cmd/loadtest`. Box provisioning
  artifacts are committed in `backend/cmd/loadtest/bench/`.
- **This machine**: runs only k6, over the public internet against port 8080.

## How to run

From `backend/`:

```
make loadtest-remote-config     # ssh to box, mint 13k tokens with baseUrl
                                # pointing at the box, scp config.json here
make loadtest-remote-smoke      # RAMP_SCALE=0.02 (~200 users) plumbing check
make loadtest-remote-rampspike  # the full 10k front-loaded QR stampede
```

All three call `scripts/loadtest-remote.sh`, which also has `prep` (reset
quiz submissions + restart server + wait for the token-warmer boot pass) and
a generic `run <label> [k6 args]`. Knobs: `LOADTEST_SSH`,
`LOADTEST_BASE_URL`, `LOADTEST_REMOTE_DIR`, `LOADTEST_TOKENS`,
`LOADTEST_WARMUP`.

A `run` does: prep → record `score_journal` count → start `pidstat`/`mpstat`
samplers on the box → run k6 **locally** with `--env BASE_URL=…` (a new
override in `freetext-quiz-rampspike.js` / `freetext-quiz-spike.js`, so a
config.json minted with a loopback baseUrl still works) → stop samplers →
scp the `.cpu` files back → print the same summary `bench/runramp.sh`
printed (thresholds, completions, score_journal delta, server/postgres peak
CPU, box busy). Local outputs land in `backend/cmd/loadtest/results/`
(gitignored); box copies stay in `/opt/wayfarer/results/`.

## Caveats

- **Numbers are not comparable to §17.** The Mac↔box RTT (public internet)
  now enters every `http_req_duration`. Compare off-box runs only to other
  off-box runs from the same network; the box-side CPU figures and the
  latency *deltas* between configurations are the meaningful signal.
- Tokens expire (`tokengen -valid-days 7`); re-run `loadtest-remote-config`
  if runs suddenly fail auth checks.
- `prep`/`run` restart `wayfarer.service` and delete the loadtest quiz's
  submissions on the box. The box is a dedicated playground; don't point
  `LOADTEST_SSH` anywhere else.
- The generator side is now a laptop: keep an eye on local CPU/Wi-Fi. If the
  full 10k spike saturates the local machine or link, that shows up as
  k6-side `dropped_iterations` — trust the box-side samplers, and rerun at
  lower `RAMP_SCALE` from a wired connection.

## Realism mode (2026-08-27)

`REALISM=1` (k6 env, toggleable per run via the loadtestui "realism" checkbox
or `loadtest-remote.sh run <label> --env REALISM=1`) turns on realistic auth
behavior in `userJourney`:

- **3% of journeys do the Auth0 dance** (`AUTH_DANCE_FRACTION`, override
  freely): `GET /token?token=<simulated-auth0-jwt>` → JWKS validation → user
  lookup → Wayfarer JWT minting; the journey continues with the returned
  token. Tagged `AuthCallback` in k6/HTTP stats.
- **50% of journeys re-fetch the Firebase token mid-session**
  (`FIREBASE_REFRESH_FRACTION`), like clients whose 1h custom token expired.
  Server-side these are mostly cache hits (30 min TTL + warmer) — that is the
  production profile.

Setup for the Auth0 dance:

1. `tokengen -auth0-count N -jwks-out jwks.json` — mints RS256 tokens shaped
   like login.bcc.no tokens into `config.json` (`auth0Tokens`) and writes the
   matching JWKS. Needs users with numeric `members_id`, `person_uuid`, and a
   church with `external_id`.
2. Give the server the same key: tokengen signs with the RSA key from
   `AUTH0_LOADTEST_PRIVATE_KEY` (or `-auth0-key`; a fresh key is generated and
   printed if neither is set). Set that same `AUTH0_LOADTEST_PRIVATE_KEY` in
   the server env — the server derives the JWKS from it and serves it at
   `GET /jwks.json` itself, so `AUTH0_JWKS_URL=http://127.0.0.1:8080/jwks.json`
   just points back at the server (boot-time fetch of the self-URL failing is
   tolerated in this mode). No files, no sidecar. Set
   `AUTH0_JWT_ISSUER=https://login.bcc.no/`. Members API stays unconfigured —
   church resolution falls back to the token's churchId claim.

Real Firestore: `backend/loadtest.env` is the load-test host env (gitignored —
it embeds the real bccm-pc25 service account as base64 directly in
`FIREBASE_SERVICE_ACCOUNT`; config.go accepts path / raw JSON / base64;
re-generate with `base64 -i <sa.json> | tr -d '\n'`). This makes the
notification writes (challenge completions,
achievement awards, quiz finalize, ...) hit real Firestore — real gRPC pool,
latency, and quota behavior, unlike the fake SA whose writes fail at the OAuth
step and get swallowed. Caveats: writes land in that project's Firestore
(`users/{id}/notifications/*`, `projects/{id}/notifications/*`), count against
quota/billing, and a high-rate run fans out one write per award/completion.

## Limit-finding campaign on interact-test (2026-08-27)

Setup: interact-test.bcc.media (12 cores / 62GB, dokploy+traefik, host
Postgres, TLS on-box), 13,162 seeded users, generator in same DC. Full 10k
rampspike, REALISM=1.

Limits found, in order (each fixed before the next appeared):
1. **Postgres connection slots** — pool (100) > max_connections (100−3).
   SQLSTATE 53300 at burst peak. Fixed: max_connections=400 (ansible),
   pool 250/250 warm (loadtest.env).
2. **App container accept backlog** — net.core.somaxconn is per-netns; host
   tuning doesn't reach containers. Fixed: sysctls in dokploy compose (8192).
3. **Traefik TLS handshake cost** — ACME default RSA-4096 ≈ 5-10ms CPU per
   handshake; 1k handshakes/s ate ~11 of 12 cores (502s, app starved).
   Fixed: keyType: EC256 in /etc/dokploy/traefik/traefik.yml (+ acme.json
   reset). Read p95 dropped ~4x.
4. **Traefik proxy throughput (structural, unfixed)** — ~5k req/s across
   ~11k concurrent TLS conns costs traefik ~11 cores regardless of key type
   (measured with THINK_SCALE=0.03, a ~30x-compressed stressor). No config
   fat left (no compression/accesslog middleware). Escalations if ever
   needed: CDN in front, proxy swap, or more cores.

Result on the tuned stack, realistic full 10k profile: 100.00% of 396k
checks, 0 failed requests, 9,780 completions, p95 1.75s during the 10s
stampede (FinalizeQuiz p95 15.75ms). Verdict: prod-on-this-box with on-box
TLS handles the realistic event with headroom; the 0.03 stressor ceiling is
the documented boundary.

Operational notes: quiz fixtures reset via /tmp/quiz_fixtures.sql on the box
(psql -d wayfarer_bench); traefik config backup at traefik.yml.bak, old RSA
cert store at dynamic/acme.json.rsa.bak; watch for stale duplicate app
containers after dokploy redeploys (traefik "multiple Services" errors).
