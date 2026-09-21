# Admin UX improvements — working note

Per-page tracker for the admin UX pass. The **structural** work (layers,
project-scoped IA, tabs → routes, shared shell, breadcrumbs, permission
consolidation) is finished and written up in
[`frontend-admin-restructure.md`](./frontend-admin-restructure.md) — read the
"Watch out for" section there before touching any project-scoped page. This note
is about what happens _inside_ the pages.

**49 `.vue` files under `layers/admin/app/pages/admin/`**, of which 44 are real
pages: one is the `[projectId].vue` parent shell and four are legacy redirect
stubs.

## How to use this

- One row per page. Fill in `Notes` as findings accumulate; tick `St.` when done.
- Status: `☐` untouched · `◐` in progress · `☑` done · `–` n/a (stub/shell)
- Cross-cutting items are listed once at the top rather than repeated per row —
  fix them in one sweep, not page by page. Per-page notes reference them by
  number (`see #1`).
- Admin is **desktop-only and Norwegian-only** — see Scope decisions. Do not
  file mobile-layout or i18n findings.
- New components follow the self-contained styling rule in Conventions.
- Update the log at the bottom as work lands, same convention as the
  restructure note.

---

## Scope decisions

Settled, so they do not get re-raised. Both of these remove work the inventory
had flagged.

**No i18n in the admin panel.** Everyone who works in admin is Norwegian, so
hardcoded Norwegian strings are the intended end state, not debt. The 37
untranslated pages need nothing. This also means the seven translated
`my-church/**` pages are the anomaly rather than the model — leave their `t()`
calls alone (they work), but do not extend the pattern to new admin pages, and
do not add keys to `adminNav.ts`.

One real bug survives this decision and is _not_ an i18n task:
`layouts/admin.vue` calls `setLocale('nb')` unconditionally on mount, and
`@nuxtjs/i18n` persists it in a cookie with nothing restoring it — so a German
user who opens `/admin` once has the **entire consumer app** switched to
Norwegian permanently. That is a consumer-app bug caused by admin code, and it
should be fixed on its own terms. `church-admin.vue` deliberately does not force
the locale, which is why the my-church area is unaffected.

**Admin is desktop-only.** No mobile or tablet work. The nine `UTable` pages
with 4-7 columns and no `overflow-x` wrapper are therefore fine as they are, and
narrow-viewport layout is out of scope everywhere in this note. (The dashboard
shell still gets mobile nav for free from `UDashboardNavbar`; nothing needs to
be removed, it just is not a target.)

**This does not excuse `sm:`/`lg:` grid breakpoints inside the panel** — see #6.
That is a _desktop_ bug, not a mobile one: the two sidebars take ~300-600px out
of the window before the panel gets any, so a viewport breakpoint fires at the
wrong width on exactly the wide screens admin is used on.

---

## Conventions

House rules for this pass. New work follows them; existing pages are brought
along as they are touched, not in a separate sweep.

### Components own their own responsive styling

**A component should look right at any width it is given, without the page or
layout knowing anything about it.** The parent decides _where_ a component goes
and how much room it gets; the component decides what to do with that room. A
layout that has to reach in and restyle its children — or a component that only
works in one particular slot — is the thing this rule exists to prevent.

In practice:

- The component puts `@container` on **its own root element**, then uses the
  container variants (`@sm:`, `@md:`, `@4xl:`) inside. It does not depend on an
  ancestor having declared a containment context, so it is genuinely
  parent-agnostic and can be dropped in a sidebar, a half-width grid cell or a
  full-bleed panel unchanged.
- Viewport variants (`sm:`, `md:`, `lg:`) are **wrong inside an admin component**
  by default. They measure the window, and the window is not what the component
  was given: the two sidebars take ~300-600px first. Reserve them for the layouts
  themselves, which legitimately do care about the window.
- Pages stop carrying child-specific layout classes. If a card needs to be
  narrower at some width, that belongs in the card.

**Two mechanics that make this bite if you skip them:**

