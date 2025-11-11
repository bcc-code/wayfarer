# Running the GraphQL Server

This document explains how to run the Wayfarer GraphQL server.

## Prerequisites

1. **Database**: Ensure you have access to a PostgreSQL instance
2. **Environment Variables**: Set up your environment variables (see Configuration below)
3. **Migrations**: Run database migrations before starting the server

## Configuration

The server requires the `DATABASE_URL` environment variable to be set. Other variables are optional:

```bash
# Required
export DATABASE_URL="postgresql://user:pass@host:port/database"

# Optional (with defaults)
export SERVER_HOST="0.0.0.0"          # Default: 0.0.0.0
export SERVER_PORT="8080"              # Default: 8080
export ENVIRONMENT="development"       # Default: development (use "production" to disable playgrounds)
export LOG_LEVEL="info"                # Default: info (options: debug, info, warn, error)
export LOG_FORMAT="json"               # Default: json (options: json, text)
```

## Running Migrations

Before starting the server for the first time, run the database migrations:

```bash
make migrate

# Or manually:
go run ./cmd/migrate -direction up
```

## Starting the Server

### Development Mode

The easiest way to run the server in development:

```bash
make dev
```

This will:
- Start the server with live reloading (via `go run`)
- Enable GraphQL Playground UI on all three endpoints
- Use development defaults
- Show detailed logs

### Building and Running

To build and run the production binary:

```bash
# Build the binary
make build

# Run the server
./bin/server
```

## API Endpoints

Once the server is running (default: `http://localhost:8080`), you'll have access to three separate GraphQL APIs:

### 1. User API - `/graphql/user`

**Purpose**: End-user facing API for mobile and web applications

**GraphQL Endpoint**: `POST http://localhost:8080/graphql/user`

**Playground** (development only): `GET http://localhost:8080/graphql/user`

**Features**:
- View projects and events
- Join projects, events, and teams
- View profile and achievements
- Update avatar

### 2. Admin API - `/graphql/admin`

**Purpose**: Administrative interface for managing the entire system

**GraphQL Endpoint**: `POST http://localhost:8080/graphql/admin`

**Playground** (development only): `GET http://localhost:8080/graphql/admin`

**Features**:
- Full CRUD operations on all entities
- User management
- Project and event management
- Challenge and achievement creation
- Team management
- Score adjustments

### 3. M2M API - `/graphql/m2m`

**Purpose**: Machine-to-machine communication for external systems

**GraphQL Endpoint**: `POST http://localhost:8080/graphql/m2m`

**Playground** (development only): `GET http://localhost:8080/graphql/m2m`

**Features**:
- Award/revoke achievements
- Complete challenges
- Track reading/listening progress
- Update streaks
- Bulk operations for performance

## Health Check

The server provides a health check endpoint:

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status": "ok"}
```

## Authentication

Currently, the server uses a placeholder JWT middleware that:
- Logs the Authorization header for debugging
- Accepts all requests (even without authentication)
- Prints warnings when Authorization header is missing

**TODO**: Implement actual JWT validation with user context extraction.

## GraphQL Playground

In development mode (`ENVIRONMENT=development`), each API endpoint has an interactive GraphQL Playground available in your browser:

- User API: http://localhost:8080/graphql/user
- Admin API: http://localhost:8080/graphql/admin
- M2M API: http://localhost:8080/graphql/m2m

The playground provides:
- Schema documentation
- Query/mutation autocomplete
- Query history
- Request headers configuration

**Note**: Playgrounds are automatically disabled in production mode.

## Example Queries

Since the resolver implementations are not yet complete, queries will return "not implemented" errors. However, you can test the server setup:

### User API - Get Projects

```graphql
query GetProjects {
  projects {
    id
    name
    description
  }
}
```

### Admin API - List Users

```graphql
query ListUsers {
  users(filter: { churchId: "CH01ARZ3NDEKTSV4RRFFQ69G5FAV" }) {
    id
    membersId
    firstName
    lastName
  }
}
```

### M2M API - Award Achievement

```graphql
mutation AwardAchievement {
  awardAchievement(
    userId: "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
    achievementId: "AC01ARZ3NDEKTSV4RRFFQ69G5FAV"
  ) {
    id
    name
  }
}
```

## Logs

The server uses structured logging (slog) with two formats:

**JSON Format** (default, production-friendly):
```json
{"time":"2025-10-31T14:30:00Z","level":"INFO","msg":"Starting server","address":"0.0.0.0:8080"}
```

**Text Format** (development-friendly):
```
time=2025-10-31T14:30:00.000Z level=INFO msg="Starting server" address=0.0.0.0:8080
```

Set `LOG_FORMAT=text` for easier reading during development.

## Graceful Shutdown

The server handles `SIGINT` (Ctrl+C) and `SIGTERM` signals gracefully:

1. Stops accepting new connections
2. Waits up to 5 seconds for active requests to complete
3. Closes database connections
4. Exits cleanly

## Troubleshooting

### "Failed to connect to database"

**Issue**: Cannot connect to PostgreSQL

**Solutions**:
- Verify `DATABASE_URL` is set correctly
- Check network connectivity to database
- Ensure database credentials are valid
- Verify SSL certificates if using `sslmode=verify-full`

### "Migration failed"

**Issue**: Database migrations didn't apply

**Solutions**:
- Check database permissions
- Verify schema hasn't been manually modified
- Check migration logs for specific errors
- Try `make migrate-status` to see current state

### Port Already in Use

**Issue**: `bind: address already in use`

**Solutions**:
- Change the port: `export SERVER_PORT=8081`
- Stop other processes using the port
- Find the process: `lsof -i :8080`

## Next Steps

1. **Implement Resolvers**: The resolver functions in `*.resolvers.go` files need implementation
2. **Add JWT Validation**: Replace the placeholder JWT middleware with actual token validation
3. **Write SQL Queries**: Create SQL queries and use sqlc to generate type-safe database code
4. **Add Tests**: Write integration tests for the GraphQL endpoints
5. **Add Seeding**: Create a seed command to populate the database with test data
