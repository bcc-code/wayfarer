# Admin UX improvements — working note

Per-page tracker for the admin UX pass. The **structural** work (layers,
project-scoped IA, tabs → routes, shared shell, breadcrumbs, permission
consolidation) is finished and written up in
[`frontend-admin-restructure.md`](./frontend-admin-restructure.md) — read the
"Watch out for" section there before touching any project-scoped page. This note
is about what happens *inside* the pages.

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

One real bug survives this decision and is *not* an i18n task:
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

---

## Cross-cutting work

Ordered by priority. These are sweeps — do them once across every affected page
rather than page by page.

### 1. Replace hardcoded query limits with pagination — 12 pages

Every list below renders at most N rows, with no pagination and no indication
that anything was cut off. `projects/[projectId]/index.vue`'s old mega-query had
exactly this bug and lost its `first: 50` during the restructure; these are the
remainder.

| Page | Limit | Shape |
| --- | --- | --- |
| `my-church/units.vue` | `first: 1000` | list, also an outlier (see #4) |
| `my-church/admins.vue` | `first: 500` | list + client fuzzy search |
| `projects/[projectId]/superteams/[superTeamId].vue` | `first: 200` (teams) | picker |
| `users/[userId]/achievements.vue` | `first: 200` | picker |
| `projects/index.vue` | `first: 100` | card grid |
| `users/[userId]/index.vue` | `first: 100` (feedback) | panel in a detail page |
| `projects/[projectId]/challenges/new.vue` | `first: 100` (events) | dropdown |
| `projects/[projectId]/superteams/distribute.vue` | `first: 100` (events) | dropdown |
| `projects/[projectId]/challenges/index.vue` | `first: 50` | list |
| `projects/[projectId]/achievements/index.vue` | `first: 50` | list, drag-reorder |
| `projects/[projectId]/events/index.vue` | `first: 50` | list |
| `projects/[projectId]/superteams/index.vue` | `first: 50` | list |
| `maintenance/check-points-journal.vue` | `first: 50` | table |
| `maintenance/fix-content-progress.vue` | `first: 50` | table |

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

| Shape | Loading | Empty | Error |
| --- | --- | --- | --- |
| List / table | skeleton rows (not just `UTable :loading`) | "nothing here yet" + the primary create action | `AdminErrorState` + retry |
| Detail | skeleton matching the layout | n/a | `AdminErrorState`; 404 distinct from failure |
| Form (`new.vue`) | only if it loads options | n/a | inline field errors + submit failure |

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

---

## Pages

### Top level

| Page | Route | LOC | St. | Notes |
| --- | --- | --- | --- | --- |
| `index.vue` | `/admin` | 134 | ☐ | Landing. No breadcrumb by design (one crumb = a title). `feedback(first: 5)`. |
| `projects/index.vue` | `/admin/projects` | 119 | ☐ | Card grid, container queries done. No search; `first: 100`. |
| `projects/new.vue` | `/admin/projects/new` | 167 | ☐ | |
| `users/index.vue` | `/admin/users` | 129 | ☐ | Best-in-class list: debounced server search + relay pagination. Use as the reference pattern. |
| `users/[userId]/index.vue` | `/admin/users/:userId` | **1063** | ☐ | Outlier, see #4. Feedback panel `first: 100`. |
| `users/[userId]/achievements.vue` | `…/achievements` | 389 | ☐ | `first: 200` picker — see #1, wants a searchable select, not pagination. |
| `churches/[churchId].vue` | `/admin/churches/:churchId` | 175 | ☐ | Drill-down leaf by design — no list page, no nav entry (decision recorded in restructure note). |
| `consents/index.vue` | `/admin/consents` | 109 | ☐ | Table, no search, no pagination, no loading state. |
| `consents/[consentId].vue` | `/admin/consents/:id` | 305 | ☐ | |
| `consents/new.vue` | `/admin/consents/new` | 187 | ☐ | |
| `feedback/index.vue` | `/admin/feedback` | 528 | ☐ | Three filter facets + pagination; realtime via Firestore. Largest list page. |

### Project-scoped (`/admin/projects/:projectId/…`)

| Page | Route | LOC | St. | Notes |
| --- | --- | --- | --- | --- |
| `[projectId].vue` | (parent shell) | 21 | – | Deliberately thin; project lives in `useCurrentProject()`. Do not delete `index.vue` beneath it — blank-page trap. |
| `[projectId]/index.vue` | `…/:projectId` | 143 | ☐ | Overview: counts-only query + section links. Redirects legacy `?tab=`. |
| `[projectId]/edit.vue` | `…/edit` | 297 | ☐ | "Innstillinger". Only route a `project_admin` can actually use. |
| `challenges/index.vue` | `…/challenges` | 119 | ☐ | `first: 50`, no search. |
| `challenges/new.vue` | `…/challenges/new` | 171 | ☐ | Events dropdown `first: 100`. |
| `challenges/[challengeId]/index.vue` | `…/challenges/:id` | 222 | ☐ | |
| `challenges/[challengeId]/quiz.vue` | `…/:id/quiz` | 417 | ☐ | Two-crumb page. |
| `challenges/[challengeId]/sessions.vue` | `…/:id/sessions` | 602 | ☐ | Two-crumb page. 13 toast calls — likely the noisiest page in the app. |
| `achievements/index.vue` | `…/achievements` | 169 | ☐ | Drag-reorder. `first: 50`. |
| `achievements/new.vue` | `…/achievements/new` | 134 | ☐ | |
| `achievements/[achievementId].vue` | `…/achievements/:id` | 324 | ☐ | |
| `events/index.vue` | `…/events` | 94 | ☐ | Created during restructure to un-orphan the two below. `first: 50`. |
| `events/new.vue` | `…/events/new` | 94 | ☐ | |
| `events/[eventId].vue` | `…/events/:id` | 188 | ☐ | |
| `superteams/index.vue` | `…/superteams` | 132 | ☐ | The real list (was a tab). `first: 50`. |
| `superteams/new.vue` | `…/superteams/new` | 103 | ☐ | |
| `superteams/[superTeamId].vue` | `…/superteams/:id` | 278 | ☐ | Teams `first: 200`. |
| `superteams/distribute.vue` | `…/superteams/distribute` | 657 | ☐ | Ladder-to-heaven tool. `@unovis/vue` charts, raw `fetch` to two plugin endpoints — not GraphQL. |
| `teams/index.vue` | `…/teams` | 132 | ☐ | Paginated. No search. |
| `teams/[teamId].vue` | `…/teams/:id` | 428 | ☐ | 14 toast calls. |
| `scores/index.vue` | `…/scores` | 233 | ☐ | Paginated. No search/date filter. |
| `scores/new.vue` | `…/scores/new` | 131 | ☐ | Project picker removed — route supplies it. |

### My church (`church-admin` layout — no navigation, see #5)

| Page | Route | LOC | St. | Notes |
| --- | --- | --- | --- | --- |
| `my-church/index.vue` | `/admin/my-church` | 77 | ☐ | |
| `my-church/units.vue` | `…/units` | **1109** | ☐ | Outlier, see #4. Worst limit in the codebase (`first: 1000`). |
| `my-church/admins.vue` | `…/admins` | 354 | ☐ | Client-side fuzzy search. `first: 500`. |
| `my-church/statistics.vue` | `…/statistics` | 194 | ☐ | |
| `my-church/kickoff.vue` | `…/kickoff` | 166 | ☐ | |
| `my-church/gamenights/index.vue` | `…/gamenights` | 59 | ☐ | |
| `my-church/gamenights/gamenight-[gamenight].vue` | `…/gamenights/:n` | 642 | ☐ | Largest my-church page. |

### Maintenance

| Page | Route | LOC | St. | Notes |
| --- | --- | --- | --- | --- |
| `maintenance/index.vue` | `/admin/maintenance` | 66 | ☐ | Static tool list. |
| `maintenance/bulk-jobs.vue` | `…/bulk-jobs` | 381 | ☐ | Paginated. |
| `maintenance/check-points-journal.vue` | `…/check-points-journal` | 149 | ☐ | Hardcoded `ACHIEVEMENT_ID`. `first: 50`. |
| `maintenance/fix-content-progress.vue` | `…/fix-content-progress` | 367 | ☐ | Job polling — was broken pre-restructure, now typed. |
| `maintenance/fix-streak-progress.vue` | `…/fix-streak-progress` | 350 | ☐ | Same shape as above. |

### Legacy redirect stubs (not UX surfaces)

| Page | Route | LOC | St. | Notes |
| --- | --- | --- | --- | --- |
| `teams/index.vue` | `/admin/teams` | 11 | – | Static redirect → project picker. |
| `teams/[teamId].vue` | `/admin/teams/:teamId` | 61 | – | Async: looks up `Team.parentProject` then redirects. Has an explanatory empty state. |
| `scores/index.vue` | `/admin/scores` | 11 | – | Static redirect. |
| `scores/new.vue` | `/admin/scores/new` | 11 | – | Static redirect. |

---

## Update log

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
