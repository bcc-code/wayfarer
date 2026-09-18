# Frontend admin restructure

Improving the UX, usability and structure of the admin pages. The admin section
is the bulk of the frontend — **41 of 52 pages, ~12,250 lines** — and grew
page-by-page without a shared shell or a consistent route structure.

Branch: `feature/admin-restructure`.

---

## The three original ideas

1. Move to Nuxt layers for user-facing / admin "domains"?
   https://nuxt.com/docs/4.x/directory-structure/layers
2. Restructure admin interface to be project-scoped
3. More standard SaaS sidebar-layout for admin, with a project-selector in the
   sidebar

## What was decided

| Idea                   | Decision                                                                                                                                  |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| **1 — Nuxt layers**    | **Yes, after the IA work.** Re-examined 2026-09-18 — see the log. For organisation, not enforcement; the boundary is enforced separately. |
| **2 — project-scoped** | **Yes** — the highest-value item. Includes moving `teams`/`scores` under a project, with redirect stubs so bookmarks survive.             |
| **3 — SaaS sidebar**   | **Yes**, and mostly assembly rather than invention.                                                                                       |
| church-admin layout    | **Stays separate** — gets navigation, stops duplicating shared blocks.                                                                    |

**Why layers come after the IA work.** See the 2026-09-18 re-examination in the
log for the verified details. In short: layers are a legitimate fit and cheaper
than first assessed, but they organise files without isolating them, so the
boundary needs separate enforcement. Sequencing them after the route work avoids
moving the same files twice. There is no bundle win either way — `ssr: false`
already splits per route, and the PWA config already sets
`globIgnores: ['**/admin/**']`.

## Why project-scoping is the real win

The **permission model is already project-scoped but the IA is not.**
`usePermissions.ts` has `RoleType.ProjectAdmin`, `hasProjectAdminFor(projectId)`,
`canEditProject`, `canManageScoresFor`, `canCreateTeamFor` — yet navigation is
global. The mismatch is visible in the code as a TODO at
`frontend/app/composables/usePermissions.ts:217`:

```
canManageTeam = (_teamId?: string) => {
  // TODO: Add project-based permission check when team's project is available
```

It cannot do the check because `/admin/teams/[teamId]` has no project in the
route. Project-scoping fixes this structurally rather than by plumbing.

Secondary: `projects/[projectId]/index.vue` (492 lines) is a `UTabs` kitchen
sink — achievements, challenges and superteams behind one mega-query with a
hardcoded `first: 50` per section and no pagination. Tabs are not URLs, so
there are no deep links, no browser back, no per-section loading and no code
splitting.

## Findings worth not re-investigating

Verified at source. These shape the design, and two of them contradict
reasonable-sounding assumptions.

**Adding `[projectId].vue` does NOT rename any route.** Nuxt 4.4 generates
routes through `unrouting` (`nuxt/dist/index.mjs:34`), whose `prepareRoutes`
contains:

```js
if (children.some((c) => c.path === "")) name = void 0; // unrouting/dist/index.mjs:271
```

When a child sits at path `""`, the **parent** is left unnamed and the child
keeps the name. So `[projectId]/index.vue` retains `admin-projects-projectId`
and all ~120 named-route bindings keep resolving. _Corollary:_ inside
`[projectId].vue` use bare `useRoute()` — the parent has no name.

**The real danger is deleting `[projectId]/index.vue`.** Without a child at
path `""`, the name lands on the parent, which renders `<NuxtPage />` with
nothing to put in it: the route resolves, the shell paints, and the content
area is **blank — no 404, no error, no type error**. That silently breaks the
project-list cards, ~20 breadcrumbs, and the post-save `navigateTo` targets in
`edit.vue`, `events/new.vue`, `achievements/new.vue`, `challenges/new.vue` and
`superteams/new.vue`. Guarded by a test — see the 2026-09-18 log entry.

**Route meta vs. middleware merge differently.**

