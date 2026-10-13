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

## Betting load test on the LT pair (2026-10-13)

Two Hetzner CCX33 boxes in Helsinki (8 dedicated vCPU, 32 GB each), close to
prod's Ryzen 5 3600 (6C/12T):

- **interact-lt-server**: a prod copy built from the prod Ansible roles via
  `ansible/loadtest.yml`. It runs
  hardening, tuning, firewall, postgres, interact and wayfarer, and skips
  tailscale, backup and jobs. `group_vars/interact_lt.yml` turns off the
  postgres role's tailnet assert and keeps Postgres on loopback.
  - The app runs natively as `wayfarer@{blue,green}` (SO_REUSEPORT,
    `DB_MAX_OPEN_CONNS=230`) and terminates TLS itself, like prod. The
    certificate is a free Let's Encrypt cert for **`<server-ip-with-dashes>.sslip.io`**,
    obtained by autocert (ECDSA); the sslip.io name resolves to the IP, so no
    DNS record is needed.
  - Firestore, S3, Pub/Sub, push and mail are not configured.
- **interact-lt-client**: k6 only, with the same
  `tuning` role (nofile, somaxconn, ports) so the generator isn't the limit.

Everything is driven by `backend/scripts/loadtest-lt.sh`. Secrets (DB
password, JWT secret, Auth0 load-test key) are generated once into
`backend/cmd/loadtest/results/lt-secrets.env`, which is gitignored.

The boxes are created for a round of tests and deleted afterwards, so their
IPs are not in the repo. `loadtest-lt.sh hosts` records them in
`results/lt-hosts.env` (gitignored); every other command reads that file, or
`LT_SERVER_IP`/`LT_CLIENT_IPS` from the environment. The Ansible inventory is
generated from it (`results/lt-inventory.yml`), and the TLS name is the
server's IP with dashes plus `.sslip.io`. A recreated box that reuses an IP
has a new SSH host key: `ssh-keygen -R <ip>`.

One-time setup per set of boxes:

```
cd backend
./scripts/loadtest-lt.sh hosts <server-ip> <client-ip> [<client-ip>...]
./scripts/loadtest-lt.sh provision   # ansible, both boxes
./scripts/loadtest-lt.sh env         # /opt/wayfarer/wayfarer.env on the server
./scripts/loadtest-lt.sh deploy      # build like CI, ship, deploy.sh, wait for TLS
./scripts/loadtest-lt.sh seed        # gendata*.sql (~13k users with points) + betting quiz
make loadtest-betting-config         # 13k user + 13k Auth0 tokens + admin token
```

Runs:

- `make loadtest-betting-smoke` (~200 users)
- `make loadtest-betting-rampspike` (full 10k)

Every run does the following:

- Resets the betting quiz with `bench/betting_box.sh reset`. This removes the
  previous run's submissions and BET/QUIZ score_journal rows, so user points
  are the same each time.
- Restarts the active color.
- Runs `k6/betting-quiz-rampspike.js` from the client with `REALISM=1`
  (Firebase fetches off).
- Samples CPU on both boxes, and the settled-bet count once per second
  (`<label>.sql.tsv`).
- Ends with a per-bet settlement check: every settled bet has exactly two BET
  journal rows (stake and winnings) whose sum equals the response's
  `points_earned`.

The k6 script reuses the free-text rampspike's front-loaded curve: cumulative
1k@1s, 5k@2s, 8k@4s and 10k@10s. Each user answers 5 PREDEFINED betting
questions, one per limit style (max 25%, max 50%, max 100 absolute, min 5 +
max 50% at ×3, unlimited at ×1.5/×0.5). Per question:

| Share | Behaviour | Expected outcome |
|---|---|---|
| 90% | A bet the app's slider allows: within the limits and the user's points minus stakes already placed (read once via `UserPoints`) | Accepted |
| 5% | No bet | Accepted |
| 3% | Double tap: the same bet sent twice in parallel | Both accepted |
| 2% | A bet above the available points, then a retry with a sensible bet | First must be rejected for points |

- **Correct answers:** a user answers correctly with probability
  `CORRECT_RATE` (0.5), so both win and loss payouts are exercised.
- **`bets_wrong_outcome` threshold is 0:** a sensible bet refused, or an
  over-limit bet accepted, fails the run.
- **Settlement is measured inside the run:** `teardown()` locks and finishes
  the session with the admin token (30m request timeout,
  `settlement_duration` metric), like the admin pressing finish after the
  quiz. Prod settles bets inside the `finishQuizSession` request, so this
  also shows whether settlement outlives `SERVER_WRITE_TIMEOUT` (5m) at camp
  scale.

The knobs are the same as the free-text rampspike (`RAMP_SCALE`,
`THINK_SCALE`, `THINK_UNIFORM`, `PRE_ENROLLED`, `REALISM`), plus:

