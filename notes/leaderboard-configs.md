# Persisted Leaderboard Configs

## Why

Previously all leaderboards were entirely ad-hoc: clients called `Project.leaderboard(entityType, filter, ...)` / `Event.leaderboard(...)` and constructed the query at request time. There was no way for an admin to define a named, reusable leaderboard once and have it served as a finished result. `LeaderboardConfig` adds that: a persisted, project-scoped (optionally event-scoped) entity admins manage via CRUD, served to clients as a ready-made `LeaderboardConnection`.

The old ad-hoc `leaderboard(...)` fields on `Project`/`Event` are **deprecated but not removed** — existing frontend call sites keep working while they migrate.

## Database

- **Migration**: `00102_add_leaderboard_configs.sql`
- **Table**: `leaderboard_configs`
  - `project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE`
  - `event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL` — nullable, same optional-scoping pattern as `challenges.event_id`
  - `name`, `slug` (unique per project via `idx_leaderboard_configs_project_slug`)
  - `entity_type VARCHAR(20)` — mirrors `LeaderboardEntityType` (`PERSONS`/`TEAMS`/`SUPERTEAMS`/`CHURCHES`), enforced by a `CHECK` constraint
  - `filter JSONB` — nullable. Stores the same shape as the GraphQL `LeaderboardFilter` input, serialized via `json.Marshal`/`json.Unmarshal` (same "typed struct <-> JSONB `[]byte`" convention as `push_notification_log.target_criteria`, see `internal/services/push/service.go`). No typed columns per filter field — this avoids a migration every time `LeaderboardFilter` grows a field.
  - `sort_order INT`, `is_active BOOLEAN` (draft/inactive configs are hidden from non-admin callers)
- **ID prefix**: `LC` (`ulid.NewLeaderboardConfigID()` / `ulid.IsLeaderboardConfigID()`)

## GraphQL

- **Schema**: `gql/leaderboards.graphqls`
- **Type**: `LeaderboardConfig` — `project`/`event`/`leaderboard` are resolver fields (`@goField(forceResolver: true)`). `filter` is exposed as the `JSON` scalar (raw string on the wire, matching this codebase's existing `JSON` scalar convention — it binds to plain Go `string`, NOT a map, see `webhooks_helpers.go`/`settings.resolvers.go`/`quiz_helpers.go` for other `JSON`-scalar usages). The **input** side (`CreateLeaderboardConfigInput.filter` / `UpdateLeaderboardConfigInput.filter`) uses the typed `LeaderboardFilter` input instead — `LeaderboardFilter` can't be reused as an output field type since GraphQL forbids using an `input` type as an object field's type.
- **Admin CRUD**: `createLeaderboardConfig` / `updateLeaderboardConfig` / `deleteLeaderboardConfig` mutations, `leaderboardConfig(id)` / `leaderboardConfigs(filter, ...)` queries — all behind `@requireRole(roles: ["admin", "superadmin"])`. The two queries are a deliberate, confirmed exception to the "`@requireRole` is mutation-only" convention, since they exist purely to expose admin CRUD-management data (including inactive/draft configs).
  - `UpdateLeaderboardConfigInput.clearFilter: Boolean` — since a plain `COALESCE(sqlc.narg('filter'), filter)` update can't distinguish "not provided" from "explicitly clear," `clearFilter: true` forces `filter` to `NULL` (ignored if `filter` is also provided in the same call). This is the one field on this entity where clearing back to unfiltered is a realistic admin action; other `Update*Input`s in this codebase (e.g. `UpdateChallengeInput.imageUrl`/`.url`) have the same unaddressed limitation but weren't changed as part of this work.
- **Serving**: `Project.leaderboards` / `Event.leaderboards` return all *active* configs (all configs, including inactive, if the caller is admin/superadmin — checked via `RoleService.IsAdmin` in the resolver, not the schema). Each config's `leaderboard(first, after, last, before)` field returns the fully computed `LeaderboardConnection`.

## Reused leaderboard engine

`LeaderboardConfig.leaderboard` does **not** reimplement leaderboard computation — it adapts a config into the existing `services.LeaderboardParams` (see `backend/internal/graph/api/leaderboards.go:buildLeaderboardParamsFromConfig`) and calls the same `LeaderboardService.GetProjectLeaderboard`/`GetEventLeaderboard` used by the ad-hoc fields, then the same `buildLeaderboardConnection`/`FilterPersonLeaderboardEntries` helpers. This means:
- Caching, pagination (rank-based cursors), and `nearestChurchRivals` all work identically whether a leaderboard was reached via a config or the old ad-hoc query.
- The full-board cache key is derived from `(context, contextID, entityType, filterMap)` — not from how the request arrived — so a config and an ad-hoc query with matching project/entityType/filters correctly share one cache entry.

## Files touched

- Migration: `backend/internal/database/migrations/00102_add_leaderboard_configs.sql`
- ULID: `backend/internal/ulid/ulid.go`
- Schema: `gql/leaderboards.graphqls`, `gql/projects.graphqls`, `gql/events.graphqls` (deprecations)
- sqlc: `backend/internal/database/queries/leaderboard_configs.sql`
- Cache: `backend/internal/cache/{keys,invalidation,sync}.go` — `InvalidateLeaderboardConfig`, following the same per-entity pattern as `InvalidateChallenge`
- Dataloaders: `backend/internal/loaders/leaderboard_config_by_id.go`, `leaderboard_configs_by_project.go`, `leaderboard_configs_by_event.go`
- Resolvers/helpers: `backend/internal/graph/api/leaderboards.resolvers.go`, `leaderboards.go`
- Pagination: `backend/internal/graph/pagination/cursor.go` (`LeaderboardConfigCursor`), `connection.go` (`BuildLeaderboardConfigConnection`) — same per-entity cursor/connection pattern as `ChallengeCursor`/`BuildChallengeConnection`
