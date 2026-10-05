# Mine utmerkelser (achievements archive)

**Status:** Implemented (frontend only, no backend change)
**Started:** 2026-10-05
**Branch:** `feature/achievements-archive`

## Goal

Let a user see every achievement they have, grouped by the project it belongs
to — earned ones lit, unearned ones dimmed, exactly as the current project card
shows them. A record of what they have done across bible studies and camps.

Ships as a page at `/settings/archive`, titled **"Mine utmerkelser"** / "My
achievements", reached from a row in settings. No points/rank row and no action
buttons per project: just the project name and the badge grid.

## What already exists (findings)

The backend already has everything needed. **No schema or resolver changes are
required for the first version.**

| Piece                         | Where                                                                         | Note                                                                                                                  |
| ----------------------------- | ----------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `User.projects: [Project!]!`  | `gql/users.graphqls:19`                                                       | All projects the user has joined                                                                                      |
| Resolver                      | `backend/internal/graph/api/users.resolvers.go:367`                           | Dataloader-backed, translation-aware                                                                                  |
| Loader                        | `backend/internal/loaders/projects_by_user.go`                                | Batched over user IDs                                                                                                 |
| Query                         | `GetProjectsByUserIDs` in `backend/internal/database/queries/projects.sql:42` | `JOIN user_projects`, `ORDER BY up.user_id, p.start_date DESC` — **no archived filter**, returns every joined project |
| `Project.achievements`        | `backend/internal/graph/api/projects.resolvers.go:596`                        | Dataloader + cache, per project                                                                                       |
| Achievement SQL               | `GetAchievementsByProjectIDs` (`achievements.sql`)                            | Filters `a.hidden = false` — hidden achievements never reach the user layer                                           |
| `achievedAt` / `celebratedAt` | `helpers.go:104` / `:126`                                                     | Per _current user_, via `UserAchievementTimestampLoader`                                                              |
| `Project.archivedAt: Boolean` | `gql/projects.graphqls`                                                       | Note: typed `Boolean`, not a timestamp, despite the name. Maps to `projects.archived`                                 |

So `me { projects { id name archivedAt achievements { ... } } }` already returns
exactly the archive data, with per-user award state resolved correctly.

### Gotcha: `myProjects` is a lie

`Query.myProjects` (`projects.resolvers.go:831`) does **not** return the user's
projects. It returns a single-element slice containing the globally configured
_current_ project (`Settings.GetCurrentProjectID`). Do not use it for the
archive — use `me { projects }`.

## Performance notes

Per archived project the resolver chain is:

- `Project.achievements` → `AchievementsByProjectLoader` (batched + cached)
- per achievement: `achievedAt` and `celebratedAt` →
  `UserAchievementTimestampLoader` (batched by user+achievement)

At current scale (`achievements` ≈ 15 rows per project, few projects per user)
this is a handful of batched queries. If project count per user grows, consider
a dedicated "user achievement awards by project" query instead of per-achievement
timestamp loads.

The archive deliberately does **not** request `myPoints`, `leaderboard` or
`myTeam` for past projects — those are the expensive fields per the front-page
performance report. Being on its own route behind settings, it does not compete
with the home tab's first paint at all — the home tab's query is unchanged
apart from using the shared achievement fragment.

## Backend work (only if needed later)

If the archive ever needs points/rank per past project, or server-side
filtering, options are:

1. Add a `ProjectFilter`-style argument to `User.projects` (e.g. `archived`).
2. Add a dedicated `Query.myProjectArchive` returning a lean projection.

Both would follow the normal flow: edit `gql/*.graphqls` → `make generate` →
implement resolver → `pnpm codegen`. Per `CLAUDE.md`, do not put custom helpers
in `*.resolvers.go`.

## What was built

A page at **`/settings/archive`**, reached from a row in settings. The home tab
is untouched.