1. **`@lg:` is not `lg:`.** Container variants resolve against the
   `--container-*` scale, viewport variants against `--breakpoint-*`, and the two
   scales differ by roughly a factor of two (verified in
   `node_modules/tailwindcss/theme.css`, Tailwind 4.3.2):

   | Name  | Viewport variant | Container variant |
   | ----- | ---------------- | ----------------- |
   | `sm`  | `sm:` 40rem      | `@sm:` 24rem      |
   | `md`  | `md:` 48rem      | `@md:` 28rem      |
   | `lg`  | `lg:` 64rem      | `@lg:` 32rem      |
   | `xl`  | `xl:` 80rem      | `@xl:` 36rem      |
   | `4xl` | —                | `@4xl:` 56rem     |

   So a mechanical `lg:` → `@lg:` rename **halves the trigger width** and the
   component reflows far too early. Pick the container value by what the
   component actually needs, which is why the earlier page conversions landed on
   `@md:`/`@4xl:` rather than a like-for-like swap.

2. **`container-type: inline-size` changes how the element sizes itself.** The
   `@container` class applies inline-axis containment, so the element takes its
   width from its parent instead of shrinking to fit its contents. Putting it on
   something that relied on shrink-to-fit (an inline-block chip, a
   content-width button row) will stretch it. Wrap instead of converting in
   place when that happens.

   Related: nested containers resolve to the **nearest** ancestor container, so
   a component with `@container` inside another one shadows the outer for
   unnamed queries. That is usually what you want; when it is not, name them
   (`@container/card` + `@md/card:`).

---

## Cross-cutting work

Ordered by priority. These are sweeps — do them once across every affected page
rather than page by page.

### 1. Replace hardcoded query limits with pagination — 12 pages

Every list below renders at most N rows, with no pagination and no indication
that anything was cut off. `projects/[projectId]/index.vue`'s old mega-query had
exactly this bug and lost its `first: 50` during the restructure; these are the
remainder.