| Mechanism                       | Semantics                                                                                        | Source                                                                                       |
| ------------------------------- | ------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------- |
| `middleware:`                   | **Unions** across every matched record — a parent guard still runs when a child declares its own | `nuxt/dist/pages/runtime/plugins/router.js:144-152` (Set-based collection over `to.matched`) |
| custom meta keys (`permission`) | **Shallow merge, child wins**                                                                    | `vue-router/dist/vue-router.js:794` (`matched.reduce((m, r) => assign(m, r.meta), {})`)      |

Both are the semantics we want: a parent guard that always runs, and a child
able to say `permission: 'project:edit'` and mean exactly that.

**`superteams/index.vue` is not a superteams list.** It is the 676-line
ladder-to-heaven distribution tool (`@unovis/vue` charts, raw `fetch` to
`/plugins/ladder-to-heaven/preview-superteams` and `/distribute-superteams`).
It merely occupies the route name `admin-projects-projectId-superteams` and has
to move to `superteams/distribute.vue` first. The superteams _list_ exists only
as a tab.

**The `events/` subtree is orphaned** — nothing links to `events/new` or
`events/[eventId]`; they are reachable only by typing the URL.

**The permission rules contradict each other in three places.** `canAccessTeams`
is `isSuperAdmin || isProjectAdmin`, but the teams pages guard with
`middleware: 'superadmin'` — a project admin sees the "Lag" nav entry and is
bounced when they click it. Likewise `canManageScores` includes project admins
while `createScoreAdjustment` is `@requireRole(["m2m","admin","superadmin"])`,
so they can fill the score form and get a 403 on submit. The server's
`@requireRole` directives in `gql/*.graphqls` are the source of truth; the
matrix must be reconciled against them rather than ported forward.

**Backend needs no changes for the route moves.** `TeamFilter` and
`ScoreJournalFilter` already accept `projectId` (`gql/teams.graphqls:66`,
`gql/scoring.graphqls:32`), and `Team.parentProject: Project!`
(`gql/teams.graphqls:14`) lets the `/admin/teams/[teamId]` redirect stub resolve
its target. That stub needs an async lookup, so it is a thin page, not a static
`definePageMeta({ redirect })`.

**Existing pieces to reuse.** `@nuxt/ui` ships the whole `UDashboard*` family
free since v4 (Pro merged in) — only `UDashboardSearch` is used today, so the
sidebar is assembly, not a new dependency. `AdminUserMenu.vue` already takes a
`collapsed` prop and is styled for a sidebar footer. `useGroupedProjects.ts`
(current/future/past) feeds the project selector directly.

---

## Update log

### 2026-09-18 — type-check gate restored (`9b30a4dc`)

`frontend/CLAUDE.md` documented `pnpm typecheck`, but no such script existed in
`package.json` and `vue-tsc` was not installed (only bare `tsc`). Since typed
routes are the safety net for the whole restructure, this had to come first.

Added `vue-tsc@3.3.11` and `"typecheck": "nuxt typecheck"`, then cleared the
**34 errors** it reported across 15 files. Two were real, user-visible bugs
rather than type noise:

- **Bulk-job polling never ran.** `fixMissingContentProgressAsync` and
  `fixMissingStreakProgressAsync` return `[BulkJob!]!`, but both maintenance
  pages did `const job = result.data.<field>; jobId.value = job.id` — reading
  `.id` off an array yields `undefined`, so `pollJobStatus()` bailed on its
  first line and the job never reported completion. `result.data` from urql's
  `executeQuery` is also a `Ref` that was never unwrapped.
- **Achievement streak items lost two fields.** The `StreakAchievement` branch
  of the query omitted `source` and `publishedAt` while the mapper read both,
  so they were silently `undefined`.

The rest were mechanical: 13 `@click="x = false"` handlers (a template
assignment expression returns the assigned value, which Nuxt UI 4's strict
`onClick` type rejects), `Locale[]` vs. disabled locales in
`hiddentreasures.ts`, `nullish()` vs. `optional()` in `edit.vue`, and the
`UDashboardSearch` groups type in `admin.vue`.

