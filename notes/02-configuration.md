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
- `JWT_SECRET` - JWT signing secret (empty for now, not validated)
- `JWT_ISSUER` - JWT issuer claim (default: "wayfarer")

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

## Next Steps

This configuration will be used by:
1. Database connection layer
2. HTTP server setup
3. Middleware configuration
4. All cmd/ tools (server, migrate, seed)