| Page                                                | Limit                   | Shape                          |
| --------------------------------------------------- | ----------------------- | ------------------------------ |
| `my-church/units.vue`                               | `first: 1000`           | list, also an outlier (see #4) |
| `my-church/admins.vue`                              | `first: 500`            | list + client fuzzy search     |
| `projects/[projectId]/superteams/[superTeamId].vue` | `first: 200` (teams)    | picker                         |
| `users/[userId]/achievements.vue`                   | `first: 200`            | picker                         |
| `projects/index.vue`                                | `first: 100`            | card grid                      |
| `users/[userId]/index.vue`                          | `first: 100` (feedback) | panel in a detail page         |
| `projects/[projectId]/challenges/new.vue`           | `first: 100` (events)   | dropdown                       |
| `projects/[projectId]/superteams/distribute.vue`    | `first: 100` (events)   | dropdown                       |
| `projects/[projectId]/challenges/index.vue`         | `first: 50`             | list                           |
| `projects/[projectId]/achievements/index.vue`       | `first: 50`             | list, drag-reorder             |
| `projects/[projectId]/events/index.vue`             | `first: 50`             | list                           |
| `projects/[projectId]/superteams/index.vue`         | `first: 50`             | list                           |
| `maintenance/check-points-journal.vue`              | `first: 50`             | table                          |
| `maintenance/fix-content-progress.vue`              | `first: 50`             | table                          |

`RelayPagination` + `usePagination` already exist and are proven on four pages:
`users/index.vue`, `feedback/index.vue` and
`projects/[projectId]/{teams,scores}/index.vue`. `users/index.vue` is the
reference — debounced server-side filter plus relay cursors.

**Not every row here wants the same fix.** Three shapes, three answers:

- **Lists** — straight `RelayPagination`, following `users/index.vue`.
- **Dropdowns and pickers** (`challenges/new.vue`, `distribute.vue`,
  `users/[userId]/achievements.vue`, `superteams/[superTeamId].vue`) — a paginated
  dropdown is worse UX, not better. These want a searchable select backed by a
  server-side query, or a justified limit with an explicit "showing first N"
  affordance.
- **`achievements/index.vue` has drag-to-reorder**, which does not compose with
  pagination — reordering across a page boundary has no meaning. Decide the
  interaction before the query: either the order is page-local, or reordering
  moves to a dedicated mode that loads everything.

### 2. Loading, empty and error states everywhere they belong

Make the three states universal and consistent. `AdminErrorState` is on 30
pages, `AdminLoadingState`/`USkeleton` on 22; empty states are ad hoc.

Decide one house pattern per page shape first, then apply it — auditing 44 pages
individually is the expensive way to do this:

| Shape            | Loading                                    | Empty                                          | Error                                        |
| ---------------- | ------------------------------------------ | ---------------------------------------------- | -------------------------------------------- |
| List / table     | skeleton rows (not just `UTable :loading`) | "nothing here yet" + the primary create action | `AdminErrorState` + retry                    |
| Detail           | skeleton matching the layout               | n/a                                            | `AdminErrorState`; 404 distinct from failure |
| Form (`new.vue`) | only if it loads options                   | n/a                                            | inline field errors + submit failure         |

Known gaps: `consents/index.vue`, `feedback/index.vue`,
`projects/[projectId]/{scores,teams}/index.vue` and all four maintenance tools
have an error state but **no loading state** — they lean on `UTable :loading`,
which shows an empty table rather than a loading one. `teams/index.vue` and
`projects/index.vue` have no empty state at all.

Do not "deduplicate" `AdminErrorState`/`AdminLoadingState` back into the shared
`ErrorState`/`LoadingState`: keeping them separate is what severs admin's
dependency on the user design system (see the restructure note).

### 3. Search, filter and sort on lists

Only three of roughly fifteen lists can be narrowed at all: `users/index.vue`
(debounced server-side `filter.query`), `feedback/index.vue` (three
`USelectMenu` facets) and `my-church/admins.vue` (client-side fuzzy).

Nothing for challenges, achievements, events, superteams, teams, scores,
consents or the maintenance tables. No list has column sorting either — all nine
`UTable` pages pass a static `:columns`.

Pairs naturally with #1: a server-side filter and relay pagination are the same
query change, and doing them together avoids touching each page twice.

### 4. Refactor the two outliers

`my-church/units.vue` (1,109 lines) and `users/[userId]/index.vue` (1,063).
Both are worth breaking up — extract the panels into components, lift the
queries out of the template. They are the hardest pages to change safely, so
this is a prerequisite for the other sweeps landing on them rather than a
separate nicety. `units.vue` also carries the worst limit in the codebase
(`first: 1000`), so #1 and #4 meet here.

### 5. Give `church-admin` a real layout

The seven `my-church/**` pages run on a header-plus-`<slot />` layout: no
sidebar, no navigation, and breadcrumbs only because they were restored by hand
after the shell migration. It is the one admin surface with no way to get
between its own pages.

Deferred through the whole restructure; now in scope. Two constraints carried
over from that note:

- **Do not route these pages through the `admin` shell as-is.** The fix is
  navigation for `church-admin`, not a merge — a church admin should not see the
  global sidebar, and `isChurchAdminOnly` exists to keep them out of it.
- **Keep `setLocale` out of any shared shell composable or nav component.** The
  my-church area is unaffected by the locale bug precisely because
  `church-admin.vue` does not force the locale; a shared shell that does would
  regress it.

Because admin is desktop-only and Norwegian-only, this is a straight navigation
problem: nav model, active state, and a landmark that says which church you are
administering.

### 6. Bring existing components onto the self-contained rule

The convention above is the target; this is the backlog of what does not meet it
yet. Small, and smaller than expected.

The restructure note's container-query sweep and this note's inventory both
scanned **`pages/` only**, so `components/` had never been checked. Scanned
2026-09-21 — **exactly one hit** across all 35 admin components:

```
admin/dashboard/AdminDashboardStats.vue:26  grid grid-cols-2 gap-4 lg:grid-cols-4
```

That is the same line as home-dashboard bug 2, so it gets fixed there rather
than as its own task. The component surface is otherwise clean.

The item stays recorded because **the audit was the gap, not the code** — nothing
detects a viewport breakpoint in a new component, and the grep needs anchoring on
`(^|[" ])` rather than `\b`, since `@` is not a word character and a `\b` match
also hits the `@md:` variants it is meant to distinguish from:

```
grep -rnE '(^|[" ])(sm|md|lg|xl|2xl):grid-cols' frontend/layers/admin/app/components
```

Worth considering a lint rule or a unit test over the component source instead,
in the spirit of the boundary enforcers — the restructure note's repeated lesson
is that a convention with no detector drifts.

---

## Page plans

Worked-out plans for pages under active design. Everything else is just a row in
the table below.

### `/admin` — home dashboard

**Status:** planned, nothing implemented. Direction agreed 2026-09-21.

Today it shows two all-time counters (13,477 users / 79,649,728 points), five
raw feedback entries, and an empty "Aktive prosjekter" block. The counters are
cumulative totals that never meaningfully move — `+0 denne uken` is the tell —
and read identically during a camp week and a quiet September. Wayfarer is built
around time-bound camps, so the page should answer **"what needs me today?"**
rather than "how big are we all-time?".

#### Fix these three bugs first

Small, independent, and worth landing before any redesign — the third one is
user-visible breakage.

1. **Upcoming projects are fetched and then thrown away.** The query asks for
   `endDateAfter: $now` — every project that has not ended — then
   `useGroupedProjects` renders only `currentProjects` (started _and_ not
   ended). `futureProjects` is computed and never used. Between camps you get
   "Her er det tomt" while the next camp's project sits unrendered in the same
   response. `admin/index.vue:56`.
2. **Two stat cards in a four-column grid.** `AdminDashboardStats.vue` builds
   2 cards into `grid-cols-2 lg:grid-cols-4`, so they fill half the row and
   leave the right half empty. It is also a viewport breakpoint where the panel
   needs a container query (the sidebars take ~300px first — see the restructure
   note's container-query sweep). The component renders 3 of the 6 fields
   `AdminDashboardStats` exposes: `totalProjects`, `totalChallenges` and
   `activeProjectsCount` are in the schema and unused. Fixing it means moving the
   grid onto `@container` **inside the component** per Conventions — and picking
   the container value deliberately, since `@lg:` triggers at half the width
   `lg:` does.
3. **A `project_admin` sees nothing but an error state.** `/admin` declares no
   `permission`, and `canAccessAdmin` includes `isProjectAdmin`
   (`usePermissions.ts:71`), so they reach the page. But `adminDashboardStats`
   is `@requireRole(["admin","superadmin"])` and the directive _returns an
   error_ on mismatch (`internal/graph/directives/auth.go`) rather than null.
   The field is non-null, so the whole query fails and `v-else-if="error"`
   replaces the entire dashboard — even though `me`, `feedback` and `projects`
   all resolved. Either split the stats into their own query so it can fail
   alone, or give the page a `permission` and send project admins to their
   project instead.

#### Agreed direction, in priority order

**1 — Current / next project block.** The highest-value item. The active project
becomes the hero of the page, not a card in a grid below the fold: participants,
teams, active challenges, and the project's own dates. When nothing is active,
the same block shows the **next** project with a countdown and a readiness
check — challenges, achievements, events, teams, branding — which turns today's
dead-end empty state into the most useful thing on the page. This is the fix for
bug 1 done properly rather than separately.

**2 — "What needs me".** Replaces the raw feedback list with a triage queue:

- unhandled feedback, as a count plus the newest few, actionable in place
- projects starting soon with incomplete setup
- projects with zero published challenges or achievements
- consents needing attention

**Maintenance jobs are deliberately not in this queue** — those tools are rarely
used, so surfacing them on the home page every day costs attention and returns
nothing. They stay at `/admin/maintenance`.

**3 — Trend sparklines.** Daily active users and challenge completions over
14–30 days, with a direction indicator. Wanted partly for the visual weight it
gives the page — the current layout is two small cards in a wide empty row.
A flat line in September is informative in a way that a cumulative total is not.
**This is the only item that needs backend work** (see below), so it should not
block 1 and 2.

**4 — Quick actions.** New project, new challenge, award points. Useful, low
priority; each is currently 2-3 navigations away. Cheap to add once the blocks
above settle, and easy to leave out.

#### What is queryable today

Verified against `gql/` — 1, 2 and 4 need no schema changes.

| Need                                               | Source                                                                                                                                            |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Participants in a project                          | `users(first: 0, filter: { projectId }) { totalCount }` — `UserFilter.projectId` exists                                                           |
| Teams in a project                                 | `teams(first: 0, filter: { projectId }) { totalCount }`                                                                                           |
| Challenge / achievement / event / superteam counts | the counts-only pattern already proven in `projects/[projectId]/index.vue`                                                                        |
| Active challenges                                  | `Project.activeChallengesCount`                                                                                                                   |
| **Unhandled feedback count**                       | `feedback(first: 0, filter: { handled: false }) { totalCount }` — `UserFeedback.handledAt` and `FeedbackFilter.handled` both exist                |
| Acting on feedback in place                        | `markFeedbackHandled(feedbackId)` mutation                                                                                                        |
| Next / upcoming projects                           | `projects(filter: { startDateAfter: $now })`, ordered client-side — `ProjectFilter` has no sort                                                   |
| Leaderboard panel                                  | `Project.leaderboard(entityType:)` — and `AdminTopPerformers.vue` already exists, fully built and **rendered nowhere**. Use it here or delete it. |
| Daily actives / completions over time              | **Nothing.** See below.                                                                                                                           |

#### What needs backend work

Only the sparklines. `AdminDashboardStats` is six scalar counters and there is no
time-series query anywhere in `gql/`, so item 3 means new SQL in
`backend/internal/database/queries/`, new GraphQL fields, then `make generate`
and `pnpm codegen`. Load the `dataviz` skill before drawing anything.

---

## Pages

### Top level

| Page                              | Route                       | LOC      | St. | Notes                                                                                                     |
| --------------------------------- | --------------------------- | -------- | --- | --------------------------------------------------------------------------------------------------------- |
| `index.vue`                       | `/admin`                    | 134      | ◐   | **Plan agreed — see Page plans.** Three bugs to fix first. No breadcrumb by design (one crumb = a title). |
| `projects/index.vue`              | `/admin/projects`           | 119      | ☐   | Card grid, container queries done. No search; `first: 100`.                                               |
| `projects/new.vue`                | `/admin/projects/new`       | 167      | ☐   |                                                                                                           |
| `users/index.vue`                 | `/admin/users`              | 129      | ☐   | Best-in-class list: debounced server search + relay pagination. Use as the reference pattern.             |
| `users/[userId]/index.vue`        | `/admin/users/:userId`      | **1063** | ☐   | Outlier, see #4. Feedback panel `first: 100`.                                                             |
| `users/[userId]/achievements.vue` | `…/achievements`            | 389      | ☐   | `first: 200` picker — see #1, wants a searchable select, not pagination.                                  |
| `churches/[churchId].vue`         | `/admin/churches/:churchId` | 175      | ☐   | Drill-down leaf by design — no list page, no nav entry (decision recorded in restructure note).           |
| `consents/index.vue`              | `/admin/consents`           | 109      | ☐   | Table, no search, no pagination, no loading state.                                                        |
| `consents/[consentId].vue`        | `/admin/consents/:id`       | 305      | ☐   |                                                                                                           |
| `consents/new.vue`                | `/admin/consents/new`       | 187      | ☐   |                                                                                                           |
| `feedback/index.vue`              | `/admin/feedback`           | 528      | ☐   | Three filter facets + pagination; realtime via Firestore. Largest list page.                              |

### Project-scoped (`/admin/projects/:projectId/…`)

| Page                                    | Route                     | LOC | St. | Notes                                                                                                              |
| --------------------------------------- | ------------------------- | --- | --- | ------------------------------------------------------------------------------------------------------------------ |
| `[projectId].vue`                       | (parent shell)            | 21  | –   | Deliberately thin; project lives in `useCurrentProject()`. Do not delete `index.vue` beneath it — blank-page trap. |
| `[projectId]/index.vue`                 | `…/:projectId`            | 143 | ☐   | Overview: counts-only query + section links. Redirects legacy `?tab=`.                                             |
| `[projectId]/edit.vue`                  | `…/edit`                  | 297 | ☐   | "Innstillinger". Only route a `project_admin` can actually use.                                                    |
| `challenges/index.vue`                  | `…/challenges`            | 119 | ☐   | `first: 50`, no search.                                                                                            |
| `challenges/new.vue`                    | `…/challenges/new`        | 171 | ☐   | Events dropdown `first: 100`.                                                                                      |
| `challenges/[challengeId]/index.vue`    | `…/challenges/:id`        | 222 | ☐   |                                                                                                                    |
| `challenges/[challengeId]/quiz.vue`     | `…/:id/quiz`              | 417 | ☐   | Two-crumb page.                                                                                                    |
| `challenges/[challengeId]/sessions.vue` | `…/:id/sessions`          | 602 | ☐   | Two-crumb page. 13 toast calls — likely the noisiest page in the app.                                              |
| `achievements/index.vue`                | `…/achievements`          | 169 | ☐   | Drag-reorder. `first: 50`.                                                                                         |
| `achievements/new.vue`                  | `…/achievements/new`      | 134 | ☐   |                                                                                                                    |
| `achievements/[achievementId].vue`      | `…/achievements/:id`      | 324 | ☐   |                                                                                                                    |
| `events/index.vue`                      | `…/events`                | 94  | ☐   | Created during restructure to un-orphan the two below. `first: 50`.                                                |
| `events/new.vue`                        | `…/events/new`            | 94  | ☐   |                                                                                                                    |
| `events/[eventId].vue`                  | `…/events/:id`            | 188 | ☐   |                                                                                                                    |
| `superteams/index.vue`                  | `…/superteams`            | 132 | ☐   | The real list (was a tab). `first: 50`.                                                                            |
| `superteams/new.vue`                    | `…/superteams/new`        | 103 | ☐   |                                                                                                                    |
| `superteams/[superTeamId].vue`          | `…/superteams/:id`        | 278 | ☐   | Teams `first: 200`.                                                                                                |
| `superteams/distribute.vue`             | `…/superteams/distribute` | 657 | ☐   | Ladder-to-heaven tool. `@unovis/vue` charts, raw `fetch` to two plugin endpoints — not GraphQL.                    |
| `teams/index.vue`                       | `…/teams`                 | 132 | ☐   | Paginated. No search.                                                                                              |
| `teams/[teamId].vue`                    | `…/teams/:id`             | 428 | ☐   | 14 toast calls.                                                                                                    |
| `scores/index.vue`                      | `…/scores`                | 233 | ☐   | Paginated. No search/date filter.                                                                                  |
| `scores/new.vue`                        | `…/scores/new`            | 131 | ☐   | Project picker removed — route supplies it.                                                                        |

### My church (`church-admin` layout — no navigation, see #5)

| Page                                             | Route              | LOC      | St. | Notes                                                         |
| ------------------------------------------------ | ------------------ | -------- | --- | ------------------------------------------------------------- |
| `my-church/index.vue`                            | `/admin/my-church` | 77       | ☐   |                                                               |
| `my-church/units.vue`                            | `…/units`          | **1109** | ☐   | Outlier, see #4. Worst limit in the codebase (`first: 1000`). |
| `my-church/admins.vue`                           | `…/admins`         | 354      | ☐   | Client-side fuzzy search. `first: 500`.                       |
| `my-church/statistics.vue`                       | `…/statistics`     | 194      | ☐   |                                                               |
| `my-church/kickoff.vue`                          | `…/kickoff`        | 166      | ☐   |                                                               |
| `my-church/gamenights/index.vue`                 | `…/gamenights`     | 59       | ☐   |                                                               |
| `my-church/gamenights/gamenight-[gamenight].vue` | `…/gamenights/:n`  | 642      | ☐   | Largest my-church page.                                       |

### Maintenance

| Page                                   | Route                    | LOC | St. | Notes                                                |
| -------------------------------------- | ------------------------ | --- | --- | ---------------------------------------------------- |
| `maintenance/index.vue`                | `/admin/maintenance`     | 66  | ☐   | Static tool list.                                    |
| `maintenance/bulk-jobs.vue`            | `…/bulk-jobs`            | 381 | ☐   | Paginated.                                           |
| `maintenance/check-points-journal.vue` | `…/check-points-journal` | 149 | ☐   | Hardcoded `ACHIEVEMENT_ID`. `first: 50`.             |
| `maintenance/fix-content-progress.vue` | `…/fix-content-progress` | 367 | ☐   | Job polling — was broken pre-restructure, now typed. |
| `maintenance/fix-streak-progress.vue`  | `…/fix-streak-progress`  | 350 | ☐   | Same shape as above.                                 |

### Legacy redirect stubs (not UX surfaces)

| Page                 | Route                  | LOC | St. | Notes                                                                                |
| -------------------- | ---------------------- | --- | --- | ------------------------------------------------------------------------------------ |
| `teams/index.vue`    | `/admin/teams`         | 11  | –   | Static redirect → project picker.                                                    |
| `teams/[teamId].vue` | `/admin/teams/:teamId` | 61  | –   | Async: looks up `Team.parentProject` then redirects. Has an explanatory empty state. |
| `scores/index.vue`   | `/admin/scores`        | 11  | –   | Static redirect.                                                                     |
| `scores/new.vue`     | `/admin/scores/new`    | 11  | –   | Static redirect.                                                                     |

---

## Update log

### 2026-09-21 — home dashboard planned

First page taken up. Plan is in Page plans above; direction agreed with the user
rather than proposed unilaterally: current/next project block first, then the
"what needs me" queue, sparklines as a visual nice-to-have, quick actions last.
Maintenance jobs were dropped from the triage queue — those tools are rarely
used, so a permanent home-page slot for them costs attention and returns
nothing.

**Three bugs found while reading the page**, all pre-existing, none previously
recorded:

- Upcoming projects are fetched and discarded (`futureProjects` computed, never
  rendered) — which is the actual reason the screenshot shows an empty "Aktive
  prosjekter" block between camps.
- Two stat cards rendered into a 4-column grid, so half the row is empty. Also
  the first confirmed **viewport** breakpoint inside an admin-layout _component_ —
  the earlier sweep and the inventory both scanned `pages/` only, so components
  are an unaudited surface for that bug class.
- A `project_admin` gets an error state for the whole page: `/admin` sets no
  `permission`, `canAccessAdmin` admits them, but `adminDashboardStats` is
  admin/superadmin-only and `@requireRole` returns an _error_ rather than null on
  a non-null field, failing the entire query. Verified in
  `internal/graph/directives/auth.go` rather than assumed — a null-returning
  directive would have made this a cosmetic gap instead of a broken page.

**A convention came out of this**, recorded in Conventions rather than only in
this plan: a component should be responsive to its own container, via
`@container` on its own root, so layouts never style their children. The
component scan it prompted found exactly one violation, so the rule is mostly
about keeping new work in line. The trap found while writing it down is that
`@lg:` and `lg:` are **not** the same width — container variants use the
`--container-*` scale, viewport variants `--breakpoint-*`, and they differ by
about 2x — so a mechanical rename reflows a component at half the intended
width.

**Two findings that made the plan cheaper than expected**, both verified in
`gql/`:

- **Feedback triage already exists server-side.** `UserFeedback.handledAt`,
  `FeedbackFilter.handled` and a `markFeedbackHandled` mutation are all present,
  so "unhandled feedback" is a `totalCount` and acting on it in place is one
  existing mutation. The triage queue needs no schema work.
- **`AdminTopPerformers.vue` is fully built and rendered nowhere.** Either it
  fills the leaderboard slot in the project block or it should be deleted.

Net: items 1, 2 and 4 are frontend-only on data that already exists. Item 3
(sparklines) is the only one needing new SQL and GraphQL — recorded so it does
not block the rest.

### 2026-09-21 — scope narrowed after review

i18n and mobile are **out**: admin is worked in by a handful of Norwegian
users, on desktop. That drops the two largest items the inventory had found — 37
untranslated pages and nine mobile-unsafe tables — and they are recorded as
decisions above so they do not come back as findings.

What survives is reordered by priority: pagination (1), the three states (2),
search/filter/sort (3), the two outliers (4), `church-admin` navigation (5).

One thing deliberately kept despite the i18n decision: the `setLocale('nb')`
side effect in `layouts/admin.vue` is a **consumer-app** bug — it permanently
switches a non-Norwegian end user's app language — so it stands on its own
rather than dying with the i18n item.

Sharpened while rewriting: #1 is not one fix but three. Four of the twelve
limits are on dropdowns and pickers, where pagination is the wrong answer, and
`achievements/index.vue` has drag-to-reorder that does not compose with paging
at all. Recorded inline so the sweep does not mechanically paginate a select.

### 2026-09-21 — inventory created

All 49 files enumerated with line counts, route names, permission meta and a
signal scan for loading/error/empty states, pagination, search, `UTable` use,
`t()` calls and toast use. Cross-cutting issues above come from that scan; no
page has been changed yet.

Two claims checked rather than assumed. Both concerned narrow viewports, which
the next entry puts out of scope — kept only so the measurement is not repeated:

- **No admin page has an `overflow-x` wrapper.** Grepped the whole pages tree,
  zero hits. Now a non-issue: admin is desktop-only.
- **The only two pages using viewport (not container) grid breakpoints are the
  `my-church/*` ones**, which is correct — `church-admin` has no sidebar, so
  container queries would be wrong there. Worth knowing before #5 gives that
  layout navigation: adding a sidebar inverts this, and those grids would then
  be measuring the wrong box.