### 2026-09-18 — Nuxt UI 4.9 → 4.11.1 (`4dd0f1fd`)

Dependency bump. The `UDashboard*` family is intact (30 components); the gate
was re-verified green afterwards.

### 2026-09-18 — frontend CI + real middleware tests (`f629baed`)

CI is **Semaphore** (`.semaphore/semaphore.yml`), not GitHub Actions — easy to
miss, since `.github/workflows/` holds only the Phrase translations job. There
was no frontend job at all beyond `pnpm run build` inside "Build release", and
`nuxt generate` does not typecheck. So a broken route name, a failing test or a
lint error all shipped silently.

Added a `Frontend tests` block running in parallel with `Backend tests` and
gating `Build release`, split into three parallel jobs: `lint + typecheck`,
`unit tests`, `component tests`. It invokes `pnpm exec eslint .` rather than
`pnpm lint`, because the latter runs with `--fix` and would "pass" by rewriting
files (it also fights Prettier, rewriting `<br />` → `<br >`).

Three new test files, all exercising the **real** modules:

| File                                       | Covers                                   |
| ------------------------------------------ | ---------------------------------------- |
| `test/unit/routes.test.ts`                 | Route manifest snapshot + two invariants |
| `test/unit/middleware-admin.test.ts`       | `admin.ts` / `superadmin.ts` — 17 tests  |
| `test/unit/middleware-auth-global.test.ts` | `auth.global.ts` — 12 tests              |

`routes.test.ts` builds the manifest with the same `unrouting` calls Nuxt makes,
so it needs no Nuxt runtime and runs in the fast `unit` project. It asserts a
committed snapshot, the parent-must-have-a-`""`-child guard, and that every
`name:` / `useRoute('…')` reference resolves. The blank-page guard was verified
to actually fire, against an isolated copy of `app/pages`:

| Case                       | `admin-projects-projectId` resolves to | Guard       |
| -------------------------- | -------------------------------------- | ----------- |
| parent **+** `index.vue`   | the child at path `''`                 | passes      |
| parent, **no** `index.vue` | the parent — content area empty        | **fails** ✓ |

**Deleted 793 lines of fake coverage.** The `Global Auth Middleware` and
`Admin Middleware` suites in `auth.test.ts` re-implemented the middleware inside
each test body and asserted against the copy (`auth.test.ts:1337` was literally
`// Simulate middleware logic`), so they could not fail when the middleware
changed — and the copy encoded the **wrong rule**: it checked
`!isSuperAdmin && !isAdmin → 403` while the real `middleware/admin.ts` also
admits `ProjectAdmin` and `ChurchAdmin`. A total auth regression would have left
`pnpm test` green.

One test in `middleware-admin.test.ts` deliberately asserts _current_ behaviour
for the project-admin/teams contradiction described above. It will need updating
when the permission matrix is reconciled — intentionally, so the change is
deliberate rather than silent.

Lint went **47 errors → 0**, mostly as a consequence of removing the fake tests.
The remainder was genuine dead code, including a stray
`import type { D } from '@vite-pwa/assets-generator/...'` in `DesignSwitch.vue`
that an IDE had auto-inserted, and a cascade in `QuizPredefinedQuestion.vue`
where removing `nextButtonText` orphaned `continueText` and then `useI18n`
itself. The three `vue/valid-v-slot` failures are false positives from
`UTable`'s dotted `accessorKey` slot names (`#user.name-cell`, which
eslint-plugin-vue reads as a directive modifier) and got scoped disables with an
explanation rather than a blanket rule change. Both middleware also moved off
`useState<any>` to `useState<GetMeQuery['me'] | null>`.

### 2026-09-18 — removed the dead Hidden Treasures link (`efdabe65`)

`app/pages/index.vue` computed `hiddenTreasuresLink` but nothing rendered it —
the supporting `getHiddenTreasureLocale` util, its unit test, and a translated
`goToHiddenTreasures` ("Gå til bibelstudie") key in every locale all still
exist, so a bible-study link appears to have fallen out of the front-page
template at some point. Confirmed as intentional and the dead computed removed;
the util and translations stay, so restoring the link is cheap.

