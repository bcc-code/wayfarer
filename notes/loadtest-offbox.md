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
