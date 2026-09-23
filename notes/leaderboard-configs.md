# Persisted Leaderboard Configs

## Why

Previously all leaderboards were entirely ad-hoc: clients called `Project.leaderboard(entityType, filter, ...)` / `Event.leaderboard(...)` and constructed the query at request time. There was no way for an admin to define a named, reusable leaderboard once and have it served as a finished result. `LeaderboardConfig` adds that: a persisted, project-scoped (optionally event-scoped) entity admins manage via CRUD, served to clients as a ready-made `LeaderboardConnection`.

The old ad-hoc `leaderboard(...)` fields on `Project`/`Event` are **not deprecated** — they're a different, still-valid way of getting at the same data (dynamic/ad-hoc queries vs. curated/admin-defined boards), not a predecessor `LeaderboardConfig` replaces. Both fields carry a plain doc comment cross-referencing the other, nothing more.

## Database

- **Migration**: `00102_add_leaderboard_configs.sql`
- **Table**: `leaderboard_configs`
  - `project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE`
  - `event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL` — nullable, same optional-scoping pattern as `challenges.event_id`
  - `name` — the only identifier besides `id`; there is deliberately no `slug` (it had no consumers — no `bySlug` lookup anywhere — and `id` already serves as the stable identifier for admin CRUD)
  - `entity_type VARCHAR(20)` — mirrors `LeaderboardEntityType` (`PERSONS`/`TEAMS`/`SUPERTEAMS`/`CHURCHES`), enforced by a `CHECK` constraint
  - `filter JSONB` — nullable. Stores the same shape as the GraphQL `LeaderboardFilter` input, serialized via `json.Marshal`/`json.Unmarshal` (same "typed struct <-> JSONB `[]byte`" convention as `push_notification_log.target_criteria`, see `internal/services/push/service.go`). No typed columns per filter field — this avoids a migration every time `LeaderboardFilter` grows a field.
  - `sort_order INT`, `is_active BOOLEAN` (draft/inactive configs are hidden from non-admin callers)
  - `created_at`/`updated_at TIMESTAMPTZ NOT NULL DEFAULT now()` — `NOT NULL`, matching `push_notifications`' stricter convention rather than the looser nullable-with-default convention used by most other tables in this codebase
- **Migration 00103** (`00103_drop_leaderboard_config_slug.sql`) exists because
  00102 was edited **after it had already been applied**. Code review dropped
  `slug` and its unique index from the `CREATE TABLE` and tightened
  `created_at`/`updated_at` to `NOT NULL`, but goose records 00102 as applied
  and never re-runs it. A database that migrated before that edit therefore
  still has `slug VARCHAR(100) NOT NULL` with no default, and every insert fails
  with `null value in column "slug" of relation "leaderboard_configs" violates
  not-null constraint (SQLSTATE 23502)` — which is exactly what the admin UI hit
  the first time it tried to create a config. 00103 drops the column and index
  conditionally, so it is a no-op on a database created from the amended 00102.
  Any environment that migrated between 27399ba2 and 729a7433 needs it.
- **ID prefix**: `LC` (`ulid.NewLeaderboardConfigID()` / `ulid.IsLeaderboardConfigID()`)
- No unique constraint besides the `id` primary key (the `(project_id, slug)` unique index was removed along with `slug`)

## GraphQL

- **Schema**: `gql/leaderboards.graphqls`
- **Type**: `LeaderboardConfig` — `project`/`event`/`leaderboard` are resolver fields (`@goField(forceResolver: true)`).
  - `filter: LeaderboardFilterView` — a dedicated output type mirroring `LeaderboardFilter`'s fields (using `AgeRange` instead of `AgeRangeInput` for the nested field), matching the `AgeRange`/`AgeRangeInput` paired-type pattern already used elsewhere in this schema. A GraphQL `input` type can't be reused as an output field's type, so this couldn't just be `filter: LeaderboardFilter`. The Go-side conversion lives in `leaderboards.go`: `filterViewToFilter` maps `*model.LeaderboardFilterView` back to `*model.LeaderboardFilter` for the leaderboard engine, and `ConvertRowToLeaderboardConfig` (in `loaders/leaderboard_config_by_id.go`) unmarshals the stored JSONB directly into `*model.LeaderboardFilterView` on read.
  - `Project.leaderboards` / `Event.leaderboards` are declared directly inside `type Project { ... }` / `type Event { ... }` (in `projects.graphqls`/`events.graphqls`), **not** via `extend type Project`/`extend type Event` from `leaderboards.graphqls`. This schema reserves `extend type` for `Query`/`Mutation` only — every other type's fields live in that type's own base declaration regardless of which domain "owns" them (see how `Challenge`/`Team`/`Achievement` fields sit inside `Project` in `projects.graphqls`).
