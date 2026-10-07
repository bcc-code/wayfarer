# Configuration Management

**Date**: 2025-10-31

## Design Pattern

**Rule**: All environment variables are loaded ONCE at startup in `main()`. No direct env access anywhere else in the codebase.

## Config Structure

The configuration is organized into logical sections:

```go
type Config struct {
    Server   ServerConfig   // HTTP server settings
    Database DatabaseConfig // Database connection
    JWT      JWTConfig      // JWT authentication
    Log      LogConfig      // Logging configuration
}
```

### Server Configuration
- `SERVER_HOST` - Bind address (default: 0.0.0.0)
- `SERVER_PORT` - HTTP port (default: 8080)
- `SERVER_READ_TIMEOUT` - Request read timeout (default: 10s)
- `SERVER_WRITE_TIMEOUT` - Response write timeout (default: 10s)
- `SERVER_IDLE_TIMEOUT` - Keep-alive timeout (default: 120s)

### Database Configuration
- `DATABASE_URL` - PostgreSQL connection string (required)
- `DB_MAX_OPEN_CONNS` - Max open connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Max idle connections (default: 5)
- `DB_CONN_MAX_LIFETIME` - Connection max lifetime (default: 5m)
- `DB_CONN_MAX_IDLE_TIME` - Connection max idle time (default: 5m)

### JWT Configuration
- `JWT_SECRET` - HMAC secret for signing and verifying Wayfarer access tokens
- `JWT_SECRET_PREVIOUS` - Previous secret, still accepted during rotation (selected by the token's `kid`)
- `JWT_ISSUER` - JWT issuer claim (default: "wayfarer")
- `JWT_ACCESS_TOKEN_TTL` - Access token lifetime (default: 168h)
- `JWT_REFRESH_TOKEN_TTL` - Sliding refresh token lifetime (default: 4392h, about 6 months)
- `AUTH0_AUDIENCE` - Expected `aud` of login.bcc.no tokens (unchecked when empty)
- `AUTH_REVOKED_USERS` - Emergency revocation list, normally empty. See `notes/14-auth-sessions.md`

### Log Configuration
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: info)
- `LOG_FORMAT` - Log format: json, text (default: json)

## Usage Pattern

```go
// In cmd/server/main.go
func main() {
    // Load ALL config at startup
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // Pass config subsections to components
    db := database.Connect(cfg.Database)
    server := server.New(cfg.Server, db)
    server.AddMiddleware(middleware.Auth(cfg.JWT))
    server.AddMiddleware(middleware.Logger(cfg.Log))
}
```

## Configuration Distribution

Components receive ONLY the configuration they need:
- Database package receives `DatabaseConfig`
- Server receives `ServerConfig`
- Auth middleware receives `JWTConfig`
- Logger middleware receives `LogConfig`

This ensures:
1. Clear dependencies
2. Easy testing (pass custom config structs)
3. No hidden environment variable access
4. Type safety

## Validation

The `Load()` function validates required fields:
- `DATABASE_URL` must be set (returns error otherwise)
- Other fields use sensible defaults

## Testing

Config package includes full test coverage:
- Default value loading
- Environment variable parsing
- Integer parsing
- Duration parsing
- Validation errors

Tests use environment variable isolation to avoid interference.

## Runtime settings (the `settings` table)

Separate from the env-based config above, a `settings` key-value table holds
configuration that can change without a redeploy. It is owned by
`backend/internal/services/settings.go`.

**The table wins over the environment.** `config.LoadSettings` reads it over a
connection of its own right after `config.Load()`, and `config.ApplySettings`
overlays the rows that are set onto the loaded config. A row that is absent,
NULL or empty leaves the environment's value in force. The read is best-effort:
an unreachable or not-yet-migrated database is not an error.

That read has to happen that early, and over its own connection, because the
logger, the tracer and the connection pool are all configured from values it
can override — `db_log_queries` is an input to building the very pool every
other query runs on.

`config.SettingSpecs` is the registry. Adding a key is one entry there plus the
row; it carries the environment variable the key overrides and whether a change
needs a restart.

| Key | Overrides | Takes effect |
| --- | --- | --- |
| `log_level` | `LOG_LEVEL` | immediately |
| `db_log_queries` | `DB_LOG_QUERIES` | on restart |
| `otel_enabled` | `OTEL_ENABLED` | on restart |
| `otel_sampling_ratio` | `OTEL_SAMPLING_RATIO` | on restart |
| `ssf_debug_mode` | `SSF_DEBUG_MODE` | on restart |

