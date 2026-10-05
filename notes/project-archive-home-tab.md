# Project Archive on the Home Tab

**Status:** Implemented (frontend only, no backend change)
**Started:** 2026-10-05
**Branch:** `feature/achievements-archive`

## Goal

Show the user's *earlier* projects below the current project card on the home
tab — project name as a section heading, with that project's achievement
badges underneath (earned ones lit, unearned ones dimmed, exactly like the
current project card). An "archive" of what the user has done in past bible
studies / camps.

Per the design: the archive sits under the current `ProfileProjectCard` and the
feedback button, repeating `name + achievement grid` per past project. There is
no points/rank row and no action buttons in the archive sections — just the
heading and the badge grid.

## What already exists (findings)

The backend already has everything needed. **No schema or resolver changes are
required for the first version.**

| Piece | Where | Note |
| --- | --- | --- |
| `User.projects: [Project!]!` | `gql/users.graphqls:19` | All projects the user has joined |
| Resolver | `backend/internal/graph/api/users.resolvers.go:367` | Dataloader-backed, translation-aware |
| Loader | `backend/internal/loaders/projects_by_user.go` | Batched over user IDs |
| Query | `GetProjectsByUserIDs` in `backend/internal/database/queries/projects.sql:42` | `JOIN user_projects`, `ORDER BY up.user_id, p.start_date DESC` — **no archived filter**, returns every joined project |
| `Project.achievements` | `backend/internal/graph/api/projects.resolvers.go:596` | Dataloader + cache, per project |
| Achievement SQL | `GetAchievementsByProjectIDs` (`achievements.sql`) | Filters `a.hidden = false` — hidden achievements never reach the user layer |
| `achievedAt` / `celebratedAt` | `helpers.go:104` / `:126` | Per *current user*, via `UserAchievementTimestampLoader` |
| `Project.archivedAt: Boolean` | `gql/projects.graphqls` | Note: typed `Boolean`, not a timestamp, despite the name. Maps to `projects.archived` |

So `me { projects { id name archivedAt achievements { ... } } }` already returns
exactly the archive data, with per-user award state resolved correctly.

### Gotcha: `myProjects` is a lie

`Query.myProjects` (`projects.resolvers.go:831`) does **not** return the user's
projects. It returns a single-element slice containing the globally configured
*current* project (`Settings.GetCurrentProjectID`). Do not use it for the
archive — use `me { projects }`.

## Plan

### 1. Frontend data

Add a separate query rather than extending `ProfilePage`, and **pause it until
the archive is actually opened**.

`frontend/app/graphql/queries/pages/profile.gql` is the home tab's critical
path, and `notes/front-page-performance-report.md` documents how sensitive it
is. Keeping the archive in its own operation means:

- the project card + leaderboard still paint at the same speed
- a slow or failing archive can't blank the home tab
- **the archive query only runs when the user expands the section** — see the
  collapsible below

Pause it the same way pages pause on auth, by `&&`-ing in the expanded state:

```ts
const expanded = ref(false)
const { data, fetching, error } = useProjectArchiveQuery({
  pause: computed(() => !isAuthReady.value || !expanded.value),
})
```

urql keeps the result cached once fetched, so collapsing and re-expanding does
not refetch.

```graphql
query ProjectArchive {
  me {
    id
    projects {
      id
      name
      startDate
      archivedAt
      achievements {
        ...ArchiveAchievementFields
      }
    }
  }
}
```

Then filter out the current project client-side (compare against
`myCurrentProject.id` from `ProfilePage`).

**Decided:** the archive shows *all* projects that are not the current one. It
does **not** filter on `archivedAt` — that flag is admin-controlled and is not
reliably set on old projects, so keying off it would silently drop history.

`archivedAt` is still selected in the query; it may be useful later for
labelling, but it must not gate what is shown.

### 2. Share an achievement fragment

`AchievementBadge.vue` currently types its prop as
`ProfilePageQuery['myCurrentProject']['achievements'][number]`, which hard-wires
it to one query. Extract the achievement selection into
`frontend/app/graphql/fragments/achievement.gql` and retype the badge against
the generated fragment type, so both `ProfilePage` and `ProjectArchive` feed the
same component.

`AchievementDetails.vue` already takes a structural prop type, so it needs no
change.

Fields the badge + details need: `__typename id name descriptionPending
descriptionCompleted imagePendingObject imageCompletedObject hidden achievedAt
celebratedAt points`, plus `totalItems`/`completedItemCount` on
`ContentAchievement` and `StreakAchievement`.

### 3. New component

`layers/user/app/components/profile/ProfileProjectArchive.vue` — a collapsible
"Earlier projects" section. Collapsed by default; expanding it triggers the
query (see above). Inside, one block per past project: the project name as a
heading, then that project's `AchievementBadge` grid, reusing the
`grid-cols-4 gap-medium` layout from `ProfileProjectCard`.

The component owns the `expanded` state and the query, so `index.vue` just
mounts it — inside the existing `TransitionGroup`, after the `UserFeedback`
block.

**There is no collapsible in the design system.** `layers/user/app/components/design/`
has Button, Card, Drawer, IconButton, Image, Input, Panel, Skeleton, Slider,
Switch, Tabs, Textarea — no accordion/disclosure. Built inline: a `<button>`
with `aria-expanded` plus `IconChevronRight` rotated 90° when open. Promote it
to a `DesignCollapsible` if a second screen wants one.

States to handle inside the expanded section: fetching (skeleton), error
(`<ErrorState>`), and empty (no past projects). When the archive is empty the
whole section should hide itself rather than expand onto nothing — but that is
only knowable *after* the query runs, which only happens after expanding. Two
options, pick one at implementation time:

