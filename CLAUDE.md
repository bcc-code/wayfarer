# Wayfarer

Wayfarer is a gamification system used as the core for a multitude of projects, usually bible studies associated with a youth camp.

## Project Structure

```
wayfarer/
├── backend/          # Go API server (Gin + gqlgen + sqlc)
├── frontend/         # Nuxt 4 SPA (Vue 3 + TypeScript + urql)
├── gql/              # Shared GraphQL schema definitions (*.graphqls)
├── schema.sql        # Reference database schema (PostgreSQL)
├── notes/            # Implementation documentation
├── docker/           # Docker and observability configs
└── docs/             # General documentation
```

See `backend/CLAUDE.md` and `frontend/CLAUDE.md` for stack-specific conventions.

## Rules

- No signatures on commits
- When creating commits sign with: Assisted by [MODEL] via [Tool]
- Generate unit tests for functions you write. Use the unit tests to verify correctness. When mocking use mockery, with the config `.mockery.yml`.
- Put notes into notes folder. Before doing a big investigation check if we already have notes on the system.
- Update the notes if you make changes to schemas and other things that invalidate the notes.
- Do not EVER automatically run the seed script!
- NEVER SEED WITHOUT EXPLICIT PERMISSION
- Do not run migration without explicit approval!
- To run codegen in the backend, run `make generate`
- To run codegen in the frontend, run `pnpm codegen` in the frontend folder
- Before claiming any backend work is finished, run `make fmt` and `make test` in the backend folder. All tests must pass.

## Generated Files — Do NOT Edit

These files are overwritten by codegen. Make changes in their source files instead.

| Generated file                                   | Source                                          | Regenerate with                                                                                  |
| ------------------------------------------------ | ----------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `backend/internal/graph/api/generated.go`        | `gql/*.graphqls`                                | `make generate` (in backend/)                                                                    |
| `backend/internal/graph/api/model/models_gen.go` | `gql/*.graphqls` + `gqlgen.yml`                 | `make generate` (in backend/)                                                                    |
| `backend/internal/graph/api/*.resolvers.go`      | `gql/*.graphqls`                                | `make generate` (in backend/) — **stubs only**; custom logic in non-resolver `.go` files is safe |
| `backend/internal/database/sqlc/*.go`            | `backend/internal/database/queries/*.sql`       | `make generate` (in backend/)                                                                    |
| `backend/internal/services/*/mocks/*.go`         | Interfaces in services                          | `make generate` (in backend/)                                                                    |
| `frontend/app/api/generated.ts`                  | `gql/*.graphqls` + `frontend/**/*.{vue,ts,gql}` | `pnpm codegen` (in frontend/)                                                                    |

## Database

The database schema is defined in `schema.sql` and uses PostgreSQL.

### ID Format

All primary keys use prefixed ULIDs for better readability and debugging:

- Format: `XX + 26-character ULID = 28 characters total`
- Example: `CH01ARZ3NDEKTSV4RRFFQ69G5FAV` (Church), `US01ARZ3NDEKTSV4RRFFQ69G5FAV` (User)
- ULIDs are lexicographically sortable by timestamp
- Must be generated in application code using `ulid.New[Entity]ID()`

### ID Prefix Reference

| Prefix | Entity                       | Generator                     |
| ------ | ---------------------------- | ----------------------------- |
| `CH`   | Churches                     | `ulid.NewChurchID()`          |
| `US`   | Users                        | `ulid.NewUserID()`            |
| `UR`   | User Roles                   | `ulid.NewUserRoleID()`        |
| `PR`   | Projects                     | `ulid.NewProjectID()`         |
| `EV`   | Events                       | `ulid.NewEventID()`           |
| `ST`   | SuperTeams                   | `ulid.NewSuperTeamID()`       |
| `TM`   | Teams                        | `ulid.NewTeamID()`            |
| `SK`   | Streaks                      | `ulid.NewStreakID()`          |
| `CL`   | Challenges                   | `ulid.NewChallengeID()`       |
| `AC`   | Achievements                 | `ulid.NewAchievementID()`     |
| `RA`   | Reading Achievement Articles | —                             |
| `LT`   | Listening Achievement Tracks | —                             |
| `SA`   | Score Adjustments            | `ulid.NewScoreAdjustmentID()` |
| `PS`   | Push Subscriptions           | —                             |
| `PN`   | Push Notification Log        | —                             |

