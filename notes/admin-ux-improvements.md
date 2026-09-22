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

## Where this stands (2026-09-21)

| #   | Item                            | State                                                                                                          |
| --- | ------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Hardcoded query limits          | 1 done, 3 deliberately left, 10 open — 4 of those are pickers/dropdowns wanting a searchable select, not pages |
| 2   | Loading / empty / error states  | **done** — 13 tables + 17 pages                                                                                |
| 3   | Search, filter, sort            | filtering on 6 lists; **sorting untouched**                                                                    |
| 4   | The 1,000-line outliers         | half — `users/[userId]` split into 6 components; `my-church/units.vue` remains                                 |
| 5   | `church-admin` navigation       | **untouched** — the biggest single-surface gap left                                                            |
| 6   | Archived projects               | **done**                                                                                                       |
| 7   | Container queries in components | **done**                                                                                                       |

Shared machinery built along the way, all with tests:
`AdminListView` + `useListState` (toolbar, footer, URL state, pagination reset),
`AdminQueryState` (loading/error/content, refetch-safe), `AdminTableEmpty` /
`AdminTableLoading`, `AdminSection` (heading + raised surface, replaces `UCard`
for page content), `AdminSparkline`. `RelayPagination` was folded into
`AdminListView` and deleted.

**`AdminSection` sweep: done 2026-09-21.** Four detail pages converted —
`churches/[churchId]`, `consents/[consentId]`,
`projects/[projectId]/teams/[teamId]`, `users/[userId]/achievements` — plus the
six `AdminUser*` panels. Nine admin pages _contain_ `UCard`; only these four
were using it as a section container.

**Four pages keep `UCard`, correctly** — they are discrete objects on a surface,
which is what cards are still for: `maintenance/index.vue` (clickable cards in a
grid), three stat tiles each in `check-points-journal` / `fix-content-progress` /
`fix-streak-progress`, and the warning callout in the two `fix-*` tools (it has a
`#header` but is an alert, not a page section). `projects/[projectId]/index.vue`
left the list on 2026-09-22 when its tiles became plain links inside
`AdminSection`.

Three things the conversion needed that the component did not have:

- **No `#footer` slot.** Two edit forms put their Avbryt/Lagre buttons there;
  they now sit at the end of the section body. A footer slot would have been the
  lazier fix — the buttons belong _to the form_, not to the section chrome.
- **A plain-string title.** The achievements page's card header held a title, a
  type badge, a tally and an awarded badge. Everything but the title moved to
  `#actions`, which keeps the same title-left / meta-right arrangement.
- **A title where there had been none.** `churches/[churchId]`'s field list was
  a headerless card; it is "Detaljer" now. Its ULID also moved from _first_ row
  to last, matching the user detail page — a support aid, not the first thing a
  reader wants.

Rows inside converted sections lost their own borders (`border rounded-md p-3`
→ `py-2`): they sit _on_ the section surface rather than in boxes within it,
which was the original complaint.

Pages given real work: `/admin` (rebuilt), `users/index`, `users/[userId]`
(+ 6 components), `feedback`, `challenges`, `teams`, `scores`,
`maintenance/bulk-jobs`.

