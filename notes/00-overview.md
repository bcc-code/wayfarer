# Wayfarer Backend - Architecture Overview

**Date Started**: 2025-10-31

## Project Overview

Wayfarer is a gamification system backend built with Go, GraphQL (gqlgen), and PostgreSQL. It provides a unified GraphQL API with role-based access control for different consumers: end users, administrators, and machine-to-machine integrations.

## Architecture Decisions

### Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin (lightweight, fast HTTP router)
- **GraphQL**: gqlgen (type-safe code generation from schemas)
- **Database**: PostgreSQL
- **SQL Access**: sqlc (generates type-safe Go from SQL queries)
- **Migrations**: goose (embedded into binary)
- **Logging**: slog (structured logging, standard library)
- **Testing**: testify (assertions and test suites)
- **IDs**: ULID with 2-character prefixes (e.g., US01ARZ3NDEKTSV4RRFFQ69G5FAV)

### Project Structure

```
wayfarer/
├── backend/              # Go backend (this implementation)
├── frontend/             # Nuxt.js frontend
├── gql/                  # Shared GraphQL schemas
├── notes/                # Implementation notes
├── schema.sql            # Reference database schema
└── docker-compose.yml    # Local development environment
```

### Backend Structure

```
backend/
├── cmd/
│   ├── server/           # Main API server
│   ├── migrate/          # Migration runner (embedded migrations)
│   └── seed/             # Database seeder
├── internal/
│   ├── config/           # Configuration loading (env vars → typed struct)
│   ├── database/         # DB connection, migration embed
│   ├── graph/            # GraphQL resolvers (gqlgen generated)
│   ├── middleware/       # HTTP middleware (auth, logging, recovery)
│   ├── services/         # Business logic layer
│   ├── seeder/           # Database seeding logic
│   └── ulid/             # Type-safe ULID generation
├── pkg/
│   └── models/           # sqlc generated models
├── migrations/           # SQL migration files
├── queries/              # SQL query definitions for sqlc
└── testdata/             # Test fixtures and seed data
```

### Key Design Patterns

#### 1. Configuration Management
- **Pattern**: Load all environment variables once at startup in `main()`
- **Structure**: Typed `Config` struct with nested configuration sections
- **Distribution**: Pass config subsections to components via dependency injection
- **Rule**: NO direct environment variable access anywhere except `main()`

#### 2. Unified GraphQL API
- **Single Endpoint**: All consumers use the same GraphQL endpoint
- **Role-Based Access**: Operations protected by `@requireRole` directive
- **Schema Files**:
  - `gql/shared.graphqls` - Types, enums, inputs, interfaces
  - `gql/schema.graphqls` - Query and Mutation root types
- **Roles**: `user`, `admin`, `m2m`, `superadmin`
- **Authentication**: JWT tokens determine which role(s) the client has access to

#### 3. Authentication Strategy (Phase 1)
- **Current**: JWT middleware that logs Authorization header, accepts all tokens
- **Purpose**: Establish middleware pattern, ready for real validation later
- **Future**: Parse JWT, validate signature, extract user context

#### 4. Database Access
- **Migrations**: Goose migrations embedded into binaries using `embed` package
- **Queries**: Hand-written SQL in `queries/*.sql`, type-safe code via sqlc
- **Connection**: Single connection pool initialized at startup, passed to components

#### 5. ID Generation
- **Format**: 2-char prefix + 26-char ULID = 28 chars total
- **Type Safety**: Separate functions per entity (e.g., `NewUserID()`, `NewChurchID()`)
- **Prefixes**: CH, US, PR, EV, ST, TM, SK, CL, AC, RA, LT, SA (13 entity types)

#### 6. Testing Strategy
- **Framework**: testify for assertions and test suites
- **Database**: Use provided test database, transaction-based isolation
- **Patterns**: Table-driven tests, test helpers, fixture data in testdata/
- **Seeding**: Can seed test database for manual testing

## API Roles in Detail

### User Role (`@requireRole(roles: ["user"])`)
**Purpose**: End-user operations from mobile/web apps

**Key Operations**:
- View projects, events, achievements, challenges
- Join projects/events/teams
- Track progress (reading, listening, streaks)
- View leaderboards
- Update profile

### Admin Role (`@requireRole(roles: ["admin"])`)
**Purpose**: Administrative operations

**Key Operations**:
- Full CRUD for all entities
- Bulk operations
- Project/event cloning
- Archiving
- Score adjustments
- User management

### M2M Role (`@requireRole(roles: ["m2m"])`)
**Purpose**: External systems notify Wayfarer

**Key Operations**:
- Notify reading article completion
- Notify listening track completion
- Notify streak activity
- Award achievements
- Record challenge completions

### Superadmin Role (`@requireRole(roles: ["superadmin"])`)
**Purpose**: System-level operations

**Key Operations**:
- All admin operations plus system management

## Development Workflow

1. **Migrations**: Write SQL → run `make migrate-up`
2. **Queries**: Write SQL → run `make generate` (sqlc)
3. **GraphQL**: Update schema → run `make generate` (gqlgen)
4. **Testing**: Write tests → run `make test`
5. **Seeding**: Need data → run `make seed`
6. **Development**: Run `make dev` for hot reload

## Notes Organization

Each major implementation phase has its own note file:
- `01-initialization.md` - Go module setup, dependencies
- `02-configuration.md` - Config struct, env var strategy
- `03-code-generation.md` - sqlc and gqlgen setup
- `04-database.md` - Migrations, connection, queries
- `05-ulid.md` - ID generation strategy
- `06-graphql.md` - Resolver implementation
- `07-testing.md` - Test infrastructure
- `08-docker.md` - Local development setup

## External Dependencies

- PostgreSQL test database (connection string in .env file)

## Next Steps

See individual note files for implementation details and decisions made during development.
