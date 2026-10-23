# Log Shipping (VictoriaLogs)

Logs from the app, Postgres, cron jobs and the host go to VictoriaLogs at
`http://10.12.140.76:9428`. Prod reaches it over a Tailscale subnet route
(`roles/tailscale` runs `tailscale set --accept-routes`).

## Design rule: logging is the first thing to fail

Log export must never add latency or compete with the app under load. Each
layer sheds load instead of applying back-pressure:

| Layer | Buffer | When it can't keep up |
| ----- | ------ | --------------------- |
| App `slog` handler (`internal/logger/export.go`) | in-memory channel, `LOG_EXPORT_BUFFER_SIZE` (10k records) | non-blocking send; record dropped, `dropped` counter++ |
| App flusher goroutine | batches of `LOG_EXPORT_BATCH_SIZE` / 1 MiB / `LOG_EXPORT_FLUSH_INTERVAL` | 2s HTTP timeout; on failure, exponential backoff (1s → 30s) during which batches are dropped, not retried |
| Vector agent (`roles/vector`) | disk buffer, 1 GiB | `when_full: drop_newest`; unit runs at `Nice=10`, `CPUQuota=50%`, `MemoryMax=512M` |

The app keeps writing to stdout/journald as before, so `journalctl -u wayfarer@blue`
still works when shipping is down. A periodic local warning
(`log export dropped entries`) and `GET /metrics/logexport` (loopback only:
`queued/sent/dropped/failed`) show when shedding happens.

## Pipeline

```
wayfarer@blue/green --gzip NDJSON--> Vector http_server 127.0.0.1:9880 --+
journald (all units except wayfarer@*) --> Vector journald -------------+
/var/log/postgresql/*.json (jsonlog) --> Vector file -------------------+--> VictoriaLogs /insert/jsonline
syslog 127.0.0.1:5514 tcp/udp --> Vector syslog ------------------------+
```

The server colors are excluded from Vector's journald source because they export
directly, with structured fields; otherwise every app line would arrive twice.

Every event is normalised to `_msg`, `_time`, `level` (`debug|info|warn|error`),
`service`, `host`. Stream fields (`_stream_fields`) are `host,service,instance`.
Keep them low-cardinality.

| Source | `service` | Extra fields |
| ------ | --------- | ------------ |
| App | `OTEL_SERVICE_NAME` (`wayfarer-backend`) | `instance` (blue/green), `env`, `trace_id`, `span_id`, all slog attrs |
| journald | unit name without `.service` (`wayfarer-job-sync-members`, `sshd`, `postgresql@17-main`), else `SYSLOG_IDENTIFIER` | `unit`, `identifier`, `pid` |
| Postgres | `postgres` | jsonlog fields: `user`, `dbname`, `pid`, `error_severity`, `state_code`, `statement`, `detail`, `backend_type`, … |
| syslog | appname | `facility`, `procid` |

## What the app exports

- Every record at `LOG_EXPORT_LEVEL` (default `info`) or above. The global
  `LOG_LEVEL` and settings-table override still gate it.
- **Export-only extras.** These never reach stdout, so journald volume is
  unchanged:
  - GraphQL responses with errors (`api.ErrorLogger`): one warn line per response with the operation name and type, plus each error's message, path and code. Variables are never logged.
  - Sampled access log (`LOG_ACCESS_SAMPLE_RATIO`, default 0 = off): method, route, status, `duration_ms`, bytes, client and user agent.
- `DB_LOG_QUERIES=true` query logs flow through the same exporter.

CLI tools (`cmd/seed`, `cmd/migrate`, …) use plain `logger.New` and don't export.

## Postgres logging

`roles/postgres/templates/90-tuning.conf.j2` switches Postgres to
`logging_collector=on`, `log_destination='jsonlog'`, one file per weekday
(`postgresql-Mon.json`, truncated on reuse, so 7 days are kept locally), and
`log_timezone=UTC`. It logs statements over 500 ms, lock waits, checkpoints,
autovacuum runs over 1 s, and temp files over 10 MB. **Changing
`logging_collector` needs a Postgres restart.** The role's handler restarts
Postgres on the next `--tags postgres` run.

## Configuration

App (`backend/.env.example`; prod values come from the per-color env files in
`roles/wayfarer`):

| Var | Default | |
| --- | ------- | - |
| `LOG_EXPORT_ENABLED` | `false` | |
| `LOG_EXPORT_URL` | — | prod: `http://127.0.0.1:9880/` (Vector); dev: VictoriaLogs `/insert/jsonline?...` directly |
| `LOG_EXPORT_LEVEL` | `info` | |
| `LOG_EXPORT_BUFFER_SIZE` | `10000` | |
| `LOG_EXPORT_BATCH_SIZE` | `1000` | |
| `LOG_EXPORT_FLUSH_INTERVAL` | `2s` | |
| `LOG_EXPORT_INSTANCE` | — | blue/green |
| `LOG_ACCESS_SAMPLE_RATIO` | `0` | 0–1 |

Ansible (`group_vars/all.yml`): `vlogs_url`, `vector_version`,
`vector_buffer_max_bytes`, `vector_cpu_quota`, `vector_memory_max`,
`wayfarer_log_export_url`. Load-test boxes set `wayfarer_log_export_url: ""`
(no Vector there).

## Useful LogsQL

```
service:wayfarer-backend level:error
service:wayfarer-backend trace_id:0af7651916cd43dd8448eb211c80319c
service:wayfarer-backend _msg:"graphql response has errors" | stats by (operation) count()
service:postgres error_severity:ERROR
service:postgres _msg:"duration:" | sort by (_time) desc | limit 20
service:~"wayfarer-job-.*"
```

Query from a tailnet machine:
`curl http://10.12.140.76:9428/select/logsql/query -d 'query=service:postgres level:error' -d 'limit=20'`.

## Syslog

Vector's journald source already covers OS syslog, because journald owns
`/dev/log`. The loopback syslog source on port 5514 (TCP/UDP, RFC 5424/3164)
is a standard endpoint for anything else on the host. To receive syslog from
other machines directly, start VictoriaLogs with `-syslog.listenAddr.tcp=:514`.
That is independent of this pipeline.