### 2026-09-18 — nav model + dashboard shell (PR 1)

No route changes: the route-manifest snapshot is untouched, and the permission
watcher in `layouts/admin.vue` is deliberately left in place so the shell
rewrite and the permission consolidation stay separately revertible.

**`app/utils/adminNav.ts` — navigation as data.** `GLOBAL_NAV` and `PROJECT_NAV`
arrays of `{ label, icon, to, match?, can? }`, plus two pure functions
(`visibleNavItems`, `isNavItemActive`). `to` and `match` are typed
`keyof RouteNamedMap` from `vue-router/auto-routes`, so a mistyped route name
fails `pnpm typecheck` rather than at runtime.

Active state now keys off route **names**, not `route.fullPath.includes(...)`.
The old substring check would light up the global "Lag" entry on
`/admin/projects/x/teams` — i.e. it breaks exactly when teams move under a
project. A test covers that case specifically.

`PROJECT_NAV` currently holds only Oversikt, Superlag and Innstillinger — the
routes that exist today. Challenges, achievements and events join it when their
list routes are created; that is an append, not a redesign.

**`app/composables/useAdminNav.ts`** feeds both the sidebar and the command
palette from that one source, so they cannot drift apart. `route.params` is a
union across every route under `typedPages`, so `projectId` is read with
`'projectId' in route.params` rather than direct access.

**`layouts/admin.vue`** rewritten as `UDashboardGroup` > `UDashboardSidebar`
(project switcher, nav, `AdminUserMenu` footer, collapse toggle) >
`UDashboardPanel` (`UDashboardNavbar` header, `<slot />` body). Mobile nav comes
free — `UDashboardSidebar` is desktop-only and `UDashboardNavbar` renders the
slideover toggle by default.

Two shell details worth knowing:

- `UDashboardGroup` is `fixed inset-0`, so **page scroll moves into the panel
  body**. Anything assuming document-level scroll needs checking.
- Pages still bring their own `UContainer` + padding, so the panel body padding
  is zeroed with `:ui="{ body: 'p-0 sm:p-0 gap-0 sm:gap-0' }"`. The responsive
  variants must be listed explicitly: the default is `p-4 sm:p-6 gap-4 sm:gap-6`
  and tailwind-merge is variant-aware, so a bare `p-0` overrides only the base
  and leaves `sm:p-6` in place. This shim goes away as pages move onto a shared
  page header.

**`AdminConfirmDialog` is finally mounted** (in the shell), and all five
`window.confirm` call sites migrated to `useConfirm()` — achievements,
challenges, sessions, superteams and events detail pages. The single native
sentence became a title plus description (`Slette "X"?` /
`Denne handlingen kan ikke angres.`). The stale comment in `useConfirm.ts`
claiming the dialog was already rendered is corrected.

**`AdminUserMenu`'s logout now works.** Its "Logg ut" item had `label` and
`icon` but no `onSelect`, so clicking it did nothing. The component already had
a `collapsed` prop and sidebar-shaped styling, so it dropped into the sidebar
footer unchanged otherwise.

**New `AdminProjectSwitcher`** in the sidebar, grouped Aktive / Kommende /
Tidligere via the existing `useGroupedProjects`, filtered by `canViewProject`,
on a deliberately lighter query than `AdminProjectsPage` with
`requestPolicy: 'cache-first'` (the sidebar mounts on every admin route, so the
default `cache-and-network` would refetch the project list on each navigation).
Switching projects keeps you in the same _section_ where one matches, else the
overview.

**Tests.** `test/unit/adminNav.test.ts` (12 tests) covers permission gating and
active-state rules without mounting anything. `routes.test.ts` gained a check
that every nav entry targets a route that exists — the existing dangling-name
scan only sees `name: '...'` literals, and `useAdminNav` builds its route
objects dynamically, so the nav model would otherwise have slipped past it.

