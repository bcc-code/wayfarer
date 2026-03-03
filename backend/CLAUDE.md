# Backend — Go API Server

## Tech Stack

- **Language**: Go 1.25
- **HTTP**: Gin
- **GraphQL**: gqlgen (code generation)
- **Database**: PostgreSQL via pgx/v5, queries via sqlc
- **Migrations**: goose v3 (embedded)
- **Caching**: Ristretto (via `internal/cache`)
- **Auth**: JWT with Firebase + Auth0
- **Logging**: slog (structured)
- **Testing**: testify + testcontainers (e2e), mockery (mocks)
- **Tracing**: OpenTelemetry + Jaeger

## Directory Layout

```
backend/
├── cmd/                          # Executables (server, migrate, seed, debug, etc.)
├── internal/
│   ├── graph/
│   │   ├── api/                  # GraphQL resolvers + generated code
│   │   │   ├── generated.go      # ⚠ GENERATED — do not edit
│   │   │   ├── model/models_gen.go # ⚠ GENERATED — do not edit
│   │   │   ├── *.resolvers.go    # ⚠ GENERATED stubs — implement logic here but don't add new functions
│   │   │   └── *.go              # ✅ Safe to edit — put helper/conversion functions here
│   │   ├── scalars/              # Custom scalars (DateTime, Date, HTML, Markdown)
│   │   └── directives/           # @requireRole directive implementation
│   ├── services/                 # Business logic (roles, settings, push, email, etc.)
│   ├── handlers/                 # HTTP handlers (auth, webhooks, uploads)
│   ├── database/
│   │   ├── migrations/           # goose SQL migrations (00001–00089+)
│   │   ├── queries/              # sqlc SQL query definitions
│   │   └── sqlc/                 # ⚠ GENERATED — do not edit
│   ├── middleware/               # HTTP middleware (JWT auth, context injection)
│   ├── loaders/                  # DataLoaders (~57 loaders for N+1 prevention)
│   ├── cache/                    # Ristretto cache with registry and named keys
│   ├── config/                   # Config loaded once from env at startup
│   ├── ulid/                     # Prefixed ULID generators
│   ├── auth0/                    # Auth0 integration
│   ├── firebase/                 # Firebase integration
│   └── members/                  # Members sync
├── e2e/                          # E2E tests (testcontainers + GraphQL client)
├── Makefile                      # Build, test, codegen, migrate commands
├── gqlgen.yml                    # GraphQL codegen config
├── sqlc.yaml                     # SQL codegen config
└── .mockery.yml                  # Mock generation config
```

## Key Commands

```bash
make generate      # Run gqlgen + sqlc codegen
make test          # Unit tests (go test -v -race ./...)
make test-e2e      # E2E tests (requires Docker)
make fmt           # Format code (go fmt ./...)
make lint          # Run golangci-lint
make build         # Build all binaries
make migrate       # Run migrations (⚠ requires approval)
make seed          # Seed database (⚠ NEVER run without explicit permission)
```

## Conventions

### Resolver Pattern

Resolvers follow this order:

1. Extract user from context: `middleware.GetUserID(ctx)`
2. Check authorization via `r.RoleService`
3. Validate input
4. Generate ID: `ulid.New[Entity]ID()`
5. Build sqlc params and execute query
6. Invalidate relevant caches
7. Convert sqlc row → GraphQL model and return

Put **helper and conversion functions** in separate `.go` files (e.g., `challenges.go`), not in `*.resolvers.go` files which are overwritten by codegen.

### Service Layer

- Accept **interfaces**, not concrete types (for testability)
- Constructor: `NewXxxService(queries XxxQuerier, cache *cache.CacheWithRegistry) *XxxService`
- Use **cache-before-query** pattern: check cache → query DB for misses → store in cache
- Log errors with `slog.Error()` for non-critical failures
- Use typed constants for enums: `type RoleType string`

### SQL Queries (sqlc)

- File location: `internal/database/queries/[entity].sql`
- **Named parameters only**: `@paramName::type` — never use `$1`, `$2`
- Result type annotations: `:many`, `:one`, `:exec`, `:execresult`
- Naming: `Get[Entity]ByXYZ`, `Count[Entity]`, `Create[Entity]`, `Update[Entity]`, `Delete[Entity]`
- Array parameters: `WHERE id = ANY(@ids::text[])`

### DataLoaders

- Defined in `internal/loaders/` — one file per domain
- Use **composite key structs** for multi-parameter loaders
- Always check cache first in batch functions, query DB only for misses
- Return results in **same order as input keys**
- Use in resolvers: `r.Loaders.XxxLoader.Load(ctx, key)` → call `thunk()`

### Error Handling

- Always wrap with context: `fmt.Errorf("failed to create challenge: %w", err)`
- Auth errors: `"user not authenticated"`, `"unauthorized to [action]"`
- Never expose internal implementation details in error messages

### Cache Invalidation

- After mutations, invalidate relevant cache entries
- Use named cache key functions from `internal/cache/` (e.g., `cache.UserKey(id)`, `cache.ProjectKey(id)`)

### Testing

**Unit tests** (service-level): Use mockery-generated mocks from `.mockery.yml`

```go
mockQueries := mocks.NewMockRoleQuerier(t)
service := NewRoleService(mockQueries, newTestCache())
mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{...}).Return(true, nil)
```

**E2E tests** (`e2e/`): Use testcontainers for real PostgreSQL

```go
require.NoError(t, dbMgr.Clean(ctx))
data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())  // deterministic seed
router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
client := testutil.NewGraphQLClient(router)
resp := client.WithAuth(adminToken).MustExecute(t, query, variables)
```

- Use `require` for setup (fail fast), `assert` for assertions
- Use subtests with `t.Run()`
- Seed with deterministic seed (42) for reproducibility
