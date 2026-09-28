# Database Setup

**Date**: 2025-10-31

## Overview

Database layer using PostgreSQL with connection pooling and embedded migrations.

## Technology Stack

- **Driver**: pgx/v5 (native Go PostgreSQL driver)
- **Migrations**: Goose v3 (embedded in binary)
- **Connection Pooling**: pgxpool (from pgx/v5)

## Connection Pooling

The `internal/database` package provides connection pooling with configuration from the `DatabaseConfig`:

```go
type DB struct {
    Pool *pgxpool.Pool
}
```

### Configuration
- `MaxOpenConns`: Maximum number of open connections (default: 25)
- `MaxIdleConns`: Minimum number of idle connections (default: 5)
- `ConnMaxLifetime`: Maximum connection lifetime (default: 5 minutes)
- `ConnMaxIdleTime`: Maximum connection idle time (default: 5 minutes)

### Usage

```go
cfg, _ := config.Load()
db, err := database.Connect(ctx, cfg.Database)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Use db.Pool for queries
conn, err := db.Pool.Acquire(ctx)
```

## Migrations

Migrations are stored in `internal/database/migrations/` and embedded into binaries using Go's `embed` package.

### Migration Files

- **Location**: `internal/database/migrations/`
- **Format**: Goose SQL migrations
- **Naming**: `00001_initial_schema.sql`, `00002_add_feature.sql`, etc.
- **Embedded**: Via `//go:embed migrations/*.sql` in `database` package

### Initial Schema

The initial migration (`00001_initial_schema.sql`) creates all 21 tables:

**Core Tables**:
- churches
- users
- projects
- events
- super_teams
- teams
- streaks
- streak_relevant_days
- challenges
- achievements

**Type-specific Achievement Tables**:
- reading_achievements
- reading_achievement_articles
- listening_achievements
- listening_achievement_tracks
- streak_achievements

**Junction Tables**:
- user_projects
- user_events
- team_members

**Progress Tracking**:
- user_achievements
- team_achievements
- super_team_achievements
- user_challenge_completions
- user_reading_progress
- user_listening_progress
- user_streak_activity

**Audit/Activity**:
- score_adjustments

## Migration Command

The `cmd/migrate` tool runs migrations with embedded SQL files.

### Usage

```bash
# Run all pending migrations
./migrate -direction up

# Roll back last migration
./migrate -direction down

# Show migration status
./migrate -direction status
```

### Implementation Details

- Uses `database/sql` with pgx stdlib driver for goose compatibility
- Reads `DATABASE_URL` from environment via config
- Embeds migrations from `database.Migrations`
- Logs all operations with slog

### Example

```bash
export DATABASE_URL="postgresql://user:pass@host:26257/db?sslmode=verify-full"
./migrate -direction up
```

Output:
```
2025-10-31T12:00:00Z INFO starting migration direction=up
2025-10-31T12:00:01Z INFO database connection established
2025-10-31T12:00:02Z INFO goose: OK   00001_initial_schema.sql
2025-10-31T12:00:02Z INFO migration completed successfully
```

## Database Schema Highlights

### ID Format
All primary keys use 28-character ULIDs with 2-character prefixes:
- Format: `XX[0-9A-Z]{26}`
- Validated with CHECK constraints
- Generated in application code (see notes/05-ulid.md)

### Key Patterns

1. **Cascading Deletes**: Most foreign keys use `ON DELETE CASCADE`
2. **Soft Deletes**: Projects have `archived` boolean flag
3. **Timestamps**: All core tables have `created_at` and `updated_at`
4. **Indexes**: Strategic indexes on foreign keys, dates, and common queries
5. **Constraints**: CHECK constraints for enums and data validation

### Achievement Types

Four types with type-specific tables:
- **SIMPLE**: No additional data
- **READING**: Articles in `reading_achievement_articles`
- **LISTENING**: Tracks in `listening_achievement_tracks`
- **STREAK**: Streak requirements in `streak_achievements`

## Next Steps

With database and migrations set up, next steps:
1. Write sqlc queries for type-safe database access
2. Configure sqlc code generation
3. Generate Go models from SQL
4. Create GraphQL resolvers using generated models

## Testing

Database connection can be tested:

```go
db, _ := database.Connect(ctx, cfg.Database)
if err := db.Ping(ctx); err != nil {
    t.Fatal("database connection failed")
}
```

Test database provided: see .env file

## `schema.sql` is a partial reference, not the schema (2026-09-21)

`schema.sql` at the repo root is described as the reference schema, and CLAUDE.md
points at it for entity structure. **It documents 46 of the 83 tables the
migrations create.** The authority is always
`backend/internal/database/migrations/` — that is also what sqlc reads
(`sqlc.yaml`: `schema: "./internal/database/migrations"`), so codegen is
unaffected by the drift. Only humans reading `schema.sql` are misled.

`score_journal` was added on 2026-09-21 while building the project activity
trend. Still missing, as of that date:

```
bulk_jobs, external_content_events, file_uploads, leaderboard_apply_queue,
leaderboard_configs, leaderboard_event_{churches,persons,superteams,teams},
leaderboard_project_{churches,persons,superteams,teams},
listening_achievements, listening_achievement_tracks,
listening_achievement_track_translations, pending_consent_events,
phrase_async_jobs, quiz_sessions, quiz_session_access, reading_achievements,
reading_achievement_articles, reading_achievement_article_translations,
settings, ssf_content, ssf_content_translations, streaks, streak_relevant_days,
streak_translations, super_team_translations, team_translations,
translation_hashes, user_challenge_enrollments, user_consents,
user_listening_progress, user_reading_progress, user_streak_activity
```

Several of those are entities CLAUDE.md documents in detail — `streaks`,
`user_streak_activity`, `reading_achievements`, `user_reading_progress` — so the
gap is not confined to incidental tables.

**Do not hand-patch the remaining 37.** Transcribing by hand is how the file got
into this state, and it is actively error-prone: `score_journal`'s `source_type`
CHECK was replaced by three later migrations, so copying the `CREATE TABLE` from
migration 00017 would have documented a constraint missing `QUIZ`, `PLUGIN` and
`BET`. The fix is to generate it — run the migrations against a scratch database
and `pg_dump --schema-only` — which also keeps it honest from then on. Worth a
`make schema` target so it cannot drift again.

Reproduce the gap with:

```bash
grep -oiE "^CREATE TABLE (IF NOT EXISTS )?[a-z_]+" schema.sql |
  awk '{print tolower($NF)}' | sort -u > /tmp/a
grep -rhoiE "CREATE TABLE (IF NOT EXISTS )?[a-z_]+" \
  backend/internal/database/migrations/*.sql |
  awk '{print tolower($NF)}' | sort -u > /tmp/b
comm -13 /tmp/a /tmp/b
```