- `CORRECT_RATE`
- `NO_BET_FRACTION`, `DOUBLE_TAP_FRACTION`, `OVER_BET_FRACTION`
- `SETTLE=0` to skip settlement
- `FINISH_TIMEOUT`

### Two generators

One CCX33 running k6 tops out at about 800% CPU during the 10k stampede.
The first full run delivered only 7,400 of the 10,000 arrivals (3,620
dropped iterations). There are now two client boxes:

- `interact-lt-client`
- `interact-lt-client2`

How a run is split across them:

- Each generator runs one shard (`SHARDS`/`SHARD`): half of every phase's
  arrival rate, and every second token of each phase's slice.
- After both shards finish, `k6/betting-settle.js` runs once on the first
  generator (the shards run with `SETTLE=0`). Settlement therefore starts
  only when every user is done.
- The generator list is every client IP given to `loadtest-lt.sh hosts`.

### Run data (for graphing afterwards)

Every run writes `backend/cmd/loadtest/results/<label>-<UTC timestamp>/`
(gitignored):

| File | Content |
|---|---|
| `manifest.json` | git commit, k6 args, hosts (nproc/mem/kernel), non-secret server env, Postgres settings, start/end UTC |
| `generator<i>/shard<i>.csv.gz` | every k6 sample with tags (`name`, `scenario`, status, …): the exact per-request data |
| `generator<i>/shard<i>.summary.json`, `.k6.log` | k6 end-of-test summary and console |
| `generator<i>/shard<i>.k6.pidstat`, `.box.mpstat` | k6 process and generator box CPU, 1s |
| `generator0/settle.*` | the settlement request (`FinishQuizSession` duration) |
| `server/server.pidstat` | server process CPU + RSS, 1s |
| `server/postgres.pidstat` | all postgres processes, 1s |
| `server/server-box.mpstat` | whole server box, 1s |
| `server/sql-sample.tsv` | settled/total bets, 1s |
| `server/bets.csv` | every response: bet, correctness, points, `answered_at`, `settled_at` |
| `server/submissions.csv` | every submission: start/complete times, score, points |
| `server/bet_journal.csv` | every BET score_journal row (stake/winnings, `created_at`) |
| `server/quiz_journal.csv` | completion points rows |
| `server/pg_stat_statements.csv` | per-query stats for exactly this run (reset at start) |
| `server/httpstats.jsonl` | the server's own HTTP stats (15s) |

`go run ./cmd/loadtest/analyzerun <run-dir>` (from `backend/`, run
automatically after every run) merges the shards. It adds graph-ready CSVs: `requests_per_second.csv` (per
second × request name: count, failed, p50/p95/p99/max), `tls_per_second.csv`,
`cpu_per_second.csv` (server/postgres/boxes/k6 side by side) and
`settlement_per_second.csv` (from the exact journal timestamps). It also
writes a `summary.json`.

Gotchas, all found while bringing the pair up:

- **SSH key:** the scripts use a dedicated key,
  `backend/cmd/loadtest/results/lt_ed25519` (gitignored, installed on both
  boxes). A developer's agent may refuse to sign for background tools (it
  prompts or confirms), and the hardening role turns password login off.
- **SSH is rate-limited:** the prod firewall role limits SSH to 6 connections
  per 30s per IP. The runners multiplex SSH (`ControlMaster`); ad-hoc loops of
  `ssh` calls get refused. Inside `while read` loops give ssh `</dev/null`,
  or it consumes the loop's input.
- **The CPU governor unit failed on VMs** (no cpufreq). The tuning role's
  loop now skips missing files instead of failing.
- **The server won't start on an empty DB:** migrations seed a
  `current_project_id` that doesn't exist, so SettingsService exits. Run
  `seed` (which repoints it) before the first successful deploy; the first
  boot still applies the migrations.
- **gendata identities aren't prod-shaped:** `members_id` is `M<n>`, and there
  is no `person_uuid` and no church `external_id`. `betting_box.sh seed`
  fills them in like prod (numeric person id, UUID, numeric church id),
  otherwise `tokengen -auth0-count` mints nothing.
- **The bet slider's input is `myCurrentProject.myPoints`** on the challenge
  page. `User.points(projectId:)` is a plain model field that is 0 here; don't
  use it.
- **Auth0 dances 401 for ~30s after a restart:** the server fetches its own
  `/jwks.json`, which fails at boot and is retried every 30s. LT runs wait
  40s (`LOADTEST_WARMUP`). Prod is not affected.
- **Reset deadlocked with the server:** the leaderboard drain worker raced
  the reset's synchronous delete trigger. The runner now stops the server
  around `betting_box.sh reset`, and reset drains the leaderboard queue first.
- **Postgres version:** the boxes run PG 17 from Debian 13, same as prod. CI
  e2e uses 16.
