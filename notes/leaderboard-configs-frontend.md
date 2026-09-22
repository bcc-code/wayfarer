# Leaderboard Configs — Frontend

Frontend work for the persisted `LeaderboardConfig` entity. The backend side is
already merged and documented in [`leaderboard-configs.md`](./leaderboard-configs.md);
read that first — it explains the schema, the full-replace update input, and why
the ad-hoc `Project.leaderboard(...)` / `Event.leaderboard(...)` fields are *not*
deprecated.

This file tracks the frontend plan and its progress.

## Order of work

The two phases are deliberately sequential, not parallel:

1. **Admin interface** — CRUD for leaderboard configs in `layers/admin`.
   Nothing can be configured until this exists, so it comes first. It also
   settles the shape of the `.gql` documents and the filter editing UI that
   phase 2 reads back.
2. **User-facing app** — render standings from the configs in `layers/user`.
   Only meaningful once real projects have configs to serve, and the user side
   should not ship a fallback-heavy implementation guessing at data that admins
   cannot yet produce.

## Status

- [x] Phase 1 — admin interface
- [ ] Phase 2 — user-facing app

## What the API gives us

| Operation | Field |
| --- | --- |
| List (admin) | `leaderboardConfigs(filter: LeaderboardConfigFilter, first, after, last, before)` |
| Read one (admin) | `leaderboardConfig(id: ID!)` |
| Create | `createLeaderboardConfig(input: CreateLeaderboardConfigInput!)` |
| Update | `updateLeaderboardConfig(id: ID!, input: UpdateLeaderboardConfigInput!)` |
| Delete | `deleteLeaderboardConfig(id: ID!)` |
| Serve (user) | `Project.leaderboards` / `Event.leaderboards` → `[LeaderboardConfig!]!` |
| Computed board | `LeaderboardConfig.leaderboard(first, after, last, before)` → `LeaderboardConnection!` |

Two shapes to keep straight:

- `LeaderboardFilter` (**input**, on create/update) vs `LeaderboardFilterView`
  (**output**, on `LeaderboardConfig.filter`). Same fields, except `ageRange` is
  `AgeRangeInput` on the way in and `AgeRange` on the way out. An edit form has
  to map view → input when it loads.
- `UpdateLeaderboardConfigInput` is **full-replace**, unlike every other
  `Update*Input` in this codebase (which are PATCH-style with optional fields).
  `name`, `entityType`, `sortOrder`, `isActive` are all required; `filter: null`
  means "no filter". The edit form must always send complete state.

`Project.leaderboards` returns only **active** configs to normal users, and all
configs to admins/superadmins. So the admin list can use either that field or
the `leaderboardConfigs` query — the latter is paginated and filterable, and is
the one to use for the admin list.

## Phase 1 — admin interface — done

Follows the existing project-scoped CRUD pattern (closest analogue:
`achievements/`, which has index + `new` + `[id]` and drag-to-reorder).

Files:

- `layers/admin/app/pages/admin/projects/[projectId]/leaderboards/index.vue` —
  list, drag-to-reorder, event/inactive badges, filter summary per row
- `.../leaderboards/new.vue`, `.../leaderboards/[leaderboardId].vue` —
  create, edit + delete
- `layers/admin/app/components/admin/leaderboard/AdminLeaderboardConfigForm.vue`
  — the form both pages share
- `layers/admin/app/utils/leaderboardConfig.ts` — labels, the
  `LeaderboardFilterView` → `LeaderboardFilter` mapping, and the list's filter
  summary
- `app/graphql/fragments/leaderboardConfig.gql`,
  `app/graphql/mutations/leaderboards.gql` — page queries stay inline `gql()`
  per the existing convention
- `PROJECT_NAV` entry "Ledertavler" in `layers/admin/app/utils/adminNav.ts`
- Tests: `test/unit/leaderboardConfig.test.ts`,
  `test/component/AdminLeaderboardConfigForm.test.ts`; the `routes.test.ts`
  manifest snapshot picked up the three new routes

### What the build settled

- **The list reads `Project.leaderboards`, not `leaderboardConfigs`.** The
  paginated query orders by `created_at DESC`
  (`GetLeaderboardConfigsFilteredCursor`); only `Project.leaderboards` /
  `Event.leaderboards` order by `sort_order`, and they already return inactive
  configs to admins. No pagination — a project has a handful of boards.
- **`Project.leaderboards` includes the project's event-scoped configs.** The
  loader filters on `project_id` alone, so an event board comes back on the
  project too. Wanted here (the admin list shows every board in the project,
  with the event named in a badge) — but phase 2 must not assume the project's
  list is project-level only.
- **Reorder is N full-replace updates, not a bulk mutation.** There is no
  `reorderLeaderboardConfigs`; a drag sends one `updateLeaderboardConfig` per
  row whose index changed, each resending the filter it is not touching (hence
  `leaderboardFilterViewToInput`). Partial failure refetches to the server's
  order, same as the achievements page.
- **All nine filter fields get a control**, including ones an admin rarely
  touches. Not thoroughness for its own sake: the update input is full-replace,
  so a field with no control would be silently wiped the first time someone
  opened an existing config and saved it.
- **Event scope is create-only.** `UpdateLeaderboardConfigInput` has no
  `eventId`, so the edit form hides the picker rather than showing a control
  whose change cannot be saved.
- **"Unset" has two shapes, not one.** reka-ui rejects a `SelectItem` whose
  value is `''` — it reserves that value for "cleared" — so an "Alle" item is
  impossible and every optional picker uses `clear` instead. A cleared
  `USelectMenu` yields `null`; an emptied `UInput` yields `''`. The form treats
  any falsy value as unset rather than pinning one sentinel, which also keeps a
  legitimate `0` score bound from collapsing to "no bound".

## Phase 2 — user-facing app

Today `layers/user/app/pages/standings.vue` hardcodes three tabs — `global`,
`local`, `unit` — backed by `app/graphql/queries/pages/standings/{global,local,unit}.gql`
and the `StandingsGlobal` / `StandingsLocal` / `StandingsUnit` components. Each
builds its own ad-hoc `leaderboard(entityType:, filter:)` call, with the age
range and church filters computed client-side.

The config-driven version renders tabs from `myCurrentProject.leaderboards`
instead, with each tab's rows coming from that config's `leaderboard` field.

Points to decide in phase 2:

- `StandingsUnit` is not an ad-hoc leaderboard at all — it reads
  `myTeam.memberLeaderboard`. It does not map onto a `LeaderboardConfig` and
  probably stays as-is.
- `StandingsGlobal`'s age-range sub-tabs and `StandingsLocal`'s persons/units
  sub-tabs are client-side switches between two filters. As configs, those
  become two separate configs each — which changes the shape of the UI from
  "tab with inner tabs" to "flat list of tabs".
- Migration: projects without configs still need a working standings page, so
  either seed configs for existing projects or keep the current components as a
  fallback when `leaderboards` is empty.
- `myCurrentProject.leaderboards` will include the project's **event-scoped**
  configs as well (see phase 1's findings). Showing an event board as a project
  standings tab is probably wrong; filter on `event` being null, or decide
  deliberately to include them.
- `tab` is persisted in `localStorage` (`standings-tab`) and in the URL by the
  literal keys `global`/`local`/`unit`. Config-driven tabs need a stable key —
  the config `id` is stable but meaningless in a URL.
