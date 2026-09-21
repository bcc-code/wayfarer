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

Settled, so they do not get re-raised. Each one removes or redirects work the
inventory had flagged.

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

**This does not excuse `sm:`/`lg:` grid breakpoints inside the panel** — see #7.
That is a _desktop_ bug, not a mobile one: the two sidebars take ~300-600px out
of the window before the panel gets any, so a viewport breakpoint fires at the
wrong width on exactly the wide screens admin is used on.

**Global all-time counters are not wanted on any admin page.** Total users,
total projects, total challenges, total points awarded — none of these are
things the people running a project care about, so `adminDashboardStats` has no
consumer. Confirmed by the user 2026-09-21 after seeing them rendered;
`AdminDashboardStats.vue` is deleted rather than left orphaned. Anything
numeric on a dashboard has to be **scoped and current** — about a project that
is running now — or it does not belong.

**Challenges are episodic, so "no active challenges" is not a problem.** They
are opened at specific points in a bible study to introduce interactivity;
stretches with nothing active are the normal state of a running project. Any
health signal built on challenge counts is therefore wrong by construction —
absence of activity carries no information here.

**Pagination is cursor-based, and stays that way.** Keyset/Relay cursors are the
convention; offset pagination is not wanted. Confirmed by the user 2026-09-21,
which closes this as a decision rather than a deferred option.

Consequences, since two of them are easy to get wrong:

- **Numbered pages are out permanently.** Every data-table reference we looked
  at shows `1 2 3 … 7`, and jumping to page 5 requires an `OFFSET` the
  convention rules out. `AdminListView`'s Forrige/Neste plus a position readout
  is the end state, not a placeholder.
- **The `from` URL param is a display-only offset.** It feeds the
  "Viser 16–30 av 13 477" label and is **never sent to the API** — the query
  variables remain `first`/`after` or `last`/`before`. Counting steps
  client-side is what makes the label possible at all; it is not offset
  pagination sneaking in.
- **Sorting is still possible, but it is a compound-cursor change, not an
  offset one.** Ordering by name means the cursor has to encode `(name, id)` —
  the sort key plus a stable tiebreaker — because "after this row" is only
  well-defined relative to the ordering. Today's cursor is `base64(id)` alone
  (`internal/graph/pagination/cursor.go:15`), so adding sort means widening the
  cursor, not adding `OFFSET`. Worth knowing before starting.
- **Do not cite `external_content` as precedent.** Its SQL has `queryoffset`
  and it has an `ExternalContentSortBy` enum, but no admin page drives either —
  `AdminContentItemSelector` just asks for `first: 500`. It is the exception,
  not the pattern to follow.

**Project events are rarely used.** Do not build anything that assumes they
are populated. Confirmed by the user 2026-09-21, and consistent with what the
restructure found: the `events/` subtree was _orphaned_ — nothing linked to
`events/new` or `events/[eventId]`, they were reachable only by typing the URL.

Consequences, since this is easy to design around by accident:

- In practice **the project is the camp**, so `Project.startDate`/`endDate` are
  the real camp dates and nothing needs an intermediate layer to be time-aware.
- The fine-grained "what is happening now" signal comes from **challenges**, not
  events — `Challenge` has `publishedAt`, `visibleAt`, `startedAt`, `endTime`.
- The three `events/**` pages stay in the inventory but rank last for UX work.
- `Event.translationStatus` is not worth surfacing; `Project` and `Challenge`
  are.

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

