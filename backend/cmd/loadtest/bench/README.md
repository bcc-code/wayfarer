# Bench-box artifacts

These files reproduce the bare-metal benchmark environment measured in
`notes/ram-first-architecture.html` §17 (AMD Ryzen 5 3600, Debian 13,
PostgreSQL 17 on localhost, `wayfarer.service` systemd unit in
`/opt/wayfarer`). They originally lived only on the box; they are committed
here because the report's reproducibility claim depends on them.

- `setup_fixtures.sh` — one-time box setup: synthetic Firebase service
  account (local RSA signing only, no real credentials), loadtest quiz
  fixtures, Postgres settings. Run **on the box** as root.
- `gendata.sql`, `gendata2.sql` — generate the production-scale dataset
  (~13k users, 790 teams, ~484k score_journal rows). Run on the box via
  `psql` after migrations.
- `runramp.sh`, `runspike.sh` — the original **on-box** run harnesses
  (k6 co-located with the server). Kept for reference; superseded for
  measurement by the off-box runner `backend/scripts/loadtest-remote.sh`
  (`make loadtest-remote-smoke` / `loadtest-remote-rampspike`), because
  co-located k6 peaked at 1.7x the server's CPU and skewed every number.