Verified beyond the gate: `pnpm build` succeeds, and both `UDashboardGroup` and
`UDashboardSidebar` theme strings appear in the output bundle, confirming the
components resolved. Unit tests 445 → 457.

### 2026-09-18 — layers re-examined, and a domain boundary that actually holds

Prompted by a second opinion from Nuxt AI arguing that layers are a documented
fit for exactly this admin/user split. Re-checked against the installed Nuxt
(4.4.8, `@nuxt/kit` 4.4.6) rather than from memory. It was right on the
mechanics, and the earlier cost estimate in this note was wrong.

**Confirmed correct:**

- `~/layers/*` really is auto-registered — `@nuxt/kit` globs `layers/*` and
  pushes them into `_extends` (`kit/dist/index.mjs:793`). No `extends` config.
- The `#layers/<name>` alias exists (`kit/dist/index.mjs:823`). It is in kit,
  not `nuxt/dist`, which is why a first grep of `nuxt/dist` came up empty.
- `layers/*/app/**` is the right Nuxt 4 shape; the kit watches exactly that.

**The cost was overstated here earlier.** Of the five items originally listed,
three need nothing at all: `codegen.ts` already globs `./**/*.vue` from the
frontend root so layer files are picked up as-is; ESLint's flat config globs
everything; typed-pages and tsconfig are generated across layers, and `i18n/`
stays at the root. Only two real edits remain — `vitest.config.ts`'s
`'~': resolve(__dirname, './app')` alias (`~` is per-layer in Nuxt), and the
hardcoded `app/pages` root in `test/unit/routes.test.ts`.

**The one claim that does not hold: layers do not enforce the boundary.**
Every layer's component dirs are flattened into one registry
(`nuxt/dist/index.mjs:3558-3565`; layer priority only breaks _name_ collisions),
and every layer's `composables/` and `utils/` into one `composablesDirs` array
(`:3843-3860`). A page in `layers/user` can auto-import an admin component with
zero ceremony and no error. Layers give file locality and per-layer config, not
import boundaries. This matters more than usual here because `pathPrefix: false`
already puts all 105 components in one flat namespace — which is _why_
`ColorModeSelector`/`AdminColorModeSelector` and friends exist as duplicate
pairs. Layers will not fix that.

**Trap to avoid in the move.** A layer's pages dir merges at the _root_ of the
route table, so the `admin/` directory must be preserved inside the layer:
`layers/admin/app/pages/admin/projects/index.vue` → `/admin/projects`, whereas
`layers/admin/app/pages/projects/index.vue` → `/projects`. The route-manifest
snapshot in `test/unit/routes.test.ts` is the check: a pure layer move must
leave it **byte-identical**.

**Decision:** do the layer split after the IA work (permissions, then
tabs → routes), so the same files are not moved twice, and enforce the boundary
separately — which landed now, in two halves:

- **`eslint.config.mjs`** gained a `wayfarer/domain-boundary` block:
  `no-restricted-imports` banning `**/components/admin/*`,
  `**/composables/useAdminNav`, `**/utils/adminNav` and `#layers/admin/**`
  (the last is inert until the layer exists) from everything outside the admin
  domain. Admin → user stays allowed on purpose: the admin panel renders
  user-facing components to preview the end-user experience
  (`AdminProjectThemePreview`, `AdminChallengeCardPreview`).
- **`test/unit/domain-boundary.test.ts`** covers what ESLint structurally
  cannot. `<AdminUserMenu />` in a template and `useAdminNav()` in a script have
  no import statement, so `no-restricted-imports` never sees them. The test
  scans user-facing source for `<Admin*` and `useAdmin*(` instead.

Both were verified to actually fire, not just to pass: a probe file importing
`~/utils/adminNav` is rejected by ESLint, and a probe component using
`<AdminUserMenu />` is reported by the test while **ESLint reports nothing** —
which is the concrete demonstration that the lint rule alone would have given
false confidence. The boundary is clean today; the only pre-existing hit was
`useAdminNav.ts` importing its own model, i.e. admin code sitting in a shared
folder rather than a real leak.