| File                                                                    | Change                                                                                          |
| ----------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| `frontend/app/graphql/fragments/achievement.gql`                        | **New.** `AchievementBadgeFields` — everything `AchievementBadge` + `AchievementDetails` render |
| `frontend/app/graphql/queries/pages/project-archive.gql`                | **New.** `ProjectArchive` query over `me { projects { ... } }`                                  |
| `frontend/app/graphql/queries/pages/profile.gql`                        | Achievement selection → the fragment                                                            |
| `frontend/layers/user/app/pages/settings/archive.vue`                   | **New.** The "Mine utmerkelser" page                                                            |
| `frontend/layers/user/app/pages/settings/index.vue`                     | New row in the existing link panel, below "Mine samtykke"                                       |
| `frontend/layers/user/app/components/achievements/AchievementBadge.vue` | Prop retyped `ProfilePageQuery[...]` → `AchievementBadgeFieldsFragment`                         |
| `frontend/layers/user/app/components/profile/ProfileProjectCard.vue`    | Same retype for its `achievements` prop                                                         |
| `frontend/i18n/locales/{nb,en_us}.json`                                 | `archive.myAchievements`, `archive.empty`                                                       |
| `frontend/test/component/ArchivePage.test.ts`                           | **New.** 9 tests                                                                                |
| `frontend/test/unit/__snapshots__/routes.test.ts.snap`                  | `settings-archive` added to the committed route manifest                                        |

No backend files touched, and nothing was added to the home tab's query.

### Naming and scope

Called **"Mine utmerkelser"**, not "Tidligere prosjekter" or "Dine
utmerkelser":

- **"Mine", not "Dine"** — the app is consistently first-person about the
  user's own things: `pages.consents` is "Mine samtykke" / "My consents",
  `navigation.profile` is "Min side" / "My page". Nothing in the app says
  "Dine".
- **Named after the achievements, not the projects.** The projects are only the
  grouping; the content is the badges. "Tidligere prosjekter" was honest but
  read like an account concept rather than something a user would open out of
  curiosity.
- **So the current project is included.** This is the part the name forces: an
  earlier revision filtered it out because its badges are already on the home
  tab, but then "Mine utmerkelser" would promise everything and quietly omit
  the camp the user is in right now. Overlapping with the home tab is the
  harmless kind — that grid is "right now", this page is the whole record.

The one thing the name does not capture is that the grid also shows
achievements the user has _not_ earned, dimmed. That is the same thing the home
card does, so it does not read as a false promise.

### `me.projects` is not "projects you have achievements in"

The first build of this assumed the current project would turn up in
`me { projects }`. It does not, reliably — and this was caught in testing, with
a user whose current-project achievements were missing from the page.

The two fields answer different questions:

- `myCurrentProject` resolves from `Settings.GetCurrentProjectID()`
  (`projects.resolvers.go:847`) — a **global** setting, no membership check.
  Everyone sees it on the home tab.
- `me { projects }` goes through `ProjectsByUserLoader` →
  `GetProjectsByUserIDs`, which **joins `user_projects`**.

A `user_projects` row is only written as a side effect of _doing_ something:
the explicit `joinProject` mutation, joining a team
(`teams.resolvers.go:55,403`), enrolling in a challenge
(`challenges.resolvers.go:558`), earning a content achievement
(`content_achievements.go:653`), or the ladder-to-heaven plugin. Awarding a
simple achievement does not. So holding achievements in a project is **not**
evidence of membership in it.

The page therefore selects `myCurrentProject` as well and merges the two lists
by id. `me.projects` was left alone — "projects you joined" is a correct
meaning for it, and bending it to mean something else would affect every other
caller.

### Why a page in settings

Two revisions got it here. First a collapsible section on the home tab,
lazy-loaded on expand; then a top-level `/archive` page linked from the home
tab; then the link moved into settings and the page with it.

Against the inline section:

- **It grows without bound.** One project per camp, ~15 achievements each.
  After three years that is 45–90 badges under the current-project card, so the
  home tab's scroll length ends up dominated by history.
- **The lazy-loading machinery only existed because it was inline** — pausing
  the query until expanded, deriving loading from `!data && !error` instead of
  `fetching`, a skeleton shaped to the collapsed layout. On a page the query
  just runs on mount.