**If you are picking this up:** #5 is the obvious next move — seven
`my-church/**` pages on a layout with no navigation at all, and it brings
`units.vue` (the last #4 outlier) with it.

### Unverified by the author, needs a human

Written down because the admin panel requires an Auth0 session, so none of it
could be exercised from the terminal:

- **The user detail page's mutations.** Add/revoke a role, withdraw a consent,
  sync, church lock. During the six-component split these moved from direct
  `refetch()` calls to emitted `changed` events, so the wiring is new even
  though the handlers are not.
- **The sparklines against real data.** Every screenshot of them was a local
  stub. `Project.activityTrend` and the gap-filling are unit-tested, but the
  path from SQL through the resolver into the chart has only run against a
  project with _no_ recent activity, where the component takes the empty branch
  and the plot is never drawn.
- **Light mode generally.** It was checked for the section surfaces after the
  fact, and not at all for the home dashboard, the list toolbars or the
  sparkline colours.

### Known open risks

- **`achievements/index.vue` truncates silently at `first: 50`** and is the
  highest such risk left, since reading achievements multiply with articles.
  Cause and preferred fix (a dedicated reorder mode) are in #1.
- **`adminDashboardStats.activeProjectsCount` disagreed with the frontend's own
  "is this project running" logic** — it rendered 0 while a project was active.
  The card using it was deleted, so this is latent rather than visible.
- **`schema.sql` documents 46 of 83 tables.** See `notes/04-database.md`; the
  fix is generating it, not hand-patching.
- **`setLocale('nb')` in `layouts/admin.vue`** permanently switches a
  non-Norwegian end user's _consumer_ app language. Survives the no-i18n
  decision because it is a consumer-app bug.

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

### Comments: only what the code cannot say

Write a comment when a reader would otherwise get it wrong — a non-obvious API
contract, a constraint that looks like it could be simplified away, a value
chosen for a reason. Not for narrating what the next line does, and not for
history or rationale: that belongs here, in this note.

One or two lines. If it needs a paragraph, it is probably a note entry.

Examples worth keeping, because removing them invites a bug:

- `UProgress`'s `role="progressbar"` sits on an inner element, so the
  accessible name must come from `get-value-label`, not an `aria-label`.
- `useAdminPage`'s `flush: 'post'` — without it, a page's label getter runs
  during setup and 500s on a TDZ.
- `sr-only` must go on a wrapper, not a `<table>`, or the `<caption>` escapes.

What came out on 2026-09-22: a comment-density pass over everything written
that week took `AdminSection` from **54% comment lines to 9%**, and the shared
components and composables to **7% overall**. Most of what went was rationale
already written down here — duplicated, and in the way of the code.

### Raised surfaces are white in light mode, tinted in dark

The admin shell sets the page ground to `bg-neutral-100 dark:bg-neutral-950`
(`layouts/admin.vue`) — deliberately **one step behind the surfaces in both
modes**, a decision from the restructure work. So "raised" means the _opposite
direction_ per mode:

| Mode  | Page ground                | A raised surface             |
| ----- | -------------------------- | ---------------------------- |
| light | `neutral-100` (grey)       | `bg-default` — **white**     |
| dark  | `neutral-950` (near black) | `bg-elevated` — lighter grey |

A grey tint on light is therefore _darker_ than the page and reads as **sunk
into** it rather than lifted off it. `bg-elevated/40` alone looks right in dark
and wrong in light, which is exactly the mistake `AdminSection` made — see the
log.

**Check both modes before judging any surface treatment.** Nothing about
"a subtle background" is mode-independent when the ground sits behind the
surfaces by design.

---

## Cross-cutting work

Ordered by priority. These are sweeps — do them once across every affected page
rather than page by page.

### 1. Replace hardcoded query limits with pagination — 13 sites, 1 done

Every list below renders at most N rows, with no pagination and no indication
that anything was cut off. `projects/[projectId]/index.vue`'s old mega-query had
exactly this bug and lost its `first: 50` during the restructure; these are the
remainder.

**Not all of them should be paginated**, which only became clear once the
component existed. A toolbar and a pagination footer over five rows is worse
than a plain grid, and a picker or dropdown wants a searchable select rather
than pages. The `Verdict` column records the call per site so it is not
re-litigated.

| Page                                                | Limit                       | Verdict                                                                                 |
| --------------------------------------------------- | --------------------------- | --------------------------------------------------------------------------------------- |
| `projects/[projectId]/challenges/index.vue`         | ~~`first: 50`~~             | ☑ **done** — paginated + type filter                                                    |
| `projects/[projectId]/achievements/index.vue`       | `first: 50`                 | **left alone** — reorder does not compose with paging; **highest truncation risk left** |
| `projects/[projectId]/superteams/index.vue`         | `first: 50`                 | **left** — no categorical facet in `SuperTeamFilter`, ~5 per project                    |
| `projects/[projectId]/events/index.vue`             | `first: 50`                 | **left** — as above, plus events are barely used                                        |
| `my-church/units.vue`                               | `first: 1000`               | open — also an outlier (#4), on the `church-admin` layout                               |
| `my-church/admins.vue`                              | `first: 500`                | open — has client-side fuzzy search                                                     |
| `projects/[projectId]/superteams/[superTeamId].vue` | `first: 200` (teams)        | open — **picker**; wants a searchable select, not pages                                 |
| `users/[userId]/achievements.vue`                   | `first: 200`                | open — **picker**, as above                                                             |
| `projects/index.vue`                                | `first: 100`                | open — card grid                                                                        |
| `users/[userId]/index.vue`                          | ~~`first: 100`~~ (feedback) | ☑ **done** — `first: 10`, matching the panel's own notice                               |
| `projects/[projectId]/challenges/new.vue`           | `first: 100` (events)       | open — **dropdown**                                                                     |
| `projects/[projectId]/superteams/distribute.vue`    | `first: 100` (events)       | open — **dropdown**                                                                     |
| `maintenance/check-points-journal.vue`              | `first: 50`                 | open — table                                                                            |
| `maintenance/fix-content-progress.vue`              | `first: 50`                 | open — table                                                                            |

**`AdminListView` + `useListState` are the tool** — see the 2026-09-21 log
entries. `RelayPagination` was folded into them and deleted. Five pages are on
it: `users`, `feedback`, `challenges`, `projects/[projectId]/{teams,scores}` and
`maintenance/bulk-jobs`; `users/index.vue` is the reference, with a debounced
server-side filter, a facet and full URL state.

**Not every row here wants the same fix.** Three shapes, three answers:

- **Lists** — `AdminListView`, following `users/index.vue`.
- **Dropdowns and pickers** (`challenges/new.vue`, `distribute.vue`,
  `users/[userId]/achievements.vue`, `superteams/[superTeamId].vue`) — a paginated
  dropdown is worse UX, not better. These want a searchable select backed by a
  server-side query, or a justified limit with an explicit "showing first N"
  affordance.
- **`achievements/index.vue` has drag-to-reorder**, which does not compose with
  pagination — reordering across a page boundary has no meaning. **Left alone
  2026-09-21**; when picked up, prefer a dedicated reorder mode that loads
  everything over page-local ordering, which is cheaper and semantically wrong.

### 2. Loading, empty and error states everywhere they belong — ☑ done 2026-09-21

Make the three states universal and consistent. `AdminErrorState` is on 30
pages, `AdminLoadingState`/`USkeleton` on 22; empty states are ad hoc.

Decide one house pattern per page shape first, then apply it — auditing 44 pages
individually is the expensive way to do this:

| Shape            | Loading                                    | Empty                                          | Error                                        |
| ---------------- | ------------------------------------------ | ---------------------------------------------- | -------------------------------------------- |
| List / table     | skeleton rows (not just `UTable :loading`) | "nothing here yet" + the primary create action | `AdminErrorState` + retry                    |
| Detail           | skeleton matching the layout               | n/a                                            | `AdminErrorState`; 404 distinct from failure |
| Form (`new.vue`) | only if it loads options                   | n/a                                            | inline field errors + submit failure         |

**List half: done 2026-09-21.** All 13 admin tables now have `#empty`
(`AdminTableEmpty`) and `#loading` (`AdminTableLoading`) — see the log entry.
`UTable`'s `#loading` slot turned out to exist and be unused everywhere, so
skeleton rows cost nothing.

**Detail and form pages: done 2026-09-21.** `AdminQueryState` on all 17 pages
that had the chain — see the log entry. The 13 pages that render only
`<AdminErrorState v-if="error" />` are correct as they are: they are lists,
whose loading and empty states now live in the table's slots.

Do not "deduplicate" `AdminErrorState`/`AdminLoadingState` back into the shared
`ErrorState`/`LoadingState`: keeping them separate is what severs admin's
dependency on the user design system (see the restructure note).

### 3. Search, filter and sort on lists — filtering largely done; sorting open

**Filtering: six lists now have it**, each taking a facet from what its own
filter input already supported (2026-09-21):

| List                        | Filter                                 |
| --------------------------- | -------------------------------------- |
| `users/index.vue`           | debounced server-side `query` + church |
| `feedback/index.vue`        | tags, platform, handled (tri-state)    |
| `challenges/index.vue`      | challenge type                         |
| `teams/index.vue`           | superteam, incl. "uten superlag"       |
| `scores/index.vue`          | source type                            |
| `maintenance/bulk-jobs.vue` | status, operation type                 |

All six are on `AdminListView` + `useListState`, so their state is URL-synced.
`my-church/admins.vue` keeps its own client-side fuzzy search (different
layout, unconverted).

**Still unfiltered:** `consents` (plain list, not a connection — see the log),
`achievements`, `events`, `superteams` (no categorical facet worth a toolbar,
handful of rows each), and the maintenance preview tables.

**Sorting: still none, anywhere.** All `UTable` pages pass a static `:columns`
and nothing in `gql/` takes a sort argument for these entities. It is **not**
blocked on the cursor decision: sorting needs the cursor to encode the sort key
plus a stable tiebreaker — `(name, id)` rather than `id` alone — which widens
the cursor rather than requiring `OFFSET`. See Scope decisions.

### 4. Refactor the two outliers — half done

`users/[userId]/index.vue` is **done** (2026-09-21): 1,182 → 240 lines across six
`AdminUser*` components, each owning its own mutations and modals.
`my-church/units.vue` (1,109 lines) remains — and note it is on the
`church-admin` layout, so it lands naturally with #5.

Originally: `my-church/units.vue` (1,109 lines) and `users/[userId]/index.vue`
(1,063).
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

### 7. Bring existing components onto the self-contained rule — ☑ done 2026-09-21

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
2. ~~`Challenge.completionCount` / `Achievement.awardedUserCount`~~ — **done
   2026-09-22.** A project-level live-challenges field is still open, since
   `activeChallenges` is viewer-relative and unusable in admin — but it may not
   be needed: `Challenge` already exposes `publishedAt` / `visibleAt` /
   `startedAt` / `endTime`, so the frontend can derive what is live.
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

| Page                              | Route                       | LOC | St. | Notes                                                                                                         |
| --------------------------------- | --------------------------- | --- | --- | ------------------------------------------------------------------------------------------------------------- |
| `index.vue`                       | `/admin`                    | 134 | ◐   | **Plan agreed — see Page plans.** Three bugs to fix first. No breadcrumb by design (one crumb = a title).     |
| `projects/index.vue`              | `/admin/projects`           | 119 | ☐   | Card grid, container queries done. No search; `first: 100`.                                                   |
| `projects/new.vue`                | `/admin/projects/new`       | 167 | ☐   |                                                                                                               |
| `users/index.vue`                 | `/admin/users`              | 199 | ◐   | **On `AdminListView`** — the reference conversion. Search + church filter + URL state.                        |
| `users/[userId]/index.vue`        | `/admin/users/:userId`      | 240 | ◐   | Split into 6 `AdminUser*` components; identity header; 4 bugs fixed. No longer an outlier.                    |
| `users/[userId]/achievements.vue` | `…/achievements`            | 389 | ☐   | `first: 200` picker — see #1, wants a searchable select, not pagination.                                      |
| `churches/[churchId].vue`         | `/admin/churches/:churchId` | 175 | ☐   | Drill-down leaf by design — no list page, no nav entry (decision recorded in restructure note).               |
| `consents/index.vue`              | `/admin/consents`           | 109 | ☐   | Plain list (`[Consent!]!`), not a connection — off `AdminListView` by decision, see log. Sibling empty state. |
| `consents/[consentId].vue`        | `/admin/consents/:id`       | 305 | ☐   |                                                                                                               |
| `consents/new.vue`                | `/admin/consents/new`       | 187 | ☐   |                                                                                                               |
| `feedback/index.vue`              | `/admin/feedback`           | 601 | ◐   | **On `AdminListView`.** Three facets + URL state; realtime via Firestore. Largest list page.                  |

### Project-scoped (`/admin/projects/:projectId/…`)

| Page                                    | Route                     | LOC | St. | Notes                                                                                                                            |
| --------------------------------------- | ------------------------- | --- | --- | -------------------------------------------------------------------------------------------------------------------------------- |
| `[projectId].vue`                       | (parent shell)            | 21  | –   | Deliberately thin; project lives in `useCurrentProject()`. Do not delete `index.vue` beneath it — blank-page trap.               |
| `[projectId]/index.vue`                 | `…/:projectId`            | 217 | ◐   | **Dashboard.** Status (participants + 14-day trend) and Innhold (section links + counts). `max-w-6xl`. Engagement needs backend. |
| `[projectId]/edit.vue`                  | `…/edit`                  | 297 | ☐   | "Innstillinger". Only route a `project_admin` can actually use.                                                                  |
| `challenges/index.vue`                  | `…/challenges`            | 227 | ◐   | **On `AdminListView`** + type filter; `first: 50` replaced by real pagination.                                                   |
| `challenges/new.vue`                    | `…/challenges/new`        | 171 | ☐   | Events dropdown `first: 100`.                                                                                                    |
| `challenges/[challengeId]/index.vue`    | `…/challenges/:id`        | 222 | ☐   |                                                                                                                                  |
| `challenges/[challengeId]/quiz.vue`     | `…/:id/quiz`              | 417 | ☐   | Two-crumb page.                                                                                                                  |
| `challenges/[challengeId]/sessions.vue` | `…/:id/sessions`          | 602 | ☐   | Two-crumb page. 13 toast calls — likely the noisiest page in the app.                                                            |
| `achievements/index.vue`                | `…/achievements`          | 169 | ☐   | Drag-reorder blocks paging — left alone by decision; **highest truncation risk left**. `first: 50`.                              |
| `achievements/new.vue`                  | `…/achievements/new`      | 134 | ☐   |                                                                                                                                  |
| `achievements/[achievementId].vue`      | `…/achievements/:id`      | 324 | ☐   |                                                                                                                                  |
| `events/index.vue`                      | `…/events`                | 94  | ☐   | **Low priority** — events rarely used, see Scope decisions. Created during restructure to un-orphan the two below. `first: 50`.  |
| `events/new.vue`                        | `…/events/new`            | 94  | ☐   | **Low priority** — events rarely used.                                                                                           |
| `events/[eventId].vue`                  | `…/events/:id`            | 188 | ☐   | **Low priority** — events rarely used.                                                                                           |
| `superteams/index.vue`                  | `…/superteams`            | 132 | ☐   | The real list (was a tab). `first: 50`.                                                                                          |
| `superteams/new.vue`                    | `…/superteams/new`        | 103 | ☐   |                                                                                                                                  |
| `superteams/[superTeamId].vue`          | `…/superteams/:id`        | 278 | ☐   | Teams `first: 200`.                                                                                                              |
| `superteams/distribute.vue`             | `…/superteams/distribute` | 657 | ☐   | Ladder-to-heaven tool. `@unovis/vue` charts, raw `fetch` to two plugin endpoints — not GraphQL.                                  |
| `teams/index.vue`                       | `…/teams`                 | 231 | ◐   | **On `AdminListView`** + superteam filter. Slot-name bug fixed.                                                                  |
| `teams/[teamId].vue`                    | `…/teams/:id`             | 428 | ☐   | 14 toast calls.                                                                                                                  |
| `scores/index.vue`                      | `…/scores`                | 271 | ◐   | **On `AdminListView`** + source-type filter.                                                                                     |
| `scores/new.vue`                        | `…/scores/new`            | 131 | ☐   | Project picker removed — route supplies it.                                                                                      |

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

### 2026-09-22 — `Challenge.completionCount` and `Achievement.awardedUserCount`

The engagement numbers the bible-study creators asked for. Both are
viewer-independent, unlike `userCompletedAt` / `achievedAt`.

**Backend.** Added to the `Challenge` and `Achievement` interfaces and all four
implementations of each, `forceResolver` on every one. Two grouped bulk queries
(`GetBulkChallengeCompletionCounts`, `GetBulkAchievementAwardedUserCounts`) feed
two new dataloaders, `ChallengeCompletionCountLoader` and
`AchievementAwardedUserCountLoader` — a list of 50 challenges costs one query.

Two decisions worth keeping:

- **Not cached in Ristretto.** These change on every completion or award and
  have no invalidation path, so they are batched per request window only — the
  same call `userProjectScoreBatchFunc` already makes for scores.
- **`awardedUserCount`, not `awardedCount`.** `team_achievements` and
  `super_team_achievements` exist, so an unqualified name would repeat the
  `activeChallenges` mistake: a number whose meaning you have to read the
  resolver to learn. The name says whose awards it counts. Challenges have no
  team equivalent, so `completionCount` needs no qualifier.

**Frontend.** Three surfaces, all sharing `AdminEngagementCount` (count, "av N ·
x %", optional proportion bar):

- Project overview: a new **Engasjement** section (`AdminProjectEngagement`,
  owns its own query), each list sorted by count descending so the tail answers
  "what is nobody doing".
- Challenges list: a **Fullført** column.
- Achievements list: the award count per row.

The share needs a denominator, so the two list pages now also ask for
`participants: users(first: 0, filter: { projectId })`.

**Second pass, same day: the panel was too tall.** Two side-by-side vertical
lists of 10 and 12 rows, each repeating "av 88 · 3 %", filled a screen. Now:

- The denominator is stated once for the panel, not per row.
- **Achievements come first** and are a wrapping grid of badges — the
  achievement's own image inside a ring whose arc is the share of participants
  (`AdminEngagementRing`), count underneath. 12 achievements fit in two rows
  instead of twelve.

  Three things the first attempt got wrong: the track was `text-elevated`, which
  is not a text utility, so it inherited body colour and drew a near-white ring
  the pink arc vanished against (now `text-dimmed` at 40 %); the image sat one
  pixel off the track (now inset by `STROKE + 4`, inside its own
  `overflow-hidden` disc so any aspect ratio still clips); and
  `/images/achievement-placeholder.png` is a **white disc**, which swallowed the
  ring entirely — an achievement without an image now gets a muted award icon
  from the component instead.

- **Challenges stay named rows**, but dense and in two columns: name, count, a
  12-unit bar. Their images do not identify them the way an achievement badge
  does ("Game Night 1 - Tipping" vs "Game Night 1 - Unit oppgave"), so a badge
  grid would have been unreadable.
- Name and exact share moved into a `UTooltip` per row or badge.

`UTooltip` teleports and needs `UApp`'s provider, so component tests stub it
with a `data-tooltip` attribute rather than asserting on rendered tooltip text.

`AdminEngagementCount` (count + "av N · x %" + optional bar) stays as it is for
the two list pages, where a table row has the width for it.

Tests: two loader mapping tests in `internal/loaders/`, and component tests for
`AdminEngagementCount` (share maths, no total, zero participants, the bar capped
at 100 % — awards can outnumber current participants) and
`AdminProjectEngagement` (ordering, per-list empty state).

Not done: no e2e test, which would need Docker.

### 2026-09-22 — project overview becomes a dashboard (no-backend half)

`/admin/projects/:projectId` was a header plus five tab-replacement links. It is
now two sections:

- **Status** — participant count and the 14-day activity trend
  (`AdminActivityTrend`), plus a timing badge in the header.
- **Innhold** — the five section links, each with its count.

Capped at `max-w-6xl`, like the home and user detail pages.

**`AdminActivityTrend` extracted.** The two-tile trend (points summed, active
users averaged) lived inside `AdminProjectSection`. It is now its own component
used by both the home section and this page; `AdminProjectSection` lost its
duplicated `trendTiles`. Own tests in `test/component/AdminActivityTrend.test.ts`.

**The trend is skipped unless the project is running.** Same `@include(if:)` rule
as the home section: a finished or unstarted project's last 14 days are all
zeroes, which reads as a broken chart rather than as "nothing is running". A line
of text says which it is instead.

#### `activeChallenges` is viewer-relative — do not use it in admin

The planned "Åpne nå" section was removed after reading
`backend/internal/graph/api/challenges.go:25`: `challengeFilterActive` means _not
completed by the viewer_ and not past its end time, and the resolver filters
through `UserChallengeCompletionTimestampLoader`. So `activeChallenges` and
`activeChallengesCount` answer "what can **I** still do", not "what is live in
this project" — for an admin who is also a participant they under-report, and the
number differs per admin looking at the same project. The
`activeChallengesCount` tile was removed from `AdminProjectSection` too.

It needs a project-level field — or, since `Challenge` does expose `publishedAt`,
`visibleAt`, `startedAt` and `endTime`, the frontend can derive "live right now"
from those without any backend work. (An earlier version of this entry claimed
those fields did not exist. They do.)

#### What the bible-study creators asked for still needs backend work

"How many users achieved achievements / have done challenges" is not answerable
with the current schema. Three fields, in order of value:

1. `Challenge.completionCount` — the count of `user_challenge_completions` rows.
   With it, "live and nobody has completed it" catches a broken QR code or a
   wrong date _during_ a camp.
2. `Achievement.awardedCount` — the count of `user_achievements` rows.
3. A project-level open/live-challenges field, viewer-independent, which is what
   the removed section wanted.

`Team.memberCount` remains wanted for the join/team funnel.

### 2026-09-22 — points panel: per-project totals + recent activity

The flat journal is gone. The panel now shows **per-project totals** and the
**five most recent entries**, chosen because the page is used for support cases:
totals answer "where does this person have points and how many", the recent
window answers "what just happened", which is where a support call starts.

**No cross-project total, deliberately.** Projects differ in length and scoring
scale, so a lifetime sum is a number nobody can act on — the same vanity metric
that came off the home dashboard. The per-project figure is the unit that means
something.

**What the old panel got wrong**, visible in the screenshot that prompted this:

- The heading said "0 poeng i Sommercamp 2026" above twelve rows totalling
  ~25 000 points in a _different_ project. `points(projectId:)` is
  project-scoped; `adminScoreJournal(filter: { userId })` is not.
- The project column was **identical on all twelve rows** — a third of each
  row's width spent saying nothing. It only carries information when the log
  actually spans projects, and then a flat log is unscannable anyway. In the new
  recent-activity list it does span projects, so the name is back to being
  useful there.
- 100 rows on a summary page, truncated with no way to see the rest. The full
  log now lives where it belongs: each total links to that project's score page
  filtered to this user.

**Backend: `User.pointsByProject: [UserProjectPoints!]!`.** Needed because
`points(projectId:)` takes one project at a time, so a caller wanting the whole
set would need an aliased field per project and the set is not known up front.
Deriving it client-side from the fetched journal was rejected: summing a
`last: 100` window and labelling it "total" is the lying-label problem twice
over.

Three choices in that query worth keeping:

- **Ordered by the user's latest entry per project**, not by project date — on a
  support call the project someone just scored in is the one being asked about.
- **Flat `projectId`/`projectName`** rather than a nested `Project!`, matching
  `ChurchAdminStatistics`. The aggregate needs a label and a link target, not a
  whole project with branding.
- **Projects that net to zero still appear.** A row exists because there is
  journal activity, and "0 poeng" for a project someone participated in is a
  real answer — quite different from the project being absent, which would
  suggest they were never in it. There is a test for exactly that.

`mapUserPointsByProject` is a pure function in `users.go` (not the generated
resolver file) with 5 tests, including that an empty result is a non-nil slice —
the field is `[UserProjectPoints!]!` and gqlgen marshals nil as `null`.

Gate: backend `make fmt` + `make test` green (27 packages); frontend typecheck
0, lint 0, 604 unit + 225 component, build exit 0.

### 2026-09-22 — "see all" links go to the list filtered to that user

Every outward link from the user detail page now lands somewhere about _that
user_, not on an unfiltered list.

| Link                | Was                     | Now                                 |
| ------------------- | ----------------------- | ----------------------------------- |
| Feedback "Vis alle" | the whole feedback list | `admin-feedback?userId=…`           |
| Score journal row   | plain text              | that project's journal, `?userId=…` |

Both target pages had to learn the filter: `userId` joins their `useListState`
filters so it survives a reload, resets pagination, and is clearable through a
chip like any other. Neither gets a _control_ for it — a user picker on the
feedback list would be a worse way to answer "show me this person's feedback"
than arriving from their page — so the chip is the only affordance, which is
what makes the filter discoverable once set.

**Each page resolves the user's name for the chip** through a small paused
query, falling back to the id while it loads or if the user is gone. A chip
reading `Bruker: US01KCMJRS055PF6TPYVPT3M642W` would be the same mistake as the
role scope printing a ULID.

**A bug of mine, found while doing this.** The journal panel's heading said
"Poenglogg **i Sommercamp 2026**" — but `adminScoreJournal(filter: { userId })`
has **no project filter**, so its rows span every project the user has scored
in. Only `points(projectId:)` is project-scoped. The heading was claiming a
scope the rows below it did not have, which is visible in an earlier screenshot:
rows labelled "Test project" under a heading naming a different project. The
project name moved to the points badge, where it is true.

That is the same class of error as the thing it was introduced to fix — an
unlabelled scope — and I created it while fixing that. Worth the reminder that
adding a label is only an improvement if the label is accurate.

It also explains why a single section-level "Vis alle" was impossible here, and
why the restructure had removed the old one: a cross-project log has no one
project page to point at. Per-row links are the honest shape.

### 2026-09-21 — `AdminSection` replaces `UCard` for page content (5 attempts)

The user's observation: "everything boxed in a card / box… would it be better to
have simple sections with headers instead?" Correct — and the rows inside those
cards were _themselves_ boxed, so every item wore chrome twice and five stacked
cards read as five competing panels rather than one page.

**`AdminSection`** is a heading (with optional count and `#actions`) above a
soft surface holding the content. All six `AdminUser*` panels use it; no `UCard`
remains in them.

**It took five attempts, and the record is the useful part:**

1. `UCard` per section, boxed rows inside — box chrome twice over.
2. A rule _above_ each heading plus a divider per row — **~35 horizontal lines**
   on a user with a full points journal. I had replaced box chrome with line
   chrome.
3. No lines at all — defensible in dark mode, **completely flat in light**,
   where everything sits on white and nothing groups a section.
4. A hairline under each heading — better, still thin grouping on white.
5. A soft surface behind the content, heading outside it — with the surface
   **white in light** and tinted in dark, per the new Conventions rule.

**The thing I should have done differently: I judged versions 1–4 in dark mode
only.** The light-mode screenshot was the first time the actual constraint was
visible, and that constraint is what decided the answer. Four of the five
iterations were avoidable. The convention above is written so the next person
starts from the relationship rather than from a colour.

Dividers survive _inside_ the two long logs only (`divide-default/60`), where
they help scan 24 rows. Short lists — teams, roles, consents at 1-4 rows — have
none.

`AdminSection` is shared, and **nine admin pages still use `UCard` as a section
container** (`churches/[churchId]`, `consents/[consentId]`, the project detail
pages, the maintenance tools). Not swept: worth confirming the treatment reads
right on one page first, which it now does.

### 2026-09-21 — role-assignment dialog: pick the scope, don't type its ULID

The dialog asked for a scope **type** the role already determines, and then for
the scope's **id as free text** — placeholder "Skriv inn church-ID". Assigning
a church admin meant leaving the page, finding a ULID, and pasting it back.

**The role determines the scope, so the type question is gone.** The pairing is
not a guess: the role service's own tests assign `ChurchAdmin` with a church id,
`ProjectAdmin` with a project id, `TeamLead` with a team id and `Admin` with
none (`internal/services/roles_test.go`). Nothing validated the combination
server-side, so the old form could express "Menighetsadmin scoped to a
Prosjekt" and it would have been stored.

**The id field is now a searchable picker** of the right entity, loaded only
while the dialog is open. Roles needing no scope say so in words rather than
leaving an empty field.

**Teams are reached through their project**, not from one flat list.
`TeamFilter` has no free-text field, so a global team list could not be searched
server-side — and a single project can hold over a thousand teams (1,125 in the
live one). Project first, then its teams.

**A new guard, and the reason it matters:** submit is disabled until a scoped
role has its scope. Previously a scoped role submitted with an empty id sent
`scopeId: undefined` and was assigned **globally** — a far larger grant than the
admin asked for, with no warning. Mutation-checked: removing the guard fails
exactly that case.

Tests: `test/component/AdminUserRoles.test.ts`, 7 cases, including the
scope-name resolution and its id fallback. `UModal` needed stubbing because it
teleports its content — the one case the repo's testing notes say requires a
stub. Component 218 → 225.

### 2026-09-21 — user detail page split into six components (#4, half)

`users/[userId]/index.vue`: **1,182 → 240 lines**, and the template is now a
list of six panels with nothing else in it.

| Component                | Lines | Owns                                                         |
| ------------------------ | ----- | ------------------------------------------------------------ |
| `AdminUserIdentity`      | 330   | header, technical details, sync + church lock (3 mutations)  |
| `AdminUserConsents`      | 248   | the three-shape normalisation, withdrawal modal (1 mutation) |
| `AdminUserRoles`         | 236   | role list, scope naming, add/revoke modal (2 mutations)      |
| `AdminUserScoreJournal`  | 86    | journal list, source-type labels                             |
| `AdminUserFeedbackPanel` | 82    | feedback list                                                |
| `AdminUserTeams`         | 41    | team list                                                    |
| the page                 | 240   | two queries, the breadcrumb, and wiring                      |

**Each panel owns its own mutations and modals, and emits `changed`** for the
page to refetch. That is what made the split worth doing rather than cosmetic:
the page previously held six mutations, four `ref` flags, two modals and nine
handlers for things it did not otherwise care about. `AdminUserRoles` is now the
only file that knows what a role scope is; `AdminUserConsents` the only one that
knows consents come back in two shapes.

**Three naming and placement details worth recording:**

- **`AdminUserFeedbackPanel`, not `AdminUserFeedback`** — that name is already
  taken by the "Gi oss tilbakemelding" widget in the admin shell. Components
  register in one flat namespace (`pathPrefix: false`), so the directory does
  **not** disambiguate them; a collision would have silently shadowed one.
- **The `gql()` mutation definitions travelled with their callers.** Codegen
  globs the whole frontend, so they generate from anywhere — but leaving
  `SyncUser` in a page that no longer calls it is the same dead-reference
  problem as any other orphan.
- **The two "showing N of M" notices are now derived** from the rows actually
  present rather than hardcoded. That is what made the feedback panel's old
  claim of "Viser 10" wrong while it rendered up to 100; the component cannot
  restate a number it does not have.

The truncation notices also fixed the score journal's inherited-by-copy version,
which said "Viser 100" for the same reason.

Gate: typecheck 0, lint 0 errors, 604 unit + 218 component, build exit 0. The
boundary tests (`domain-boundary`, `shared-root`, `layers`) pass unchanged —
worth checking explicitly, since six new components in a layer is exactly the
shape those tests police.

**`my-church/units.vue` (1,109) is now the only remaining #4 outlier.**

### 2026-09-21 — user page: scope names, width cap, consents sorted

**1. The roles card names its scope instead of printing a ULID.**
`RoleScope` (`gql/roles.graphqls:5`) already resolves `church`, `project` and
`team` server-side — the query only asked for `{ id type }`, so the card showed
`Omfang: Church (CH01K9VZ865699692N7FVTXYR4AQ)`. It now reads
`Church — Østfold`. The id survives as the _fallback_, not the default: a scope
pointing at something deleted still has to render, and then the raw id is the
only honest thing left.

This also closed an inconsistency the same page had just created — machine ids
were collapsed into "Tekniske detaljer" as support aids, while two of them stayed
inline in the most human-readable panel on the page.

**2. Capped at `max-w-6xl`**, the same as the home dashboard. Full panel width
stretched every card to ~1640px with its content in the left third, and left
each role row's delete button orphaned ~1500px from the label it deletes — you
had to track across empty space to see what you were about to remove. That
distance was the real cost, not the emptiness.

**3. Consents are one list, grouped by sorting.** Three sub-headings
(Ventende / Akseptert / Avvist) each wrapped rows that _already_ carried a badge
saying the same word. Now sorted by status — pending first, since it is the only
one wanting an admin to act — then alphabetically within each.

The normalisation is the interesting part: the API returns **two different
shapes**. `pendingConsents` are bare `Consent`s, while `acceptedConsents` and
`rejectedConsents` are `UserConsent`s wrapping one, with an `actionDate` and a
`managementType` that decides whether the consent can be withdrawn here. A
`ConsentRow` type flattens all three, and `rowKey` is prefixed per source
because ids can collide across the lists.

**Worth being honest about size:** the template lost ~50 lines and the script
gained ~115, so the file went **1,115 → 1,182**. Flattening three copy-pasted
blocks into one is still the right trade, but the normalisation logic belongs in
an `AdminUserConsents` component rather than in a page that is already the #4
outlier. This is the third change to make that split more attractive rather than
less.

### 2026-09-21 — user header, second pass: three things I got wrong

Shown the before/after, the user's verdict was "I don't know if it is better".
Fair — the four bug fixes were wins, but the layout had two regressions I
introduced and one thing I made worse. Recorded because the pattern is the same
each time: **adding something prominent is not the same as adding something
useful.**

- **Role badges in the header were strictly redundant, and worse than what they
  duplicated.** They rendered raw enums (`CHURCH_ADMIN`, `TEAM_LEAD`) while the
  "Roller og tillatelser" card one screen below shows the same roles with proper
  Norwegian labels _plus_ their scope and a delete action. The page already had
  a `roleLabels` map I did not use. Removed.
- **"Østfold 25 år nb" ran together as one string.** The old labelled rows were
  scannable and I traded that for compactness; a bare `nb` also says nothing to
  a reader who does not already know it. Now separated with `·`, and the code
  goes through the existing `dbLanguageToLocale` (the DB stores `no` where the
  app uses `nb`) into `Intl.DisplayNames`, so it reads "norsk bokmål".
- **Menighetslås was over-promoted.** It had been hiding as the third `dt`
  inside a definition list, so I gave it a full-width card — which put the
  _rarest_ action on the page in its second-most prominent slot. It is now a
  compact labelled row sharing a quiet line with "Tekniske detaljer", with the
  state still spelled out rather than reading "Synk-lås: Ikke låst".

The middle position was the right one both times and I overshot it in each
direction before finding it.

### 2026-09-21 — user detail page: identity first, four bugs fixed

Analysis of `users/[userId]/index.vue` found **four things that were simply
wrong**, separate from any layout opinion. All four are fixed; the layout
changes are the first three items of the agreed plan.

1. **`image` was fetched and never rendered.** The query asked for the avatar
   URL and no `UAvatar` existed on the page. Same class as the `logoImage` bug
   on the project cards: paid for, dropped. On a page about a person it is the
   fastest identity cue available.
2. **`email` was not fetched at all** — while the users _list_ shows it under
   every name. The detail page therefore displayed **less** identifying
   information than the row that links to it.
3. **Points were project-scoped and presented as global.** Both
   `points(projectId:)` and `adminScoreJournal(filter:)` are filtered to the
   current project, but the panel said only "Poenglogg" and "N poeng" — which
   any reader takes as a lifetime total. The header now names the project.
4. **The feedback panel's own notice was untrue.** It rendered
   `Viser 10 av N oppføringer` while `feedbackEntries` was unsliced over a
   `first: 100` query — so it showed up to 100 rows and claimed 10. The query is
   `first: 10` now, which makes the notice true _and_ removes a 100-row wall
   from a panel that already links to the full feedback page. A label that lies
   is worse than a missing one.

**Layout changes:**

- **The header is an identity card**: avatar, name, email, linked church,
  age/language, role badges. It previously carried the least information on the
  page — a single string — while church, roles, teams and points all sat below
  the fold.
- **Machine identifiers are collapsed** behind "Tekniske detaljer" (user ULID,
  Members-ID, Members-UUID, church ID, created-at). They had been the _first_
  block on the page, above age, language and church: support aids in the prime
  position, the same irrelevant-first pattern as the old dashboard counters.
  Still one click away.
- **Church sync-lock got its own row.** It was the third `dt` inside a
  definition list about the church — a real, consequential admin action hiding
  in reference data. It now states what the lock _means_ ("synk fra Members
  endrer ikke menighet") rather than just "Synk-lås: Ikke låst", and is gated on
  `canAssignRoles`.

**Still open on this page**, from the same analysis: the two logs are
hand-rolled lists of bordered divs rather than the standardised table
machinery, actions remain spread across four places, two heading scales
disagree (`h3 text-xs uppercase` for info groups vs `h2 text-xl` for cards),
and the file is 1,108 lines with no component extraction — cross-cutting #4.
Splitting it is the prerequisite for the rest.

### 2026-09-21 — fixed: every user detail page was a 500

`/admin/users/<id>` and `/admin/users/<id>/achievements` both returned
**500 — can't access lexical declaration 'data' before initialization**.

**Cause.** `useAdminPage` set the page label with a plain `watchEffect`, which
runs its effect **synchronously on creation**. Both pages call

```ts
useAdminPage(() => data.value?.user.name);
```

_above_ the query that declares `data` — not carelessly: the main query needs a
project id computed from a first query, so the natural reading order puts the
breadcrumb line before it. Evaluating the getter mid-setup reads `data` in its
temporal dead zone and throws, and Nuxt renders that as a 500 page rather than
a missing breadcrumb.

**Not caused by this session's work.** The hazard arrived with `useAdminPage` in
the 2026-09-18 breadcrumbs change and had been latent since; it needed someone
to open a user detail page. Checked by diffing the script region across commits
— the declaration order is untouched by the `AdminQueryState` conversion, which
only rewrote templates.

**Fixed at two levels, deliberately.**

- `watchEffect(..., { flush: 'post' })` in the composable, so the getter is
  never called during a caller's setup. This is the fix that matters: **23 pages
  call `useAdminPage`**, 8 of them with a getter over query data, and the next
  one written in the natural order would have reintroduced it. Call order is now
  irrelevant. The crumb appears one tick later, which is invisible — the data it
  names has not loaded yet either.
- Both call sites moved below their query anyway, so the correct order is what a
  reader sees.

**Two things this exposed about my own detection.** A first scan for
use-before-declare reported only `achievements.vue`, missing `index.vue`
entirely, because `data` is bound in a multi-line destructure. A second,
broader scan reported _zero_ — it matched
`const { data: currentProjectData } = …` and concluded `data` was declared
above, when that line binds `currentProjectData`. The bug was found by reading
the file, not by either scan. Renaming destructures defeat naive
declaration-order greps; worth remembering before trusting one.

**Tests.** `test/component/useAdminPage.test.ts` reproduces the fault with a
getter that throws until "initialized", standing in for a TDZ binding — it
fails with the user's exact error message when `flush: 'post'` is removed.
Mutation-checked in both directions.

Three cases in the existing `test/unit/useAdminPage.test.ts` needed an
`await nextTick()`, which is the honest consequence of the deferral rather than
a workaround, and is documented at the top of that file. Unit 604 (unchanged),
component 214 → 218.

### 2026-09-21 — AdminQueryState; #2 complete

**`AdminQueryState`** replaces the
`v-if="fetching" / v-else-if="error" / v-else` chain on **all 17 pages** that
had it. With the table half done earlier, cross-cutting #2 is closed.

**It is not primarily a dedupe — it fixes a bug those 17 pages shared.**
`v-if="fetching"` unmounts the entire page body on _every_ refetch: after a
mutation, on a cache refresh, on a route-param change. The page flashes back to
a spinner despite already having content to show. Three `my-church` pages had
noticed and hand-rolled a `hasLoadedOnce` ref watching `data`; the other
fourteen had not. The component shows the spinner only before the query has
ever settled, so every page gets the good behaviour and the three hand-rolled
refs are deleted.

Two details that took care:

- **Settling is keyed on the `true → false` transition of `fetching`, not on
  `fetching` being falsy.** Every admin query is paused until auth is ready, so
  `fetching` is false before the query ever runs — treating that as "settled"
  would suppress the first spinner entirely. There is a test for exactly this.
- **Error outranks a stale fetch once settled**, so a failed refetch cannot
  leave a page spinning forever.

**The conversion was mechanical but not safe to do blind.** Wrapping means
finding each chain's following element and its matching close tag across
nested Vue templates. A mismatched tag is a hard SFC compile error, so
`pnpm build` is the check that actually matters here — it passed, and two pages
were done by hand rather than by the transformer: `achievements/index.vue`,
whose chain continues into a `UEmpty` branch, and `projects/index.vue`, where an
explanatory comment sits between the chain and the element it explains and had
to travel with it.

Tests: 7 cases, mutation-checked — restoring the naive `showLoading = fetching`
fails exactly the refetch case. Component 207 → 214.

Two components, and **all 13 admin tables now carry both states** — up from
10 with `#empty` and **0** with `#loading`.

**`AdminTableEmpty`** goes in `UTable`'s `#empty` slot. Justified by repetition
rather than taste: I had written a near-identical 12-line block six times
(users, feedback, teams, scores, challenges, bulk-jobs) before extracting it.
It separates the two cases a list actually has — nothing exists yet vs. nothing
matched — because telling someone "no challenges yet" when they have filtered
them all away is actively misleading and leaves no way back. An icon decorates
only the genuinely-empty case; a filtered miss is transient and decorating it
overstates it.

**`AdminTableLoading`** goes in `#loading`. **`UTable` has had a `#loading`
slot all along and no page used it** — so `:loading` drew a thin progress bar
over an _empty_ table, which looks exactly like an empty list until data lands.
Skeleton rows say "something is coming" and hold the height. The slot renders
inside one cell spanning every column, so it stacks bars rather than faking a
column grid — a fake layout that disagrees with the real one is worse than
honest bars.

**The sibling-empty-state bug is now extinct: six instances, all fixed.**
feedback and bulk-jobs were caught during their conversions; `consents` and the
three `maintenance/fix-*`/`check-*` tools were found by a scan for an empty
block following `</UTable>`. Rendered as a sibling it appears _below_ the
table's own empty row, so an empty list showed two empty states. Six
independent occurrences is a pattern, not carelessness — which is exactly why
it is a component now.

**Two pages moved off a correct-but-worse pattern.** `events/index.vue` and
`superteams/index.vue` used a `v-if` chain (loading → error → empty → table).
That is not the sibling bug and it renders correctly, but `v-if="fetching"`
unmounts the whole table on _any_ refetch, so the page blanks rather than
showing stale rows under a loading indicator. They now use `:loading` plus the
slots, keeping their icons via the new `icon` prop.
`challenges/[challengeId]/sessions.vue` had the same shape with a hand-rolled
empty div and a `'all'` filter sentinel, now wired to the component's `clear`.

Tests: 8 component cases, including the two that encode the intent — a filtered
miss must say something different from an empty list, and the way out must exist
only in the filtered case. Component 199 → 207.

**Still open in #2:** detail and form pages. 30 pages use `AdminErrorState` and
21 use `AdminLoadingState`, largely as the same `v-if error / v-else-if fetching`
chain — the candidate for an `AdminQueryState` wrapper, not yet built.

### 2026-09-21 — challenges paginated; the other three project lists left alone

Of the four project-scoped lists stuck on `first: 50`, **only challenges was
worth converting.** The decision came from each entity's filter input and its
real cardinality, not from consistency for its own sake.

| List         | Verdict       | Why                                                                                                                              |
| ------------ | ------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| challenges   | **converted** | Core entity, grows over a long study; `ChallengeFilter.challengeType` is a real facet                                            |
| superteams   | left          | `SuperTeamFilter` offers only min/max teams/members — no categorical facet — and a project has a handful (5 in the live project) |
| events       | left          | Same, plus events are barely used (scope decisions); a date-range facet on an empty list is chrome                               |
| achievements | **blocked**   | Drag-to-reorder does not compose with pagination; needs a product decision, see below                                            |

A toolbar with no filter and a pagination footer over five rows is worse than
what is already there, so superteams and events keep their plain grids. Both
still carry `first: 50`; the risk is low but not zero, and the honest fix if it
ever bites is a count in the header rather than paging five items.

**Challenges** now paginates properly and filters by type. `eventId` was
available as a second facet and deliberately not offered — events are barely
used, so it would be a permanently empty control.

**One duplication collapsed while converting.** The table labels a challenge
from its GraphQL `__typename` (`QuizChallenge` → "Quiz") and the new filter
labels the same four kinds from the `ChallengeType` enum (`QUIZ` → "Quiz").
Two label maps for one concept is a drift waiting to happen, so the typename now
maps to the enum and there is a single set of labels. The old "unknown typename
falls back to Enkel" behaviour is kept, so a challenge kind added server-side
degrades instead of rendering blank.

**Achievements: left alone**, on the user's call the same day. The conflict is
real — `VueDraggable` reorder over a paginated list has no meaning, because
dragging an item to the top of page 3 cannot express "make this first overall",
and `reorderAchievements` takes the full ordered id list.

It remains the **highest truncation risk left in the admin panel**: reading
achievements multiply with articles, so a large project is likelier to pass 50
here than anywhere else, and the 51st achievement silently does not render.
Recorded so that the day it bites, the cause is already written down rather
than rediscovered.

Whenever it is picked up, the shape to prefer is a **dedicated reorder mode**:
the list pages normally, and a "Endre rekkefølge" toggle loads the full set and
enables dragging. Page-local ordering is cheaper and semantically wrong.

### 2026-09-21 — consents stays off AdminListView

`consents: [Consent!]!` is a **plain list**, not a connection — no args, no
`pageInfo`, no `totalCount` (`gql/consents.graphqls:45`). `AdminListView`
requires a pagination, so adopting it there meant making `pagination` optional
in both the component and `useListState`, with a count-only footer for the
unpaginated case.

That was built, then **reverted** on the user's call once the reason was clear.
Worth recording why the revert was the right move rather than keeping the
flexibility "for later": consents is the only genuinely unpaginated admin list,
and the four that are unpaginated today (`challenges`, `achievements`,
`events`, `superteams` at `first: 50`) are slated to _gain_ pagination, so they
will take the paginated path. Optional-pagination support would have been an
unused branch in a shared component — the same thing as `AdminTopPerformers`,
the placeholder tile and the dead `actions` column, all of which this effort
deleted.

If consents ever wants the toolbar, the cheaper route is to give
`consents` a proper connection server-side, which fixes the truncation question
at the same time. Making the shared component tolerate two shapes is the more
expensive option, not the less.

### 2026-09-21 — teams, scores and bulk-jobs converted; RelayPagination retired

All four previously-paginated lists are now on `AdminListView`, so
`RelayPagination.vue` is **deleted**. Each page also gained a filter it did not
have, chosen from what its own filter input already supported:

| Page      | Filter added                              | Source                                   |
| --------- | ----------------------------------------- | ---------------------------------------- |
| teams     | superteam, with an "Uten superlag" option | `TeamFilter.superTeamId` + `noSuperTeam` |
| scores    | source type                               | `ScoreJournalFilter.sourceType`          |
| bulk-jobs | (kept its two) now URL-synced             | `BulkJobFilter`                          |

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

**`useListState` stays string-only and the page bridges.** URL params _are_
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
`<UEmpty v-if="!fetching && feedbacks?.length === 0">` as a _sibling_ of
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