- **Admin CRUD**: `createLeaderboardConfig` / `updateLeaderboardConfig` / `deleteLeaderboardConfig` mutations use `@requireRole(roles: ["admin", "superadmin"])`. The two read queries (`leaderboardConfig(id)` / `leaderboardConfigs(filter, ...)`) do **not** use `@requireRole` — that directive is reserved for mutations in this schema (per `backend/CLAUDE.md`); authorization for these two admin-only reads is enforced in the resolver instead, via the `requireAdminUser` helper in `leaderboards.go`.
- **`UpdateLeaderboardConfigInput` is full-replace**, not partial-patch: `name`, `entityType`, `sortOrder`, `isActive` are all required (`!`); the caller always resends the complete desired state. `filter` stays nullable — `filter: null` unambiguously means "no filter," `filter: {...}` means "this filter." There is no `clearFilter` flag and no absent-vs-explicit-null ambiguity to solve, because there's no "leave unchanged" case for any field except by resending its current value. This is a deliberate departure from every other `Update*Input` in this codebase (which are PATCH-style, all-optional with `COALESCE`-based partial updates) — chosen specifically to avoid the ambiguity a nullable `filter` field would otherwise create.
- **Serving**: `Project.leaderboards` / `Event.leaderboards` return all *active* configs (all configs, including inactive, if the caller is admin/superadmin — checked via `RoleService.IsAdmin` in the resolver, not the schema). Each config's `leaderboard(first, after, last, before)` field returns the fully computed `LeaderboardConnection`.

## Not yet supported: a per-config size limit

A config says *who* is on a board, not *how many* rows it shows. There is no
`max_entries` / "top N" anywhere — not in `leaderboard_configs`, not on
`LeaderboardConfig`, not in either input. The only size control is the `first`
pagination argument the **client** passes to `LeaderboardConfig.leaderboard`,
which `getLeaderboardForConfig` forwards straight to the engine.
(`filter.minScore` / `maxScore` are score bounds, not a rank cap.)

This is a real gap — deferred deliberately to keep the frontend PR small, to be
implemented by a backend developer later. Sketch:

1. Migration: `max_entries INT` on `leaderboard_configs`, nullable (null = no cap).
2. `gql/leaderboards.graphqls`: `maxEntries: Int` on `LeaderboardConfig` and on
   both `CreateLeaderboardConfigInput` and `UpdateLeaderboardConfigInput`
   (nullable in the update input, the same treatment `filter` gets in that
   full-replace input).
3. `leaderboard_configs.sql` + `make generate`.
4. `getLeaderboardForConfig`: clamp the effective `first` to `maxEntries` when
   set, so the cap holds whatever a client asks for.

Until then the client picks the size, which is why the user-facing standings
page hardcodes one `BOARD_SIZE` for every board — see
`leaderboard-configs-frontend.md`.

## Reused leaderboard engine

`LeaderboardConfig.leaderboard` does **not** reimplement leaderboard computation — it adapts a config into the existing `services.LeaderboardParams` (see `backend/internal/graph/api/leaderboards.go:buildLeaderboardParamsFromConfig`) and calls the same `LeaderboardService.GetProjectLeaderboard`/`GetEventLeaderboard` used by the ad-hoc fields, then the same `buildLeaderboardConnection`/`FilterPersonLeaderboardEntries` helpers (these take `rivalsContextID`/`rivalsIsEvent`/`rivalsFilter` and compute `nearestChurchRivals` lazily — see `feat/nearest-neighboor-ranking`'s "avoid computing rivals on every leaderboard query" refactor, which this branch rebased onto). This means:
- Caching, pagination (rank-based cursors), and `nearestChurchRivals` all work identically whether a leaderboard was reached via a config or the old ad-hoc query.
- The full-board cache key is derived from `(context, contextID, entityType, filterMap)` — not from how the request arrived — so a config and an ad-hoc query with matching project/entityType/filters correctly share one cache entry.

## Files touched

- Migration: `backend/internal/database/migrations/00102_add_leaderboard_configs.sql`
- ULID: `backend/internal/ulid/ulid.go`
- Schema: `gql/leaderboards.graphqls`, `gql/projects.graphqls`, `gql/events.graphqls`
- sqlc: `backend/internal/database/queries/leaderboard_configs.sql`
- Cache: `backend/internal/cache/{keys,invalidation,sync}.go` — `InvalidateLeaderboardConfig`, following the same per-entity pattern as `InvalidateChallenge`
- Dataloaders: `backend/internal/loaders/leaderboard_config_by_id.go`, `leaderboard_configs_by_project.go`, `leaderboard_configs_by_event.go`
- Resolvers/helpers: `backend/internal/graph/api/leaderboards.resolvers.go`, `leaderboards.go`
- Pagination: `backend/internal/graph/pagination/cursor.go` (`LeaderboardConfigCursor`), `connection.go` (`BuildLeaderboardConfigConnection`) — same per-entity cursor/connection pattern as `ChallengeCursor`/`BuildChallengeConnection`