| Page                                                | Limit                   | Shape                                   |
| --------------------------------------------------- | ----------------------- | --------------------------------------- |
| `my-church/units.vue`                               | `first: 1000`           | list, also an outlier (see #4)          |
| `my-church/admins.vue`                              | `first: 500`            | list + client fuzzy search              |
| `projects/[projectId]/superteams/[superTeamId].vue` | `first: 200` (teams)    | picker                                  |
| `users/[userId]/achievements.vue`                   | `first: 200`            | picker                                  |
| `projects/index.vue`                                | `first: 100`            | card grid                               |
| `users/[userId]/index.vue`                          | `first: 100` (feedback) | panel in a detail page                  |
| `projects/[projectId]/challenges/new.vue`           | `first: 100` (events)   | dropdown                                |
| `projects/[projectId]/superteams/distribute.vue`    | `first: 100` (events)   | dropdown                                |
| `projects/[projectId]/challenges/index.vue`         | `first: 50`             | list                                    |
| `projects/[projectId]/achievements/index.vue`       | `first: 50`             | list, drag-reorder                      |
| `projects/[projectId]/events/index.vue`             | `first: 50`             | list — low priority, events rarely used |
| `projects/[projectId]/superteams/index.vue`         | `first: 50`             | list                                    |
| `maintenance/check-points-journal.vue`              | `first: 50`             | table                                   |
| `maintenance/fix-content-progress.vue`              | `first: 50`             | table                                   |

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

### 6. Archived projects are not filtered out anywhere — ☑ fixed 2026-09-21

`ProjectFilter.archived` exists, but the SQL is
`archived IS NULL OR archived = sqlc.narg('archived')`
(`backend/internal/database/queries/projects.sql:102,119`) — so **omitting the
filter includes archived projects**. The resolver passes the filter straight
through (`internal/graph/api/projects.go:21`) and adds no default.

None of the three places that list projects passes it:

| Site                          | Query                                                                                                        |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `admin/index.vue:56`          | `projects(filter: { endDateAfter: $now })` — an archived project with a future end date counts as **active** |
| `projects/index.vue:9`        | `projects(first: 100)`                                                                                       |
| `AdminProjectSwitcher.vue:11` | `projects(first: 100)` — archived projects in the sidebar switcher                                           |

**Fixed 2026-09-21** by adding `archived: false` to all three queries, as part
of the home dashboard work.

Still open, deliberately: whether the default belongs in the **resolver** rather
than at three call sites. The switcher and the list want the same thing, and a
fourth caller will forget — nothing detects a missing filter.

### 7. Bring existing components onto the self-contained rule

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

**Status:** items 1 and 3 shipped 2026-09-21 (see the log). Item 2 ("what needs
me") and item 4 (quick actions) outstanding.

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

**1 — Active / next project block.** The highest-value item, and the one that
has to handle **multiple simultaneously active projects** — theoretically
supported, so the block is designed for `N`, not for one.

- Each active project gets a vitals card: name, dates and countdown,
  participants, teams, active challenges.
- **Ordered by ending soonest.** A camp ending tomorrow needs attention more
  than one that runs for three months; start date is the tiebreak.
- **Capped at three, plus an "N more" link.** Rendering an unbounded number of
  heroes stops the page being answerable at a glance — the
  `projects/[projectId]/index.vue` UTabs kitchen sink is the in-repo cautionary
  tale.
- **Scoped by role.** `getProjectAdminIds` already exists, so a `project_admin`
  sees only their own projects. This is also the cleanest fix for bug 3: they
  get a dashboard that works rather than an error.
- **Zero active** falls back to the next upcoming project with its countdown and
  a readiness check — challenges, achievements, teams, branding. That turns
  today's dead-end empty state into the most useful thing on the page, and is
  bug 1 fixed properly rather than separately.
- **Exclude archived** — see #6; nothing filters them today, so the "active" set
  is currently wrong.

One self-contained `AdminProjectVitals` component covers every count: full width
with one active project, sharing a row with three, the page deciding only the
grid. This is the first real case for the Conventions rule.

**2 — "What needs me".** Replaces the raw feedback list with a triage queue:

- unhandled feedback — a count plus the newest few, actionable in place, and
  **broken down by tag** rather than five truncated messages. A tag summary
  answers "what is breaking" much better; the load complaint visible in the
  current page ("appen ikke fungerer når alle i salen skal bruke den samtidig")
  lines up with the loadtest notes in this repo.
- ~~`activeChallengesCount === 0` while a project is running~~ — **rejected.**
  Challenges are episodic (see Scope decisions), so zero active is ordinary and
  the "alarm" fires constantly on healthy projects. It shipped, was seen, and
  was removed the same day.
- challenge lifecycle problems, derivable from `publishedAt` / `visibleAt` /
  `startedAt` / `endTime`. These still hold, because each is a **contradiction**
  rather than an absence: a challenge published but with `visibleAt` unset will
  never appear; one whose `endTime` has passed while still marked active is
  inconsistent. Absence of activity is not a signal here — only incoherent
  configuration is.
- **incomplete translations** — `translationStatus` on `Project` and
  `Challenge`. Unlike admin, the user-facing app _is_ translated, so a missing
  translation is a real end-user defect. Queryable today; nothing surfaces it.
- projects with zero published challenges or achievements; consents needing
  attention.

**Maintenance jobs are deliberately not in this queue** — those tools are rarely
used, so a permanent home-page slot costs attention and returns nothing. They
stay at `/admin/maintenance`.

**3 — Trend sparklines.** Daily active users and challenge completions over
14-30 days, with a direction indicator. Must be **scoped to a running project**,
not global: the same objection that killed the counters applies to a global
trend line. A flat line for an active camp is informative; a flat line across
all projects ever is not.
**The only item needing backend work** (see below), so it must not block 1 and 2.
Load the `dataviz` skill before drawing anything.

**4 — Quick actions.** New project, new challenge, award points. Useful, low
priority; each is currently 2-3 navigations away. Cheap to add once the blocks
above settle, and easy to leave out.

#### Considered and rejected

- **An event timeline.** Would have been inherently multi-project, and was the
  first answer to the `N` active projects problem. Dropped: project events are
  rarely used (see Scope decisions), so it would be a prominent block over an
  empty table. Challenge lifecycle fields give the same "what is happening now"
  signal from the entity that _is_ used.
- **A join/team funnel** (joined vs. actually in a team) — genuinely useful at
  kickoff, but `Team` has no member-count field, only `members: [TeamMember!]!`,
  so counting means fetching every member of every team. Needs a schema addition
  first; not free.
- **Push-send as a quick action.** `sendPushNotification` exists and is
  admin-only, but a button that messages every participant does not belong on a
  dashboard without a deliberate confirm step. Revisit with quick actions.

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

Everything in items 1, 2 and 4 is frontend-only on data that already exists.
Three things are not, in rough order of value:

1. ~~Time-series for the sparklines~~ — **done 2026-09-21.**
   `Project.activityTrend(days:)` over `score_journal`.
2. **Per-challenge completion counts** — would give the best detector of the
   set, "this challenge is live and nobody has completed it", which catches a
   broken QR code or a wrong date _during_ a camp rather than after.
3. ~~`Project.participantCount` / `teamCount`~~ — **not needed.** Resolved on
   the frontend instead: a per-project component runs its own counts query, so
   the root `users`/`teams` connections suffice. See the 2026-09-21 section
   entry. (A `Team.memberCount` would still be needed for the join/team funnel,
   which is a different thing.)

Each means new SQL in `backend/internal/database/queries/`, new GraphQL fields,
then `make generate` and `pnpm codegen`.

---

## Pages

### Top level

| Page                              | Route                       | LOC      | St. | Notes                                                                                                     |
| --------------------------------- | --------------------------- | -------- | --- | --------------------------------------------------------------------------------------------------------- |
| `index.vue`                       | `/admin`                    | 134      | ◐   | **Plan agreed — see Page plans.** Three bugs to fix first. No breadcrumb by design (one crumb = a title). |
| `projects/index.vue`              | `/admin/projects`           | 119      | ☐   | Card grid, container queries done. No search; `first: 100`.                                               |
| `projects/new.vue`                | `/admin/projects/new`       | 167      | ☐   |                                                                                                           |
| `users/index.vue`                 | `/admin/users`              | 199      | ◐   | **On `AdminListView`** — the reference conversion. Search + church filter + URL state.                    |
| `users/[userId]/index.vue`        | `/admin/users/:userId`      | **1063** | ☐   | Outlier, see #4. Feedback panel `first: 100`.                                                             |
| `users/[userId]/achievements.vue` | `…/achievements`            | 389      | ☐   | `first: 200` picker — see #1, wants a searchable select, not pagination.                                  |
| `churches/[churchId].vue`         | `/admin/churches/:churchId` | 175      | ☐   | Drill-down leaf by design — no list page, no nav entry (decision recorded in restructure note).           |
| `consents/index.vue`              | `/admin/consents`           | 109      | ☐   | Table, no search, no pagination, no loading state.                                                        |
| `consents/[consentId].vue`        | `/admin/consents/:id`       | 305      | ☐   |                                                                                                           |
| `consents/new.vue`                | `/admin/consents/new`       | 187      | ☐   |                                                                                                           |
| `feedback/index.vue`              | `/admin/feedback`           | 601      | ◐   | **On `AdminListView`.** Three facets + URL state; realtime via Firestore. Largest list page.              |

### Project-scoped (`/admin/projects/:projectId/…`)

| Page                                    | Route                     | LOC | St. | Notes                                                                                                                           |
| --------------------------------------- | ------------------------- | --- | --- | ------------------------------------------------------------------------------------------------------------------------------- |
| `[projectId].vue`                       | (parent shell)            | 21  | –   | Deliberately thin; project lives in `useCurrentProject()`. Do not delete `index.vue` beneath it — blank-page trap.              |
| `[projectId]/index.vue`                 | `…/:projectId`            | 143 | ☐   | Overview: counts-only query + section links. Redirects legacy `?tab=`.                                                          |
| `[projectId]/edit.vue`                  | `…/edit`                  | 297 | ☐   | "Innstillinger". Only route a `project_admin` can actually use.                                                                 |
| `challenges/index.vue`                  | `…/challenges`            | 119 | ☐   | `first: 50`, no search.                                                                                                         |
| `challenges/new.vue`                    | `…/challenges/new`        | 171 | ☐   | Events dropdown `first: 100`.                                                                                                   |
| `challenges/[challengeId]/index.vue`    | `…/challenges/:id`        | 222 | ☐   |                                                                                                                                 |
| `challenges/[challengeId]/quiz.vue`     | `…/:id/quiz`              | 417 | ☐   | Two-crumb page.                                                                                                                 |
| `challenges/[challengeId]/sessions.vue` | `…/:id/sessions`          | 602 | ☐   | Two-crumb page. 13 toast calls — likely the noisiest page in the app.                                                           |
| `achievements/index.vue`                | `…/achievements`          | 169 | ☐   | Drag-reorder. `first: 50`.                                                                                                      |
| `achievements/new.vue`                  | `…/achievements/new`      | 134 | ☐   |                                                                                                                                 |
| `achievements/[achievementId].vue`      | `…/achievements/:id`      | 324 | ☐   |                                                                                                                                 |
| `events/index.vue`                      | `…/events`                | 94  | ☐   | **Low priority** — events rarely used, see Scope decisions. Created during restructure to un-orphan the two below. `first: 50`. |
| `events/new.vue`                        | `…/events/new`            | 94  | ☐   | **Low priority** — events rarely used.                                                                                          |
| `events/[eventId].vue`                  | `…/events/:id`            | 188 | ☐   | **Low priority** — events rarely used.                                                                                          |
| `superteams/index.vue`                  | `…/superteams`            | 132 | ☐   | The real list (was a tab). `first: 50`.                                                                                         |
| `superteams/new.vue`                    | `…/superteams/new`        | 103 | ☐   |                                                                                                                                 |
| `superteams/[superTeamId].vue`          | `…/superteams/:id`        | 278 | ☐   | Teams `first: 200`.                                                                                                             |
| `superteams/distribute.vue`             | `…/superteams/distribute` | 657 | ☐   | Ladder-to-heaven tool. `@unovis/vue` charts, raw `fetch` to two plugin endpoints — not GraphQL.                                 |
| `teams/index.vue`                       | `…/teams`                 | 231 | ◐   | **On `AdminListView`** + superteam filter. Slot-name bug fixed.                                                                                                           |
| `teams/[teamId].vue`                    | `…/teams/:id`             | 428 | ☐   | 14 toast calls.                                                                                                                 |
| `scores/index.vue`                      | `…/scores`                | 271 | ◐   | **On `AdminListView`** + source-type filter.                                                                                               |
| `scores/new.vue`                        | `…/scores/new`            | 131 | ☐   | Project picker removed — route supplies it.                                                                                     |

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

### 2026-09-21 — teams, scores and bulk-jobs converted; RelayPagination retired

All four previously-paginated lists are now on `AdminListView`, so
`RelayPagination.vue` is **deleted**. Each page also gained a filter it did not
have, chosen from what its own filter input already supported:

| Page | Filter added | Source |
| --- | --- | --- |
| teams | superteam, with an "Uten superlag" option | `TeamFilter.superTeamId` + `noSuperTeam` |
| scores | source type | `ScoreJournalFilter.sourceType` |
| bulk-jobs | (kept its two) now URL-synced | `BulkJobFilter` |

None of the three has a free-text field in its filter input, so all are
`:searchable="false"` — a search box that cannot work is worse than none.

**The teams superteam filter folds two API fields into one control.**
`superTeamId` and `noSuperTeam: Boolean` are separate inputs; a single select
with a `__none__` sentinel gives three states — all / a specific superteam /
unassigned. The third is the question worth asking before running the
distribution tool, and it had no UI at all.

**Two bugs found while converting:**

- **`teams/index.vue` had a slot that never rendered.** The column declares
  `id: 'superTeam'` and the template was `#superTeam`, but Nuxt UI's UTable
  looks for `<id>-cell`. So the custom cell — whose only job is an em-dash
  fallback for teams with no superteam — was dead, and those teams rendered
  blank. Every other slot on that page was correctly suffixed, which is why it
  went unnoticed.
- **`bulk-jobs.vue` had the same double empty state as feedback**: a sibling
  `<UEmpty v-if="!fetching && jobs?.length === 0">` below `<UTable>`, so an
  empty list showed two. Both are now in the table's `#empty` slot with the
  filtered-miss distinction.

**Two typing details worth knowing for the next conversion.** Codegen emits
**real TS enums**, not string unions, so a list of string literals does not
typecheck against `ScoreSourceType` — use `Object.values(...)`. And filter
values must stay plain `string` in the items passed to `USelect`: typing them
as the enum makes the component demand an enum-typed model, which fights
`useListState`'s string-only contract. The cast back to the enum belongs at the
query, not in the control.

Gate: typecheck 0, lint 0 errors, 604 unit + 199 component, build exit 0.

### 2026-09-21 — feedback list converted to AdminListView

Second adopter, and the one that proved the component generalises: feedback has
three filters where users has one, a multi-select and a tri-state among them,
and **no free-text search at all** — `FeedbackFilter` exposes only `userId`,
`tags`, `handled` and `platform`, so `:searchable="false"` rather than a search
box that cannot work.

**`useListState` stays string-only and the page bridges.** URL params *are*
strings, so teaching the composable about arrays and booleans would push URL
encoding into it. Instead two `computed` with getters/setters adapt:
`tags` ⇄ comma-joined string, and `handled` ⇄ a tri-state where unset means "no
filter" and is distinct from `handled: false` — collapsing those two would
silently turn "show me everything" into "show me unhandled".

Removed from the page: its own reset watcher (`watch([selectedTags,
selectedPlatform, handledFilter], …)`), its hand-rolled "Nullstill" button, and
`RelayPagination`. All three are now the component's job, and the reset in
particular is exactly the bug the composable exists to prevent — this page had
three filters in that watcher's dependency array and a fourth would have been
easy to miss.

**A double empty state, found while converting.** The page rendered
`<UEmpty v-if="!fetching && feedbacks?.length === 0">` as a *sibling* of
`<UTable>`, so an empty list showed Nuxt UI's own empty row **and** that block
beneath it. Folded into the table's `#empty` slot, where it also gained the
filtered-miss/truly-empty distinction.

**One limitation recorded rather than engineered around:** tags are
comma-joined in the URL, so a tag containing a comma splits into two filter
values. Tags are admin-authored via `UInputTags` and the breakage is visible in
the chip rather than silent. Repeatable `?tags=a&tags=b` params are the robust
fix if tags ever become user-authored.

Gate: typecheck 0, lint 0 errors, 604 unit + 199 component, build exit 0. No new
tests — the conversion added no logic of its own, and the composable and
component are already covered.

### 2026-09-21 — AdminListView + useListState; page position in the URL

A standardised list toolbar and footer, plus URL-synced list state, built
against the users page. Four reference dashboards the user supplied all agree
on the same anatomy: **toolbar above** (primary action · search · filters),
**footer below** (position + page size left, pagination right).

**`AdminListView`** renders toolbar → the page's own table in the default slot
→ footer, with `#filters` and `#actions` slots. Slot-based rather than a
config-driven table, so an odd page composes instead of fighting it.

**`useListState`** owns search, filters and position, synced to the URL. It also
owns the **pagination reset**, which is the correctness half: each list page
currently wires its own `watch(debouncedSearch, () => pagination.reset())`, and
a page that adds a second filter and forgets to extend that watch asks for page
4 of a one-page result and gets an empty table with no error. Mutation-checked —
disabling the reset fails exactly the two reset cases.

**The page position _is_ in the URL, and my first answer that it should not be
was wrong.** I told the user a cursor "shifts as data changes" and would
"silently land somewhere else later". That is backwards. The cursor is
`base64(id)` (`internal/graph/pagination/cursor.go:15`) and the SQL is
`WHERE u.id > @aftercursor` — a **comparison, not a lookup**. So a keyset cursor
means "the rows after user X" regardless of inserts or deletes before it, and it
still resolves if user X is deleted. A page _number_ is the thing that shifts.

What a cursor genuinely cannot carry is the position _label_, which is the real
obstacle and a much smaller one: `from=<offset>` rides along for the "Viser
16–30" text, and a stale `from` mislabels a correct page rather than showing the
wrong rows.

The shape that made it exact: **persist the pagination variables, not a page
number.** With keyset pagination the variables _are_ the position, so
`?after=<cursor>&from=15` (or `before` for backward) round-trips precisely, with
no cursor stack and no walking. `usePagination.restore()` applies it before the
first query, so a deep link fetches the right page once rather than fetching
page 1 and jumping.

URL params are split into two groups because they behave oppositely: `q` and
filters **must** reset the position; `after`/`before`/`from`/`size` **must not**,
or the list could never leave page 1. Both directions are pinned by tests.

**Position display needed no backend change either** — another thing I had
assumed. Keyset cannot _report_ a position, but it can be counted: `nextOffset`
/ `previousOffset` track steps, and `pageRange` derives "Viser 16–30 av 13 477"
from the offset plus the rows actually returned (so a short final page does not
overshoot). Only _jumping_ to an arbitrary page needs `OFFSET` on the backend.

**Numbered pages and column sorting were left open here and have since been
settled** — see Scope decisions. Numbered pages are **out**: they need `OFFSET`,
and cursor pagination is the convention. Sorting remains possible but requires
widening the cursor to encode the sort key, not adding `OFFSET`.

**A bug in my own code that a test caught:** `useListState` compared page size
against `pageSizes[0]` rather than the pagination's own default, so any list
defaulting to 20 while offering 15 first would have carried `?size=20` on every
URL forever. It now captures the default before the URL can override it.

**Users page, converted.** Gains `?q=`/`?churchId=`/position in the URL, a
searchable church filter (the first of `UserFilter`'s seven unused dimensions),
filter chips with a reset, the position readout, a page-size selector, labelled
Forrige/Neste below the table, a "nothing matched" empty state distinct from
"no users", and the dead `actions` column removed (it declared
`{ id: 'actions' }` with no cell template — unique among admin tables).

On the consumer filter-panel reference: pills, segmented controls and the live
result count transfer; the **Cancel/Apply panel does not**. It hides filter
state behind a step and fights the linkable-URL goal. If a list outgrows an
inline toolbar, the data-table pattern is a "Filtre" popover with the chips
still visible — not a modal.

Tests: 10 unit (`pagination-range.test.ts`), 9 component (`AdminListView`),
15 component (`useListState`). Unit 594 → 604, component 175 → 199.

### 2026-09-21 — AdminTopPerformers deleted

The leaderboard panel that was built but rendered nowhere. It was flagged in the
first pass over this page ("use it here or delete it") and stayed open through
the whole effort — deleted rather than carried further. No references anywhere
and it declared no queries of its own, so nothing else moved.

`components/admin/dashboard/` now holds only `AdminRecentActivity.vue`, which
is itself a misnomer: it renders unhandled feedback, not recent activity. Worth
renaming when that panel is next touched.

### 2026-09-21 — sparkline interaction, empty state, accent ring removed

Follow-ups from looking at the rendered chart. Three of the four were only
findable by looking; the gate was green throughout.

**Hover did not work at all.** Two causes, and the second was the user's.

- The hit target was the painted bar, which on a quiet day is a 2px sliver with
  nothing to aim at. Each day now has a transparent full-height column spanning
  its whole slot, painted above the bars, so the reader only has to be over the
  right column — a zero-value day is hoverable for the first time. The dataviz
  guidance says this outright ("the hit target is bigger than the mark"); I
  had read it and still shipped bar-only hit areas.
- The native SVG `<title>` was the wrong mechanism: ~1s delay, unstyleable, and
  never shown on keyboard focus, so it reads as nothing happening. Replaced with
  a styled tooltip carrying value + date, positioned by the slot centre as a
  percentage so it tracks the column at any container width, with the hovered
  bar lifting to full opacity.

Deliberately pointer-only. Adding `tabindex` to 14 columns per chart is 28 tab
stops on this page; the visually-hidden table is the keyboard and screen-reader
path, which is why it exists.

**An all-zero window now says so in words.** A flat series still occupies the
plot's full height, so it rendered as a tall empty box with a lone baseline
adrift in it — reading as a broken chart, the exact thing the baseline was added
to prevent. It is a sentence now ("Ingen aktivitet siste 14 dager"). Quiet
stretches are normal here, so this is the common state, not an edge case.

**The branding accent ring is gone from the project card.** It was meant to tell
stacked sections apart, but with one active project there is nothing to tell
apart and an arbitrary per-project hue on a border just reads as random — the
logo already carries identity. Removing it made `branding.colors` dead, so it
came out of the query and the prop type too.

**A test-assertion habit worth naming, because it went wrong twice.** Both the
permission-gating case and the first tooltip cases asserted against
`wrapper.text()` — but the `sr-only` table lists every date and value, so page
text always contains them and the tooltip tests would have passed with the
tooltip completely broken. Assertions are now scoped to the element that owns
the claim: `<nav>` for shortcuts, `[data-slot="tooltip"]` for the tooltip.

### 2026-09-21 — sparklines shipped (item 3), backend included

The last outstanding item on this page. Two stat tiles per running project —
label, aggregate, 14-day bar sparkline — sitting between the header and the
shortcuts.

**Backend.** `Project.activityTrend(days: Int = 14): [ProjectActivityPoint!]!`,
fed by a new `GetProjectActivityTrend` query over **`score_journal`**. That table
is the right source: it has `project_id` directly (no join), an index on
`created_at`, and it records _every_ point award — achievements, quizzes, manual
adjustments — so it reflects all activity rather than challenge completions
alone.

Design choices worth keeping:

- **The SQL returns only days that have rows; Go fills the gaps.** A sparse
  series plotted as adjacent bars lies — three days a week apart read as steady
  activity rather than three isolated spikes. Filling in Go instead of with a
  `generate_series` join keeps the query a plain grouped scan and makes the
  windowing unit-testable without a database.
- **`days` is clamped to 1-90**, and a non-positive value falls back to the
  default rather than erroring: an unbounded window is an easy way to make the
  server do work proportional to whatever a caller types.
- **`trendWindowStart` is shared** between the query's `since` bound and the
  series construction, so the two cannot disagree about the window and the
  query cannot return a day the series has no slot for.
- **Rows outside the window are dropped rather than trusted**, so a stale bound
  or clock skew cannot stretch the series past `days` points.
- **The resolver body lives in `projects.go`, not `projects.resolvers.go`.** The
  generated file holds a one-line delegation, so `make generate` has nothing of
  ours to preserve and its import list stays as gqlgen wrote it.
- Buckets by **UTC day**. Converting to a fixed local zone only moves activity
  between 00:00 and 02:00 local onto the adjacent day — invisible at sparkline
  resolution and not worth hardcoding a timezone in SQL for.
- Ungated, like every other `Project` field. It is aggregate, not per-user.

**Frontend.** Form picked from the `dataviz` skill: this is a _stat tile with a
sparkline_, not a chart. Three of its rules changed what got built:

- **Two tiles, never one chart with two y-axes.** Points run to thousands and
  active users to dozens; a shared scale flattens one into the baseline. A
  second axis is the single most common charting mistake and is never the answer.
- **Bars, not a line.** The data is one discrete bucket per day, and a bar is
  its own hit target — a 2px polyline needs a crosshair layer to be hoverable,
  more machinery than a tile-sized chart earns.
- **The tooltip is not the only path to a value.** Native `<title>` tooltips do
  not appear on keyboard focus, so each sparkline ships a visually-hidden table
  of its daily values and the plot itself is `aria-hidden`.

No palette validation was run, deliberately: the validator checks categorical
palettes for colour-vision separation, and this is a single series in one hue at
two emphasis levels (most recent day accented, the rest de-emphasised). Both are
design-system tokens rather than project branding — branding accent is arbitrary
per project and cannot be contrast-checked up front, which is why the chart does
not use it even though the card's ring does.

**The aggregates differ, and that is a correctness point, not a style one.**
Points are additive so the window total is meaningful. Active users is a
distinct count _per day_ — summing it would count the same person once per day
they appeared — so it is shown as a daily average. `dailyAverage` exists only to
make that impossible to get wrong by reflex, and a test pins it.

**The trend is skipped entirely for a project that has not started**, via
`@include(if: $withTrend)`. Fourteen empty days read as a broken chart rather
than as "not yet".

Two bugs caught by looking at the rendered page rather than by the gate:

- **`sr-only` on a `<table>` does not hide its `<caption>`.** Tailwind's
  `sr-only` sets no `display`, so the element keeps `display: table` and the
  caption is laid out outside the 1px clip box — it rendered as visible text
  under the chart. The class has to go on a wrapping div. Pinned by a test that
  asserts no `table.sr-only` exists.
- **An all-zero window painted nothing**, so the tile looked like a chart that
  had failed to load. There is now a baseline rule, which is also the honest
  reading: flat at zero. A quiet stretch is normal for this domain, so this is
  the common case, not an edge case.

Tests: 13 unit (`sparkline.test.ts`), 8 Go (`projects_activity_trend_test.go`),
7 component (`AdminSparkline.test.ts`) plus 5 more on the section. Unit 577 →
590, component 159 → 171, backend 27 packages green.

One test had to be loosened correctly rather than fixed: the permission-gating
assertion checked page text for "Poeng" and now collides with the "Poeng siste
14 dager" tile label. It is scoped to the `<nav>`, which is what it always
meant.

### 2026-09-21 — page width capped; shortcut tiles go horizontal

**The home page is capped at `max-w-6xl`.** This is a deliberate exception to
the restructure's full-width decision, which swapped `UContainer` for a plain
div on all 40 admin pages. That call was right for tables and card grids and
wrong here: this page is a single column of prose-like blocks, and stretched
across a wide screen the project section became seven tiles spread over
~1600px above a feedback list of one-line entries. The sweep always preserved
`max-w-*` on deliberate elements; this is one.

**Shortcut tiles are one line each** — icon, label, count — instead of three
stacked rows. Denser, and entries with no count (Poeng, Innstillinger) now
simply have none rather than reserving empty space where a number would go.
The grid tracks widened to `minmax(10rem,1fr)`, which lands all seven on one
row at the capped width.

Process note: I first read "the card is very long" as height and made the tiles
horizontal, which was not what was asked. Corrected to a width fix, reverted the
tile change as unrequested — then the user said the horizontal tiles were good
too, so they came back. The width cap was the actual fix; the tile layout was a
lucky accident, and shipping it was the user's call rather than mine.

### 2026-09-21 — progress bar removed; it measured nothing anyone acts on

**Rejected after seeing it: the project progress bar.** The countdown badge
already carries the actionable part ("101 dager igjen"), and elapsed-fraction
informs no decision — being 60% through a project does not change what you do
next. As an unlabelled bar it also did not read as project length at all, which
is what prompted the question.

`projectProgressPercent` and its six unit tests are **deleted** rather than left
behind, along with the four component assertions covering the bar. Dead helpers
with tests still passing are the most convincing kind of dead code.

Worth being honest about why it was there: I chose it partly for the visual
weight it gave a sparse page. That is the wrong reason to put something on a
dashboard, and it is the third design element on this page I added and then
removed for the same underlying mistake — **decoration standing in for
information**. The stats cards, the no-challenges alarm and this bar all failed
the same test.

That space is where the per-project **sparkline** belongs once there is data
behind it. A real trend line earns it; an elapsed-fraction bar does not.

Two things kept from the removal:

- **The branding accent moved to the card's ring.** It was only there to tint
  the bar, but it does real work: once two or three sections stack they need to
  be tellable apart at a glance, and a tinted ring does that without implying a
  measurement.
- **The shortcut grid is now `auto-fit`.** The fixed 6-column grid orphaned the
  7th tile on its own row, and the count varies from 4 to 7 with the viewer's
  permissions, so no fixed number is right —
  `grid-cols-[repeat(auto-fit,minmax(9rem,1fr))]` collapses empty tracks and
  stretches whatever is visible to fill one row.

Gate: typecheck 0, lint 0 errors, **577 unit** (583 → −6) and **159 component**
(163 → −4), build exit 0.

**One unexplained intermittent failure, recorded rather than hidden.** A single
unit test failed twice — `1 failed | 576 passed` — and in neither run did vitest
name it before the summary. Both occurrences were in a shell command that ran
several `pnpm` invocations back to back, which suggested contention between the
two vitest projects over `node_modules/.cache/nuxt`.

That hypothesis is **not confirmed**: ~20 subsequent runs, including three
deliberate reproductions of the exact chained command shape and twelve
consecutive solo runs, are all clean at 577. So it is real but unreproduced and
undiagnosed. If it recurs, the thing to capture is the test name — run the unit
project alone with full output rather than piped through `grep`, which is what
hid it both times.

### 2026-09-21 — active project becomes a section with shortcuts, not a card

The card was the wrong shape. Each active project now gets **its own section**:
header (logo, name, dates, participant count, countdown, progress bar) plus a
grid of **shortcuts into that project's sub-pages, each with a live count**.
Usually there is exactly one active project, so the page is mostly this.

`AdminProjectSummaryCard.vue` is deleted — two card designs in one day, both
replaced. Keeping the write-up of them because the reasoning is what carried
over: the progress bar, the timing helpers and the `UProgress` a11y wiring all
survived into the section unchanged.

**The finding that made this possible: the per-project count constraint was a
document constraint, not a schema one.** I had recorded that participants and
teams could not be counted for N projects and needed new `Project` fields. That
holds only for a single page-level query — GraphQL has no dynamic aliasing. A
**component that owns its own query** sidesteps it completely: each mounted
section asks for its own `projectId`, and urql keys the document cache on
variables so nothing is fetched twice. So participants, teams, superteams,
challenges, achievements and events counts are all live now, with **no backend
work**, and `Project.participantCount`/`teamCount` are off the backend list.

**Shortcuts are `PROJECT_NAV`**, the same model the sidebar renders from, so they
cannot drift from it and permission gating comes for free — `Lag`, `Poeng` and
`Innstillinger` are gated behind `canAccessTeams` / `canAccessScores` /
`canEditProject`. "Oversikt" is filtered out because the section title already
links to the project root.

That gating is also the one thing the component test caught: with no roles
mocked, three shortcuts correctly did not render and my first assertion was
simply wrong. It is now asserted in both directions — granted and denied — which
is the more useful test than the one I set out to write.

Counts are rendered **plainly, never as health signals**, per the scope
decisions: episodic challenges mean a zero is ordinary, and there is a test
asserting a zero count renders as "0" with no warning class.

**Still outstanding: the sparklines.** Nothing changed there — `gql/` has no
time-series query of any kind, so per-project trends (daily active users,
completions per day) need new SQL and new GraphQL fields. This is now the only
part of the home page that needs backend work.

### 2026-09-21 — active-project card redesigned as a hero band

The tile-grid card went too. With the challenge count removed as well (not
something anyone cares about on the home page), the two-tile grid held one real
number and one "Åpne prosjekt" link — structure with nothing to structure.

`AdminProjectVitals.vue` is replaced by **`AdminProjectSummaryCard.vue`**: logo,
name, date range, countdown badge and arrow on one line, with a thin progress
bar underneath showing how far through the project we are. The whole band is the
link. Renamed because "vitals" described stats it no longer shows.

Chosen from three sketched options rather than guessed at — the previous two
design calls on this page were both mine and both wrong, so the direction was
worth one question.

Three things worth keeping:

- **The card is reused for the "next project" block**, which previously had its
  own bespoke header. Progress reads 0% for something that has not started,
  which is correct and meaningful, so one component covers both states and the
  readiness tiles simply sit below it.
- **`activeChallengesCount` is out of the query.** Nothing renders it any more,
  so fetching it was waste.
- **The bar is `UProgress`, not a hand-rolled div** (the user's suggestion; my
  first cut was a div, which threw away real semantics for nothing).

**`UProgress`'s accessibility props are not attributes, and this is worth
knowing before using it anywhere else.** `role="progressbar"` sits on an _inner_
element (`data-slot="base"`), so an `aria-label` on the `<UProgress>` tag lands
on the role-less outer wrapper and is silently ignored — while reka-ui's own
default fills the real element with `aria-label="66%"`, i.e. a percentage as the
element's _name_. The working levers are the props: `get-value-label` feeds that
element's `aria-label` and `get-value-text` its `aria-valuetext`. Found by a
component test failing (`expected '50%' to contain 'Spring Revival'`) and
confirmed against the rendered markup, not reasoned about — the attribute form
looks correct in the template and produces no warning.

`projectProgressPercent` is pure and clamped at both ends, and returns 0 rather
than `NaN` for a zero-length, inverted or unparseable range — bad data renders
an empty bar instead of `NaN%`. Six unit tests cover exactly those edges.

`test/component/AdminProjectSummaryCard.test.ts` (8 cases) is the first
component test in the admin layer. Two of its assertions exist because the
failure mode is silent: the `aria-label`/`aria-valuetext` wiring above, and that
`:ui="{ indicator: 'bg-(--accent)' }"` actually **merges** — a `ui` override
that fails to merge leaves the bar on the default colour with no error anywhere.

Gate: typecheck 0, lint 0 errors, **583 unit** (577 → +6) and **159 component**
(151 → +8), build exit 0.

### 2026-09-21 — stats cards and the challenge alarm removed after review

Both were mine, and both were wrong in the same way: **a number is only useful
on this page if it is scoped and current.**

**The four stat cards are gone**, and `AdminDashboardStats.vue` is deleted
rather than left orphaned. Total users, total projects, total challenges, total
points awarded — nobody running a project cares about any of them. Note that I
had already written the diagnosis for this: the original page's two counters
were criticised in this very note as "cumulative totals that never meaningfully
move". I then fixed the _grid_ they sat in and added two more of them. The
layout bug was real; fixing it was beside the point.

`adminDashboardStats` now has no consumer in the frontend at all. The
`AdminHomeStats` query and the `canViewGlobalStats` permission added for it are
both removed. The query split it motivated still stands on `feedback` alone,
which is also `@requireRole(["admin","superadmin"])` — so the project-admin fix
survives the removal.

**The "ingen aktive utfordringer" warning is gone.** I had called it "the
strongest alarm available" and "the best signal on the page". It is not a signal
at all: challenges are **episodic**, opened at specific points in a bible study
to introduce interactivity, so a running project with none active is the normal
state. The alarm would have fired on healthy projects more or less permanently.
The count stays as plain information, with no warning colour.

Recorded as two scope decisions, because both generalise beyond this page: no
global all-time counters anywhere in admin, and no health signal built on the
_absence_ of challenge activity. The surviving lifecycle detectors in item 2 are
the ones that flag a **contradiction** (published but never visible; ended but
still active) rather than an absence — that distinction is the whole reason they
are still worth building.

An unexplained observation, noted rather than chased since the card is deleted:
`adminDashboardStats.activeProjectsCount` rendered **0** while the page's own
selection logic found one active project. The backend's definition of "active"
disagrees with `startDate <= now <= endDate`. Worth knowing if that field is
ever used again.

### 2026-09-21 — home dashboard: bugs fixed, project block shipped (item 1)

Gate: typecheck 0, lint 0 errors, **577 unit** (560 → +17) and 151 component
tests green, `pnpm build` exit 0.

**Four bugs fixed**, one more than the plan listed:

1. Upcoming projects are no longer discarded. The selection logic is now a pure
   function, `utils/homeProjects.ts`.
2. `AdminDashboardStats` is 4 cards in a container-query grid and renders all
   six fields the schema exposes, instead of 2 cards in a `lg:grid-cols-4` that
   left half the row empty.
3. The page no longer breaks for a `project_admin`. **One query per concern**
   instead of one per page: `adminDashboardStats` and `feedback` are _both_
   `@requireRole(["admin","superadmin"])`, so the single document had two fields
   a project admin cannot read — I had only found one when writing the plan.
   Each is now its own document, paused behind a permission flag.
4. **New:** `AdminProjectCard` reads `branding.logoImage?.url`, but the home page
   _and_ the projects list both queried the deprecated `branding.logo`. Project
   logos have therefore never rendered on any card, silently. Both queries now
   ask for `logoImage { url }`.

**Cross-cutting #6 fixed at the same time** — `archived: false` added to all
three project queries (home, list, switcher), since the "active projects" set
this page is built on was including archived projects.

**What shipped for item 1:** `AdminProjectVitals.vue`, one self-contained card
per active project — `@container` on its own root, so one card at full width and
three sharing a row both come out right and the page decides only the grid.
Active projects are ordered by **ending soonest**, capped at three with an
"N flere aktive" link, and filtered through `canViewProject` so a project admin
sees only theirs. Zero active falls back to the next upcoming project with its
countdown and a readiness check (challenges / achievements / teams, zero
highlighted). `activeChallengesCount === 0` while running renders as a warning
on the card.

Also landed early from item 2, because the data was already in hand: the
feedback panel now filters `handled: false`, shows a total count, and renders
each entry's **tags** as badges.

**A constraint the plan got wrong, worth knowing before item 2.** The note
listed participants and teams as queryable per project. They are queryable for
_one_ project via the root `users`/`teams` connections, but **not for N projects
in a single document** — `Project` has no participant or team count field, and
GraphQL has no dynamic aliasing. So the vitals card ships with
`activeChallengesCount` (a real `Project` field) plus dates and countdown, and
the readiness check uses root counts because it only ever runs for one project.
Participants and teams need `Project.participantCount` / `teamCount` on the
backend; a tile with a placeholder in it would have been worse than no tile, so
there is none.

`canViewGlobalStats` was added to `usePermissions` mirroring
`adminDashboardStats`'s own directive, with the reason in the doc comment — the
gate has to exist client-side because the directive errors rather than returning
null.

**Tests:** `test/unit/homeProjects.test.ts`, 17 cases over the two pure modules
— ordering, tie-break stability, the cap and overflow count, and the day-boundary
cases (`describeProjectTiming` compares at day granularity so a camp ending at
23:59 today still reads as running, not ended). Mutation-checked rather than
assumed: reinstating the original discard-upcoming bug fails exactly two cases
and nothing else.

### 2026-09-21 — home plan reworked for N active projects; events ruled out

Two corrections from the user, both of which changed the design rather than
just annotating it.

**Multiple projects can be active at once**, so the "hero project" framing was
wrong. The block is now designed for `N`: vitals cards ordered by ending
soonest, capped at three plus "N more", role-scoped via the existing
`getProjectAdminIds`, falling back to the next upcoming project when none are
active. One self-contained `AdminProjectVitals` sized by its container covers
every count — the first real use of the Conventions rule, which is why that rule
landing first was useful.

**Project events are rarely used**, which killed the event-timeline idea I had
just proposed as the answer to the multi-project problem. Recorded as a scope
decision, since it is easy to design around by accident. I should have weighted
this from the note's own evidence: the restructure recorded the `events/` subtree
as _orphaned_, with nothing linking to `events/new` or `events/[eventId]`. Low
usage was the obvious reading and I proposed building a headline feature on it
anyway.

The replacement is better than what it replaced. If events are unused then **the
project is the camp**, so project dates are the real camp dates and the
current/next block needs no intermediate layer; and the fine-grained "what is
happening now" signal comes from **challenges**, which are the core unit and
carry `publishedAt` / `visibleAt` / `startedAt` / `endTime`. That yields the best
alarm on the page for free: `activeChallengesCount === 0` while a project is
running means participants have nothing to do, and nothing surfaces it today.

**A third bug found while checking the multi-project case** — archived projects
are filtered nowhere (new cross-cutting #6). The SQL is
`archived IS NULL OR archived = $x`, so omitting the filter _includes_ them, and
none of the three project-listing sites passes it. It is cross-cutting rather
than home-specific, but it matters here first: the "active projects" set this
page is built on is currently wrong.

Also folded in: feedback **tag breakdown** instead of five truncated messages
(`feedbackTags` exists), and **`translationStatus` on Project and Challenge** as
a detector — the user app is translated even though admin is not, so a missing
translation is a real end-user defect that nothing surfaces.

Rejected ideas are recorded with their reasons in the plan, so they do not get
re-proposed: the event timeline, the join/team funnel (no `Team` member count
field), and push-send as a quick action (needs a confirm step).

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