1. Accept it: show an empty-state line inside the expanded section.
2. Have `ProfilePage` return a cheap count of the user's projects so the
   section can hide before any archive data is loaded. Costs a backend change.

Went with (1).

Sort newest-first. The backend already orders by `start_date DESC`, but don't
rely on it silently — assert the order in the component test.

### 4. Celebration must not fire for archived projects

`index.vue` calls `useAchievementCelebration(achievements)` with only
`myCurrentProject.achievements`. Keep it that way — an old uncelebrated
achievement from a past project must not pop confetti on the home tab. Verify
that `AchievementBadge`'s own `markCelebrated` call on drawer close is
acceptable for archived achievements (it will fire; probably harmless, but it is
a silent mutation on old data — confirm this is wanted).

### 5. i18n

Needed: the collapsible's label ("Earlier projects" / nb: "Tidligere
prosjekter") and an empty-state line. Locales live in
`frontend/i18n/locales/*.json`, default `nb`.

### 6. Tests

- Component test for `ProfileProjectArchive` under `frontend/test/component/`
  (`// @vitest-environment nuxt`, `mountSuspended`, mock the generated query
  composable) — covering: collapsed by default, query paused while collapsed,
  loading, error, empty (no past projects), one project, many projects, current
  project excluded, newest-first order.
  Asserting "paused while collapsed" means checking the `pause` option the
  mocked query composable was called with — it is a `ComputedRef`, so read
  `.value` after toggling.
- No backend tests needed if no backend change lands.

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
performance report.

## Backend work (only if needed later)

If the archive ever needs points/rank per past project, or server-side
filtering, options are:

1. Add a `ProjectFilter`-style argument to `User.projects` (e.g. `archived`).
2. Add a dedicated `Query.myProjectArchive` returning a lean projection.

Both would follow the normal flow: edit `gql/*.graphqls` → `make generate` →
implement resolver → `pnpm codegen`. Per `CLAUDE.md`, do not put custom helpers
in `*.resolvers.go`.

## What was built

| File | Change |
| --- | --- |
| `frontend/app/graphql/fragments/achievement.gql` | **New.** `AchievementBadgeFields` — everything `AchievementBadge` + `AchievementDetails` render |
| `frontend/app/graphql/queries/pages/project-archive.gql` | **New.** `ProjectArchive` query over `me { projects { ... } }` |
| `frontend/app/graphql/queries/pages/profile.gql` | Inlined achievement selection replaced by the fragment |
| `frontend/layers/user/app/components/profile/ProfileProjectArchive.vue` | **New.** The collapsible section; owns the paused query |
| `frontend/layers/user/app/components/achievements/AchievementBadge.vue` | Prop retyped `ProfilePageQuery[...]` → `AchievementBadgeFieldsFragment` |
| `frontend/layers/user/app/components/profile/ProfileProjectCard.vue` | Same retype for its `achievements` prop |
| `frontend/layers/user/app/pages/index.vue` | Mounts `ProfileProjectArchive` after `UserFeedback` |
| `frontend/i18n/locales/{nb,en_us}.json` | `archive.earlierProjects`, `archive.empty` |
| `frontend/test/component/ProfileProjectArchive.test.ts` | **New.** 9 tests |

No backend files touched.

### Notes for review

- **Sorting is client-side.** The backend already orders `start_date DESC`, but
  the component re-sorts rather than trusting it silently; the order is asserted
  in a test. Uses `[...].sort()`, not `toSorted` — the latter is missing on iOS
  Safari < 16.4.
- **`AchievementBadge` is stubbed in the archive test.** It owns a teleporting
  `DesignDrawer` and a celebration mutation; the archive's job is only to hand
  it the right achievements, so the test asserts on the props it receives.
- **Celebration is unchanged.** `index.vue` still passes only
  `myCurrentProject.achievements` to `useAchievementCelebration`, so no confetti
  fires for old projects. The badge's own "mark celebrated on drawer close" does
  still fire for an archived achievement opened from the archive — see the open
  question below.
- **Only `nb` and `en_us` have the new strings.** Every other maintained locale
  falls back to `nb` (the `defaultLocale`; no `fallbackLocale` is configured).
  The other 18 locale files are kept in sync at 23 top-level keys, so these two
  keys need to go through the normal translation process.

## Open questions

- **Archived-achievement celebration.** Opening a past project's unearned →
  earned achievement from the archive and closing the drawer fires
  `markAchievementCelebrated` on old data. Harmless as far as the schema goes,
  but it is a silent write; confirm it is wanted, or gate it on the achievement
  belonging to the current project.
- **Empty archive still expands onto a message** rather than hiding the section
  — unavoidable without a cheap project count on `ProfilePage` (see above).

## Checklist

- [x] Decide: all non-current projects, or `archivedAt == true` only
      → **all non-current projects**
- [x] Decide: eager or lazy → **lazy, paused behind a collapsible**
- [x] Extract shared achievement fragment; retype `AchievementBadge`
- [x] Add `ProjectArchive` query + `pnpm codegen`
- [x] Build `ProfileProjectArchive.vue` (collapsible, owns its own query)
- [x] Mount in `layers/user/app/pages/index.vue`
- [x] i18n strings (`nb`, `en_us`)
- [x] Component tests — 9, all passing
- [x] `pnpm lint` (0 errors), `pnpm typecheck` (clean), `pnpm test` (415
      component + 738 unit, all passing)
- [ ] Confirm celebration behaviour for archived achievements
- [ ] Translate `archive.*` into the remaining locales
- [ ] Verify against real data (needs a user with more than one project)
