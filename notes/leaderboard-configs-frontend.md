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
- [x] Phase 2 — user-facing app

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

## Phase 2 — user-facing app — done

Today `layers/user/app/pages/standings.vue` hardcodes three tabs — `global`,
`local`, `unit` — backed by `app/graphql/queries/pages/standings/{global,local,unit}.gql`
and the `StandingsGlobal` / `StandingsLocal` / `StandingsUnit` components. Each
builds its own ad-hoc `leaderboard(entityType:, filter:)` call, with the age
range and church filters computed client-side.

### Decided

- **Config-driven only, no fallback.** A project with no configs shows an empty
  state; the old `StandingsGlobal` / `StandingsLocal` components and their
  `.gql` documents are deleted rather than kept as a fallback path. Existing
  projects need configs created in the admin UI before this ships — that is the
  accepted cost of not carrying two code paths.
- **Flat list of tabs**, one per config, in `sortOrder`. The current nested
  sub-tabs (`StandingsGlobal`'s age ranges, `StandingsLocal`'s persons/units)
  disappear: each becomes its own config and therefore its own top-level tab.
- **The unit tab stays as-is, always last.** `StandingsUnit` reads
  `myTeam.memberLeaderboard`, not an ad-hoc leaderboard, so it does not map onto
  a `LeaderboardConfig`. It keeps its own query and stays gated on the user
  having a team.

### The constraint that shapes this

`leaderboardConfig(id)` and `leaderboardConfigs(...)` are **admin-only** — the
resolvers call `IsAdmin` and return "permission denied" otherwise (see
`leaderboards.resolvers.go`). A normal user can only reach configs through
`myCurrentProject.leaderboards`, which takes no arguments.

So there is no way to fetch *one* config's board by id as a user. The page
fetches the whole list with each config's `leaderboard` in a single query and
switches tabs client-side. Consequences to keep in mind:

- Every board is computed on page load, not just the open tab. That is close to
  what the page already does — `StandingsLocal` computes two boards in one
  query — and the backend caches full boards keyed by
  `(context, contextID, entityType, filter)`, shared with the ad-hoc path.
- One `first` applies to **all** boards in that query; it cannot vary per
  config the way the old code used 20 for persons and 500 for teams. Going with
  `first: 100`.
  - The real fix is a per-config size limit, which the backend does not have —
    an admin can define who is on a board but not how many rows it shows. See
    "Not yet supported: a per-config size limit" in
    [`leaderboard-configs.md`](./leaderboard-configs.md). Deferred deliberately
    to keep this PR small; a backend developer will add it, and `BOARD_SIZE`
    then goes away.
- Tab switching costs no request, which is a UX gain over the current
  `v-if`-per-tab components.

### What was built

- `layers/user/app/pages/standings.vue` — one `StandingsPage` query fetching
  every config and its board; tabs are the configs in `sortOrder`, with `unit`
  appended last when the user has a team. The tab bar is hidden when there is
  only one tab.
- `layers/user/app/components/standings/StandingsBoard.vue` — presentational:
  one config's name, its entries, and the viewer's own row via `getExtraItems`.
- Deleted: `StandingsGlobal.vue`, `StandingsLocal.vue`,
  `app/graphql/queries/pages/standings/{global,local}.gql`, and
  `test/component/Standings{Global,Local}.test.ts`.
- Tests: `test/component/StandingsPage.test.ts`,
  `test/component/StandingsBoard.test.ts`.
- `AGE_RANGE_YOUNG` / `AGE_RANGE_ADULT` kept — `pages/index.vue` still uses
  them.

### Navigation

`layers/user/app/layouts/default.vue` hid the standings tab when the persons
leaderboard had no rows. It now hides it when the project has **no leaderboards
configured** — `myCurrentProject.leaderboards` is empty. The old check asked the
wrong question once the page became config-driven: a configured board that is
still empty is a page worth opening, and rows without a config are not reachable
at all.

One consequence to know: `leaderboards` returns inactive configs to
admins/superadmins, so an admin sees the tab in a project whose every board is
inactive. That matches what they see on the page itself.

### Resolved while building

- **Tab identity.** The config `id` is what goes in `?tab=` and the
  `standings-tab` localStorage key. An unknown value — a deleted or deactivated
  config, a different project, or the old `global` / `local` literal left over
  from before this change — falls back to the first tab. Note that
  `useLocalStorage` with a string default uses the raw String serializer, not
  JSON; the value is stored unquoted.
- **Event-scoped configs are shown.** `myCurrentProject.leaderboards` returns
  them (the loader filters on `project_id` alone) and the user app has no event
  pages, so filtering them out would make an event board unreachable. An admin
  created and named it deliberately, and `isActive` already controls whether
  anyone sees it.
- **Tab labels come from config names**, which are admin-authored and
  untranslated. `standings.unit` stays for the unit tab; `standings.global`,
  `local`, `top`, `u18`, `o18` and `units` were removed from all 16 locale files
  that carried them.
- **Analytics.** `LeaderboardTabChanged` sends tab *labels*, not ids — a config
  id says nothing in a dashboard, where the old values were readable
  (`global` / `local` / `unit`). `TeamLeaderboardViewed` still fires on the unit
  tab.

### Before this ships

Existing projects have no configs, so their standings tab disappears and the
page shows an empty state. Configs have to be created in the admin UI (or
seeded) for every live project first. This is the accepted cost of the
"config-driven only, no fallback" decision above — it is not a bug to discover
later.