Unit tests 457 → 459.

### Gate status after the above

| Check           | Before | After                          |
| --------------- | ------ | ------------------------------ |
| Lint errors     | 47     | **0**                          |
| Type errors     | 34     | **0**                          |
| Unit tests      | 436    | **457**                        |
| Component tests | 149    | 149                            |
| Frontend in CI  | none   | lint + typecheck + both suites |

---

## Remaining sequence

Each step is independently shippable.

1. **Permission consolidation.** `definePageMeta({ permission })` + one global
   middleware reading `to.meta.permission`; delete the `routePermissions`
   watcher in `layouts/admin.vue:81-117` and the triplication. Reconcile the
   contradictions above against `@requireRole` first.
2. **Project context + tabs → routes.** `[projectId].vue` parent with a
   `useCurrentProject()` composable, split the mega-query, add
   `challenges/index.vue`, `achievements/index.vue` (keep the
   `vue-draggable-plus` reorder), `events/index.vue`, move LADD to
   `superteams/distribute.vue` and write the real superteams list.
3. **Route moves.** `teams`/`scores` under `[projectId]` with redirect stubs;
   resolve the `canManageTeam` TODO.
4. **Nuxt layers.** `git mv` into `layers/user/app/**` and `layers/admin/app/**`,
   keeping the `admin/` directory inside the admin layer's `pages/`. Verified by
   the route-manifest snapshot staying byte-identical. Then widen the `~` alias
   in `vitest.config.ts`, add the layer pages dirs as roots in
   `routes.test.ts`, and tighten the ESLint boundary group to `#layers/admin/**`.

Optional follow-ups, deliberately out of scope: an `AdminPage` scaffold to
absorb the ~20 copy-pasted inline breadcrumb headers; admin i18n (the nav model
should hold keys from day one so this is a labelling change later); splitting
the 1,000-line outliers (`my-church/units.vue` 1,105, `users/[userId]/index.vue`
1,068); and giving `churches/[churchId].vue` a home — it has no list page and no
nav entry, reachable only from `users/[userId]/index.vue:588`.

## Watch out for

- **Static query variables + a parent that no longer remounts.** Every project
  page passes `variables: { projectId: route.params.projectId }` as a plain
  object (`[projectId]/index.vue:96`, `edit.vue:41`, `superteams/index.vue:33`,
  `challenges/new.vue`, …). That is correct only because the whole page
  remounts on navigation. Once a persistent `[projectId].vue` parent exists,
  switching projects changes the param **without a remount** and a static object
  keeps querying the old project. These must become `computed()`. Most likely
  regression in the whole effort.
- **urql refetch storm.** The client default is `cache-and-network`
  (`plugins/0.urql.ts:12`). With the project query subscribed from the parent,
  the sidebar, the breadcrumb and the switcher, first paint fires 3-4 identical
  requests — use `cache-first` plus an explicit `refresh()` after mutations.
  Tab switching likewise goes from zero-network to a request per section.
- **`setLocale('nb')` is a persistent global side effect.** `layouts/admin.vue:11`
  calls it unconditionally on mount and `@nuxtjs/i18n` persists the choice in a
  cookie with nothing restoring it, so a German user who opens `/admin` once has
  the **entire consumer app** switched to Norwegian permanently. It must stay out
  of any shared shell composable or nav component — `my-church/**` is the one
  fully translated admin area and `church-admin.vue` deliberately does not force
  the locale.
- **`admin.css` leaks globally.** It sets `--ui-primary` / `--ui-radius` on
  `:root` but is imported from inside `admin.vue` / `church-admin.vue`. In an SPA
  the chunk stays loaded, so the admin theme persists after navigating back to
  the user app. Pre-existing; worth scoping while rewriting the layout.
- **`ErrorState` / `EmptyState` use user-app design tokens** and will look wrong
  inside the dashboard chrome.