Against a link on the home tab: looking back at old camps is a rare, deliberate
visit. The home tab is for the project you are in now, and a permanent row
pointing at history earns its space only if people use it — which is not the
expectation. Settings is already where the "about me, occasionally" rows live
(consents, add to home screen), and the archive sits in that panel without
needing to justify itself.

Living under `/settings/` also means it inherits the bottom-nav hiding from the
`/settings` prefix in `layouts/default.vue`, with no special case, and closes
back to settings exactly like `add-to-home.vue` and `consent.vue` do.

### Details worth keeping in mind

- **The settings row is unconditional.** An earlier version hid a home-tab row
  when the user had no history, which cost `me { projects { id } }` on the home
  tab's critical path. In settings a static row is the norm — "Mine samtykke"
  shows whether or not you have any — so the row is always there and the page
  carries an empty state instead. The home tab pays nothing.
- **Projects with no achievements are kept**, deliberately — participation is
  worth showing even with nothing earned. Covered by a test.
- **No card per project.** Each is a centered name over its badge grid directly
  on the page background. An earlier version wrapped each in a `DesignCard`,
  which made achievement-less projects render as empty bars.
- **Sorting is client-side.** The backend already orders `start_date DESC`, but
  the page re-sorts rather than trusting it silently; the order is asserted in a
  test. Uses `[...].sort()`, not `toSorted` — the latter is missing on iOS
  Safari < 16.4.
- **No entrance animation.** A staggered reveal was built while the archive was
  an inline collapsible, where the sections appeared in place and the motion
  explained the expansion. On a page the list _is_ the content, so it was
  removed rather than carried over.
- **Loading skeleton mirrors the real layout** — two placeholder project groups,
  each a centered name bar over an 8-badge round grid — so the page does not
  reflow when data lands.
- **`AchievementBadge` is stubbed in the page test.** It owns a teleporting
  `DesignDrawer` and a celebration mutation; the page's job is only to hand it
  the right achievements, so the test asserts on the props it receives.
- **Celebration is unchanged.** `index.vue` still passes only
  `myCurrentProject.achievements` to `useAchievementCelebration`, so no confetti
  fires for old projects.
- **Only `nb` and `en_us` have the new strings.** Every other maintained locale
  falls back to `nb` (the `defaultLocale`; no `fallbackLocale` is configured).
  The other 18 locale files are kept in sync at 23 top-level keys, so these two
  keys need to go through the normal translation process.
- **`test/unit/routes.test.ts` holds a committed route-manifest snapshot.**
  Adding a page fails it until the snapshot is updated — that is the guard
  working, not a flake.

## Open questions

- **Archived-achievement celebration.** Opening a past project's unearned →
  earned achievement from the archive and closing the drawer fires
  `markAchievementCelebrated` on old data. Harmless as far as the schema goes,
  but it is a silent write; confirm it is wanted, or gate it on the achievement
  belonging to the current project.
- **Per-project detail.** The page shows name + badges only. Points, rank, team
  and the project's dates are all available on `Project` and would make the
  archive more of an archive — but each is an extra per-project resolver, so
  measure before adding.

## Checklist

- [x] Decide: which projects → **all of them, current one included**
- [x] Decide: the name → **"Mine utmerkelser" / "My achievements"**
- [x] Decide: inline section vs. page → **page at `/settings/archive`**
- [x] Extract shared achievement fragment; retype `AchievementBadge`
- [x] Add `ProjectArchive` query + `pnpm codegen`
- [x] Build `settings/archive.vue`
- [x] Link it from the settings link panel
- [x] Loading skeleton mirroring the real layout
- [x] i18n strings (`nb`, `en_us`)
- [x] Component tests — 9, all passing
- [x] Route manifest snapshot updated
- [x] `pnpm lint` (0 errors), `pnpm typecheck` (clean), `pnpm test` (416
      component + 738 unit, all passing)
- [ ] Confirm celebration behaviour for archived achievements
- [ ] Translate `archive.*` into the remaining locales
- [ ] Verify against real data (needs a user with more than one project)
