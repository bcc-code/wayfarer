# Wayfarer Backend - Implementation Plan

**Created**: 2025-10-31
**Status**: Phase 1 Complete - Foundation Established

## Project Overview

Building a Go backend for the Wayfarer gamification system with three separate GraphQL APIs (User, Admin, M2M), using CockroachDB, and following best practices for configuration, testing, and deployment.

---

## Phase 1: Foundation ✅ COMPLETE

### 1.1 Project Setup ✅
- [x] Initialize backend directory structure
- [x] Initialize go.mod with all dependencies
- [x] Create tools.go for code generation tool tracking
- [x] Document initialization in notes/01-initialization.md

### 1.2 Configuration Management ✅
- [x] Create internal/config package with typed Config struct
- [x] Implement centralized env var loading (load once at startup)
- [x] Create configuration subsections (Server, Database, JWT, Log)
- [x] Write tests for config package
- [x] Document configuration in notes/02-configuration.md

### 1.3 ULID Generation ✅
- [x] Create internal/ulid package
- [x] Implement type-safe generators for all 14 entity types:
  - CH (Churches), US (Users), PR (Projects), EV (Events)
  - ST (SuperTeams), TM (Teams), SK (Streaks), SD (StreakRelevantDays)
  - CL (Challenges), AC (Achievements), RA (ReadingAchievements)
  - LT (ListeningAchievements), SA (ScoreAdjustments)
- [x] Add thread-safe ID generation with mutex
- [x] Create validation functions for each entity type
- [x] Write comprehensive tests (including concurrent generation)
- [x] Document ULID strategy in notes/05-ulid.md

### 1.4 Database Layer ✅
- [x] Create internal/database package with connection pooling
- [x] Convert schema.sql to Goose migration (00001_initial_schema.sql)
- [x] Embed migrations in database package
- [x] Create cmd/migrate tool with up/down/status commands
- [x] Test migrations against CockroachDB cloud instance
- [x] Verify all 21 tables created successfully
- [x] Document database setup in notes/04-database.md

---

## Phase 2: Code Generation & Queries 🔄 IN PROGRESS

### 2.1 SQLc Configuration ⏳ NEXT
- [ ] Create sqlc.yaml configuration file
  - Set SQL engine to PostgreSQL
  - Configure output directory (pkg/models)
  - Set package name and emit settings
- [ ] Write initial SQL queries in queries/ directory:
  - churches.sql (CRUD operations)
  - users.sql (CRUD + lookups by church, gender, age)
  - projects.sql (CRUD + archived filtering)
  - events.sql (CRUD + project filtering)
  - teams.sql (CRUD + join code lookup)
  - achievements.sql (CRUD + type filtering)
  - user_progress.sql (achievement tracking, challenge completions)
- [ ] Run `sqlc generate` to create type-safe models
- [ ] Test generated code compiles

### 2.2 GraphQL Configuration ⏳
- [ ] Create gqlgen.yml configuration
  - Reference ../gql/ schemas directly (no symlink)
  - Configure three separate APIs (user, admin, m2m)
  - Set up resolver generation paths
  - Map GraphQL types to Go types (ULID IDs, etc.)
- [ ] Run `gqlgen generate` to create resolvers
- [ ] Verify generated code structure

### 2.3 Makefile Creation ⏳
- [ ] Create Makefile with common targets:
  - `make generate` - runs both sqlc and gqlgen
  - `make migrate-up` - runs migrations up
  - `make migrate-down` - runs migrations down
  - `make migrate-status` - shows migration status
  - `make dev` - starts development server with hot reload
  - `make test` - runs all tests
  - `make build` - builds all binaries
  - `make clean` - cleans build artifacts
- [ ] Test all Makefile targets

### 2.4 Documentation ⏳
- [ ] Create notes/03-code-generation.md documenting:
  - sqlc configuration and workflow
  - gqlgen configuration and workflow
  - Makefile usage
  - Code generation best practices

---

## Phase 3: HTTP Layer & Middleware

### 3.1 Middleware Implementation
- [ ] Create internal/middleware/logger.go
  - Request/response logging with slog
  - Log request ID, method, path, status, duration
  - Structured log fields
  - Accepts LogConfig from main
- [ ] Create internal/middleware/auth.go
  - Extract Authorization header
  - Log JWT token (for now, no validation)
  - Accept all tokens (placeholder for future validation)
  - Accepts JWTConfig from main
- [ ] Create internal/middleware/recovery.go
  - Panic recovery with slog error logging
  - Return 500 with error message
- [ ] Write tests for all middleware
- [ ] Document middleware patterns

### 3.2 GraphQL Server Setup
- [ ] Configure three GraphQL endpoints:
  - `/graphql/user` - User API
  - `/graphql/admin` - Admin API
  - `/graphql/m2m` - M2M API
