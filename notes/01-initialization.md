# Backend Initialization

**Date**: 2025-10-31

## Go Module Setup

**Module name**: `github.com/bcc-media/wayfarer`

## Dependencies Added

### Core Dependencies

1. **github.com/gin-gonic/gin** - Web framework
   - Lightweight HTTP router and framework
   - Fast performance, minimal overhead
   - Good middleware ecosystem

2. **github.com/jackc/pgx/v5** - PostgreSQL/CockroachDB driver
   - Native Go PostgreSQL driver
   - Best performance for CockroachDB
   - Supports connection pooling

3. **github.com/99designs/gqlgen** - GraphQL code generation
   - Type-safe GraphQL server generation
   - Schema-first approach
   - Good performance, minimal reflection

4. **github.com/sqlc-dev/sqlc** - SQL code generation
   - Generates type-safe Go from SQL queries
   - Catches SQL errors at compile time
   - No ORM overhead

5. **github.com/pressly/goose/v3** - Database migrations
   - Simple, reliable migration tool
   - Supports embedding migrations
   - Good CockroachDB support

6. **github.com/oklog/ulid/v2** - ULID generation
   - Lexicographically sortable IDs
   - Timestamp-based
   - More readable than UUIDs

7. **github.com/stretchr/testify** - Testing utilities
   - Assertions: `assert` and `require`
   - Test suites with setup/teardown
   - Mock generation

### Standard Library

- **log/slog** - Structured logging (Go 1.21+)
  - Built-in, no external dependency
  - Structured, leveled logging
  - JSON and text output

## Directory Structure Created

```
backend/
├── cmd/
│   ├── server/           # Main API server
│   ├── migrate/          # Migration runner
│   └── seed/             # Database seeder
├── internal/
│   ├── config/           # Configuration
│   ├── database/         # DB connection
│   ├── graph/            # GraphQL resolvers
│   ├── middleware/       # HTTP middleware
│   ├── services/         # Business logic
│   ├── seeder/           # Seeding logic
│   └── ulid/             # ID generation
├── pkg/
│   └── models/           # sqlc generated
├── migrations/           # SQL migrations
├── queries/              # SQL queries
├── testdata/             # Test fixtures
├── go.mod
├── go.sum
└── tools.go
```

## Tools Tracking

Created `tools.go` with build tag to ensure code generation tools are tracked in `go.mod`:
- gqlgen
- sqlc
- goose

This ensures `go mod tidy` doesn't remove these dependencies.

## Next Steps

1. Create configuration package with typed Config struct
2. Setup code generation (sqlc, gqlgen)
3. Implement ULID package
4. Create database layer with migrations
