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

**Only `current_project_id` is live.** It names the project every end user
sees, and is read by `Query.currentProject` / `myCurrentProject` /
`myProjects`, the church-admin statistics resolver, the Firebase token warmer,
and the two ladder-to-heaven URL handlers.

**Every other row is inert.** `log_level`, `db_log_queries`, `otel_enabled`,
`otel_sampling_ratio` and `ssf_debug_mode` duplicate environment variables that
`internal/config/config.go` reads instead, and nothing calls
`GetBoolSetting` / `GetIntSetting` / `GetFloatSetting`. `services.editableSettings`
is the allowlist that encodes this: the GraphQL `Setting.editable` field is
false for them, the admin UI renders them read-only, and `SetSetting` refuses
to write them. Wiring one up means reading it through the service *and* adding
it to that allowlist.

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

`SetCurrentProjectID` checks `ProjectExists` **before** writing, which is what
keeps the validation failure above unreachable through the API. `SetSetting`
routes that key through the same function so the check cannot be bypassed.

Afterwards the resolver (`internal/graph/api/settings.go`) invalidates:

- `InvalidateProject(old)` and `InvalidateProject(new)` — each drops the entire
  `gqlresponse:` prefix, which is what evicts the cached `currentProject` /
  `myCurrentProject` responses. Without it the old project is served until the
  30-second response-cache TTL expires.
- `InvalidateSettings()` — reloads the settings map here and broadcasts
  `InvalidationTypeSettings` so other instances reload now rather than on their
  own five-minute tick. See `notes/12-cache-invalidation.md`.

Admin UI: `/admin/settings` (the project picker plus a read-only view of the
inert rows) and a "Sett som gjeldende prosjekt" action on the project overview
page. Both are gated on `settings:manage`, superadmin only.

## Next Steps

This configuration will be used by:
1. Database connection layer
2. HTTP server setup
3. Middleware configuration
4. All cmd/ tools (server, migrate, seed)