### Core Entities

- **Churches** (`churches`) — Organizations users belong to. Categorized by size (S, L, XL).
- **Users** (`users`) — End users. Each belongs to one church, can participate in multiple projects/events.
- **Projects** (`projects`) — Top-level groupings containing branding, events, achievements, challenges, and streaks. Projects have start/end dates and can be archived. Every other entity below is scoped to a project.
- **Events** (`events`) — In-person or time-bound events within a project. Inherit from their parent project.
- **Teams** (`teams`) — Groups of users within a project. Users join via unique join codes.
- **SuperTeams** (`super_teams`) — Collections of multiple teams (two-level hierarchy).
- **Streaks** (`streaks`) — Activity tracking over time. Uses `DATERANGE[]` for relevant date ranges.
- **Challenges** (`challenges`) — Tasks users can complete. Belong to a project, optionally to an event.
- **Achievements** (`achievements`) — Rewards users can earn. Four types: Simple, Reading (→ `reading_achievement_articles`), Listening (→ `listening_achievement_tracks`), Streak (→ `streaks`).

### Progress Tracking

User progress is tracked through junction tables:

- `user_projects` / `user_events` — Project/event participation
- `team_members` — Team membership
- `user_achievements` / `team_achievements` / `super_team_achievements` — Achievement awards
- `user_challenge_completions` — Challenge completion
- `user_reading_progress` / `user_listening_progress` — Reading/listening progress
- `user_streak_activity` — Daily streak activity

### Scoring

Scores are calculated on-demand from achievements and score adjustments. No pre-aggregated score tables. Score adjustments are logged in `score_adjustments` for audit purposes.

## GraphQL API

The system exposes a unified GraphQL API defined in `gql/`:

- **Schema**: `shared.graphqls` (types, enums, inputs, interfaces) + `schema.graphqls` (Query/Mutation roots) + domain-specific files
- **Access Control**: `@requireRole` directive — **only on mutations, never on queries**
- **Roles**: `user`, `admin`, `m2m`, `superadmin`
- **Resolver stubs** are generated — do not add custom functions to `*.resolvers.go` files; they will be overwritten by `make generate`. Put helper functions in separate `.go` files in the same package.

## Concepts

### Projects

Projects are the top-level groupings. Every other concept is associated with _one_ project. Users are top-level and not scoped to a single project. Projects contain branding (logo, colors, fonts), events, points, achievements and leaderboards.

### Events

Each project can contain multiple events, usually an in-person event. Events have their own points, achievements and leaderboards.

## Common Workflows

### Adding a new GraphQL query or mutation

1. Edit the appropriate `gql/*.graphqls` file (or create a new one)
2. Run `make generate` in `backend/` — this creates resolver stubs
3. Implement the resolver logic in the generated `*.resolvers.go` file
4. If new SQL queries are needed, add them to `backend/internal/database/queries/*.sql`
5. Run `make generate` again to regenerate sqlc
6. Run `pnpm codegen` in `frontend/` to pick up schema changes

### Adding a new database query

1. Add the SQL to the appropriate file in `backend/internal/database/queries/`
2. Use named parameters: `@paramName::type` (not `$1`, `$2`)
3. Use `:many`, `:one`, `:exec`, or `:execresult` result annotations
4. Run `make generate` in `backend/`

### Adding a new entity

1. Create the migration SQL in `backend/internal/database/migrations/` (get approval first!)
2. Add the GraphQL types/inputs/queries/mutations in `gql/`
3. Add SQL queries in `backend/internal/database/queries/`
4. Run `make generate` in `backend/`
5. Add a ULID prefix and generator in `backend/internal/ulid/`
6. Implement resolvers, add dataloader if needed
7. Update `schema.sql` reference and `notes/` if applicable