`log_level` is the exception because `internal/logger` resolves its threshold
through a shared `slog.LevelVar`, which the handler clones in
`WithAttrs`/`WithGroup` read too — so `SetLevel` reaches loggers that already
exist, and `SettingsService.applyLiveSettings` calls it on every refresh.

`frontend_config` is a free-form JSON blob served by `Query.frontendConfig`,
which reads it from the database per request — so edits are live. Its keys are
ad hoc and project-specific (`team_name_changed_challenge_id`,
`gamenight_1_betting_quiz_id`), consumed by the my-church kickoff and gamenight
pages as entity ids. **No migration creates this row**; where it exists it was
inserted by hand, and `SetSetting` cannot create it.

`current_project_id` is not configuration at all: it is application data naming
the project every end user sees, read by `Query.currentProject` /
`myCurrentProject` / `myProjects`, the church-admin statistics resolver, the
Firebase token warmer and the two ladder-to-heaven URL handlers. It is
deliberately absent from `SettingSpecs`.

### How a value is read

The whole table is loaded into a process-local `atomic.Value` map at boot and
refreshed every five minutes. Reads are in-memory map lookups, never a query.

The initial load is fail-fast: an invalid `current_project_id` aborts startup.
A later refresh is not — it logs and keeps the previous map. That asymmetry is
deliberate: this used to `panic()` on every path, so a bad value written
straight to the database took the process down from the background ticker,
arbitrarily far from whatever caused it.

### How a value is changed

`setCurrentProject(projectId:)` and `setSetting(key:value:)`, both
`@requireRole(roles: ["superadmin"])`. The `settings` query authorises
in-resolver instead, per the project convention.

### What the API can and cannot write

Three gates, because this is remote control over server configuration:

1. **superadmin only** — both mutations carry the directive.
2. **No key can be created.** `SetSettings` requires each key to already be a row,
   and the only queries it runs are `UpdateSetting*`, which are plain `UPDATE
   ... WHERE key = @key`. The `SetSetting*` upserts in `settings.sql` would
   insert, and are deliberately left uncalled. There is no delete path either.
3. **Only known keys are writable.** `services.IsEditableSetting` admits
   `config.SettingSpecs` plus `appDataSettings` — membership in the table is
   not enough. So a row added to the database does not become remotely writable
   by virtue of existing, and making a future sensitive key writable takes a
   deliberate registry entry. The `settings` query still lists unknown rows,
   marked `editable: false` and rendered read-only.

`appDataSettings` is the second group: keys the application reads directly
rather than through config, so they override no environment variable and need
no restart. Today that is `current_project_id` and `frontend_config`.

`SetSettings` takes a batch and validates every entry — known key, editable,
value parses — **before** writing any of them, then writes them in one
transaction. A form with one bad field therefore changes nothing rather than
landing half-applied, which is why the mutation is plural.
`SetCurrentProjectID` checks `ProjectExists` **before** writing, which is what
keeps the validation failure above unreachable through the API, and `SetSetting`
routes that key through it so the check cannot be bypassed.

Afterwards the resolver (`internal/graph/api/settings.go`) invalidates:

- `InvalidateProject(old)` and `InvalidateProject(new)` — each drops the entire
  `gqlresponse:` prefix, which is what evicts the cached `currentProject` /
  `myCurrentProject` responses. Without it the old project is served until the
  30-second response-cache TTL expires.
- `InvalidateSettings()` — reloads the settings map here and broadcasts
  `InvalidationTypeSettings` so other instances reload now rather than on their
  own five-minute tick. See `notes/12-cache-invalidation.md`.

Admin UI: `/admin/settings`. The configuration section is a staged form: every
field edits a local draft, changed rows are badged, and one button commits them
all through `setSettings`. `useUnsavedChanges` guards navigating away. The
current project is deliberately **not** in that batch — it changes what every
end user sees, so it keeps its own confirm dialog and its own mutation. and a "Sett som
gjeldende prosjekt" action on the project overview page. Both are gated on
`settings:manage`, superadmin only.

## Next Steps

This configuration will be used by:
1. Database connection layer
2. HTTP server setup
3. Middleware configuration
4. All cmd/ tools (server, migrate, seed)