- [ ] Wire up gqlgen handlers
- [ ] Apply middleware to GraphQL routes
- [ ] Add GraphQL playground endpoints (dev only)

### 3.3 Main Server Implementation
- [ ] Create cmd/server/main.go
  - Load config once at startup
  - Initialize slog logger (JSON/text based on config)
  - Connect to database with pooling
  - Create Gin router
  - Register middleware (logger, auth, recovery)
  - Register GraphQL endpoints
  - Add health check endpoint (/health)
  - Implement graceful shutdown
- [ ] Test server starts and responds
- [ ] Test all three GraphQL APIs accessible

### 3.4 Documentation
- [ ] Create notes/06-graphql.md documenting:
  - Three API separation strategy
  - Resolver implementation patterns
  - Middleware flow
  - Server configuration

---

## Phase 4: Database Seeding

### 4.1 Seeder Package
- [ ] Create internal/seeder package structure
- [ ] Implement entity seeders:
  - seeder.go (orchestration)
  - churches.go (S, L, XL churches in various countries)
  - users.go (users across churches, various ages/genders)
  - projects.go (projects with branding)
  - events.go (events within projects)
  - teams.go (teams and super_teams)
  - achievements.go (all 4 types)
  - challenges.go (various challenges)
  - progress.go (user progress data)
- [ ] Use sqlc-generated queries for inserts
- [ ] Use ULID generators for all IDs

### 4.2 Seed Command
- [ ] Create cmd/seed/main.go
  - Load config at startup
  - Connect to database
  - Support `--clear` flag to truncate tables
  - Support `--count` flag for data volume
  - Run seeder with configuration
  - Log progress with slog
- [ ] Create testdata/ fixtures for repeatable seeds
- [ ] Test seeding against test database
- [ ] Document seeding strategy

---

## Phase 5: Testing Infrastructure

### 5.1 Test Helpers
- [ ] Create internal/database/testutil.go:
  - SetupTestDB() - connects to test database
  - TeardownTestDB() - cleans up
  - SeedTestData() - seeds minimal test data
  - Transaction-based test isolation
- [ ] Create test fixtures in testdata/
- [ ] Document testing patterns in notes/07-testing.md

### 5.2 Test Coverage
- [ ] Write integration tests for:
  - Database connection and migrations
  - ULID generation (already done ✅)
  - Configuration loading (already done ✅)
  - Middleware behavior
  - GraphQL query execution
  - Database queries via sqlc
- [ ] Write unit tests for:
  - Business logic in services/
  - Seeder functions
- [ ] Ensure all tests pass with `make test`

---

## Phase 6: Local Development Environment

### 6.1 Docker Compose
- [ ] Create docker-compose.yml:
  - CockroachDB service (local instance)
  - Backend service with hot reload (air)
  - Volume persistence for database
  - Network configuration
  - Environment variable setup
- [ ] Create .env.example with all configuration options
- [ ] Test full local environment

### 6.2 Documentation
- [ ] Create notes/08-docker.md documenting:
  - Docker Compose setup
  - Local development workflow
  - Environment variables
  - Troubleshooting
- [ ] Create backend/README.md with:
  - Quick start guide
  - Development workflow
  - Common commands
  - Project structure overview

---

## Phase 7: Final Integration & Testing

### 7.1 End-to-End Testing
- [ ] Test complete workflow:
  - Run migrations
  - Seed database
  - Start server
  - Execute GraphQL queries via all three APIs
  - Verify data integrity
- [ ] Test with both local and cloud databases
- [ ] Verify all environment configurations work

### 7.2 Documentation Finalization
- [ ] Review and update all notes for accuracy
- [ ] Create comprehensive backend/README.md
- [ ] Update notes/00-overview.md with final architecture
- [ ] Document any gotchas or best practices discovered

---

## Current Status

**Phase 1 Complete**: Foundation is solid and tested
- ✅ Project structure established
- ✅ Configuration management working
- ✅ ULID generation with 14 entity types
- ✅ Database migrations tested against CockroachDB cloud
- ✅ All 21 tables created successfully

**Next Steps**:
1. Configure sqlc and write initial queries
2. Create Makefile for development workflow
3. Implement middleware layer
4. Set up GraphQL servers

**Estimated Remaining Work**:
- Phase 2: ~2-3 hours
- Phase 3: ~2-3 hours
- Phase 4: ~2 hours
- Phase 5: ~2 hours
- Phase 6: ~1 hour
- Phase 7: ~1 hour

**Total**: ~10-14 hours to complete full backend setup

---

## Notes

- No symlinks used - reference ../gql/ directly in configs
- All env vars loaded once at startup via config.Load()
- Migrations embedded in binaries via Go embed
- Type-safe ID generation for all entities
- Three separate GraphQL APIs for different consumers
- Testing against real CockroachDB cloud instance
