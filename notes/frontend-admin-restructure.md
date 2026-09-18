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

| Idea                   | Decision                                                                                                                                       |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| **1 — Nuxt layers**    | **Done — three layers**: root (shared), `layers/user`, `layers/admin`. For organisation, not enforcement; the boundary is enforced separately. |
| **2 — project-scoped** | **Yes** — the highest-value item. Includes moving `teams`/`scores` under a project, with redirect stubs so bookmarks survive.                  |
| **3 — SaaS sidebar**   | **Yes**, and mostly assembly rather than invention.                                                                                            |
| church-admin layout    | **Stays separate** — gets navigation, stops duplicating shared blocks.                                                                         |

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

### 2026-09-18 — admin visual design, adapted from the Nuxt UI calendar template

Reference: github.com/nuxt-ui-templates/calendar. Three things about it are worth
recording, because none are obvious from reading the repo.

**Its `app.config.ts` cannot be copied on its own.** It is the visible half of a
"liquid glass" system whose other half is `app/assets/css/main.css`: the config
builds surfaces from a `glass-material` utility and `--glass-bg`, `--control-bg`,
`--control-bg-hover`, `--well-bg` and `--overlay-blur` custom properties, all
declared in that stylesheet. Copying the config alone leaves every dropdown,
modal, popover and select with a **transparent** background.

Rather than port the reference's CSS verbatim, the system was rebuilt on stock
Tailwind, which is the house rule here:

| Reference                                     | Here                                                                             |
| --------------------------------------------- | -------------------------------------------------------------------------------- |
| `@utility glass-material` + `--glass-filter`  | `backdrop-blur-xl backdrop-saturate-150 backdrop-brightness-105`                 |
| `bg-(--glass-bg)`                             | `bg-glass` (a `@theme` colour)                                                   |
| `bg-(--control-bg)` / `-hover`                | `bg-control` / `bg-control-hover`                                                |
| `bg-(--well-bg)`                              | `bg-well`                                                                        |
| `backdrop-blur-(--overlay-blur)`              | `backdrop-blur-sm`                                                               |
| `@media (prefers-reduced-transparency)` block | `@custom-variant reduceTransparency`, applied once via a shared `solid` constant |

The tokens live in the existing `@theme` block with their dark values in the
existing `.dark` block, matching how the rest of the file already declares
colours. No hand-written CSS rules were added.

**Its `sidebar:` block targets `USidebar`, not `UDashboardSidebar`.** The
reference hand-rolls its shell (`<div class="isolate relative flex h-svh
overflow-hidden">`) around `<USidebar variant="floating">`. Our shell uses
`UDashboardSidebar`, which carries the resizable persisted width and the mobile
slideover that `USidebar` has no companion component for, so the treatment was
split: the glass surface goes on `dashboardSidebar.slots.root` in
`app.config.ts` (shared theme), and the geometry that makes it float — `m-2`,
the radius, `min-h-0` to undo the theme's `min-h-svh` (which would otherwise
overflow the viewport by that margin) and `border-e-0` to drop the theme's rule
in favour of the ring — is passed as a `ui` prop from `layouts/admin.vue`. That
belongs to the one layout rather than to every dashboard sidebar, and an
instance `ui` prop is appended after the theme rather than merged into it, so
it cannot lose a merge.

**`--ui-radius` was deliberately not taken.** The reference uses `0.5rem`; admin
stays at `0.3rem`. Every `rounded-*` utility in the app derives from that token
(`--radius-lg` is `calc(var(--ui-radius) * 2)`), and because `--radius-*` is
declared at `:root` its `var()` resolves there — so overriding `--ui-radius` on
a wrapper element does **not** rescope it. Raising it would have rounded the
user-facing `Design*` components too.

Blast radius is small despite `app.config.ts` being global: user-facing code uses
about a dozen Nuxt UI components, nearly all `UIcon`, because it runs on the
`Design*` system. In practice this dresses the admin panel only.

Also in this pass:

- **Semantic colours stopped leaking in from the user app.** `assets/styles/user.css`
  — imported by the _user-facing_ `default.vue` — sets `--ui-success` and
  `--ui-error` on `:root` from the brand accents (`#9ed63c`, a lime). Those are
  per-layout CSS chunks, so once that layout has been visited the declarations
  stay for the rest of the SPA session and the admin panel renders the user
  app's lime as its success colour. Admin now declares its own semantic tokens
  in `admin.css`, with the palettes (`emerald`/`rose`/`amber`/`blue`) chosen in
  `app.config.ts`.

  Two things learned while verifying, both non-obvious:
  - Nuxt UI injects its generated palette inside `@layer theme`, and **unlayered
    CSS beats any layer**, so plain `:root` declarations in `admin.css` win
    without extra specificity. (The reference template relies on the same thing
    for `--ui-radius`.)
  - `app.config.ts` cannot choose the _shade_ a semantic token resolves to — it
    is fixed at 500 light / 400 dark (`IC(role, 500)` / `IC(role, 400)` in the
    runtime generator). Changing that needs a CSS token, which is why the shade
    lives in `admin.css` while the palette lives in app.config. Admin uses 600
    in light mode, where 500 reads as neon on the pale tints that `subtle`
    badges use; dark stays at 400, matching the default.

- **Admin page background put on one palette.** The shell inherited
  `bg-background-default` from `nuxt.config`'s `rootAttrs` — the _user app's_
  grey (`#efefef` / `#222222`) — while `UCard` and the tables sit on Nuxt UI's
  `--ui-bg` (white / zinc-900). In dark mode that made the page **lighter** than
  the surfaces on it, so panels read as sunk into it rather than raised off it;
  in light mode it put three surfaces within ~7% lightness of each other, drawn
  from two different palettes. The shell now carries
  `bg-neutral-100 dark:bg-neutral-950`: one step behind the surfaces, in both
  modes, in the palette Nuxt UI maps `neutral` to.

  Worth knowing for future debugging: `--ui-color-neutral-*` is **not** in the
  built CSS. Nuxt UI generates the palette at runtime from `app.config`
  (`--ui-color-${color}-${shade}: var(--color-${name}-${shade}, …)`), so a
  `bg-neutral-*` class greps as "referenced but never defined" in the bundle and
  looks broken when it is not.

- **Collapsible sidebar dropped** at the user's request, along with
  `UDashboardSidebarCollapse`.
- **Search is discoverable.** `UDashboardSearch` was mounted but nothing opened
  it, so the command palette was reachable only by shortcut. The sidebar now
  leads with `UDashboardSearchButton`, which renders the meta+K hint itself.
- **The navbar has a title**, derived from the nav model via a new
  `currentTitle` in `useAdminNav` (deepest matching entry wins, so a project
  section beats the global one).
- **Content is full width.** `UContainer` (max-width 80rem, centred) was leaving
  most of a wide screen empty inside the dashboard panel. All 40 admin pages
  swapped it for a plain `div`, dropping the `py-*`/`my-*` the panel body now
  owns while preserving deliberate `max-w-*`; the `p-0` panel shim from PR 1 is
  gone.
- **`QuickAccess` moved** from `fixed bottom-4 left-4` to bottom-right, where it
  no longer sits on top of the user menu in the sidebar footer.

Verified: typecheck 0, lint 0, 459 unit + 149 component, `pnpm build` exit 0, and
the emitted CSS actually contains
`.glass-material{-webkit-backdrop-filter:var(--glass-filter);backdrop-filter:var(--glass-filter)}`
plus the light/dark tokens and the `prefers-reduced-transparency` fallback —
the check that matters, since a config referencing an unregistered `@utility`
would fail silently. **Not verified: appearance.** The admin layout requires an
Auth0 session, so this needs a human look.

### 2026-09-18 — permission consolidation (`32ecac79`)

Three copies of the admin access rules collapsed into one: `middleware/admin.ts`
and `middleware/superadmin.ts` are deleted, the `routePermissions` watcher is
gone from `layouts/admin.vue` (70 lines), and pages now declare
`definePageMeta({ permission })` enforced by `02.admin-permission.global.ts`.

**The matrix was reconciled against `@requireRole` first**, and the decisive
finding is how little `project_admin` actually grants:

| Role                | Operations the server accepts it for |
| ------------------- | ------------------------------------ |
| `superadmin`        | 130                                  |
| `admin`             | 120                                  |
| `church_admin`      | 20                                   |
| **`project_admin`** | **1 — `updateProject`**              |

So a project admin may edit a project they own and nothing else. Everywhere
`usePermissions` granted them more — `canAccessTeams`, `canManageScores`,
`canManageScoresFor`, `canCreateTeamFor` — the server was already returning 403.
Those gates now match, which fixes the nav-offers-then-bounces bug at its source
rather than in the guard, because the sidebar and the middleware read the same
flags. It also closes the long-standing `canManageTeam` TODO: no team mutation
accepts `project_admin`, so the answer was role-only and never needed the team's
project.

Behaviour change worth knowing: someone holding **both** church-admin and
project-admin was previously confined to `/admin/my-church` and locked out of
the projects they run. `isChurchAdminOnly` now treats `ProjectAdmin` as another
admin role.

Four things that only surfaced by running it:

- **Global middleware run in alphabetical order.** `admin-permission.global.ts`
  sorted _before_ `auth.global.ts`, so the guard ran before auth had
  initialised. The old `admin.ts` was a _named_ middleware, and those always run
  after globals — making it global silently changed the order. Both are now
  numerically prefixed (`01.auth`, `02.admin-permission`), matching the
  convention `plugins/` already uses.
- **`isLoading` does not mean the user is loaded.** `useAuth` pauses the `me`
  query until a token exists and runs `watch(fetching, …, { immediate: true })`,
  so `isLoading` is false on the first tick with `me` still null. A guard that
  waited on it redirected signed-in superadmins to the front page. It waits on
  `me` itself now; `!token.value` terminates the unauthenticated case, which is
  reliable only because `01.auth.global.ts` has already cleared an expired
  token by then.
- **Nuxt drops the injection context across an `await`.** `usePermissions()`
  after the `until()` threw "use\* function must be called within a reactive
  context" and 500'd the page. Every composable is now resolved before the
  first await.
- **Augmenting `RouteMeta` alone does not type `definePageMeta`.** Nuxt's
  `PageMeta` carries an index signature, so a mistyped permission compiled
  fine. `app/types/route-meta.d.ts` augments both; a typo now fails typecheck
  with a "did you mean" suggestion.

Tests: `adminPermissions.test.ts` pins the full role × permission matrix (60
cases, every allow _and_ deny); `middleware-admin.test.ts` drives the real guard
against the real `usePermissions` (17). The last two bugs above were each
reproduced as a failing test before being fixed — the context-loss one needed an
emulation of Nuxt's context (plain stubs cannot see it), and the redirect one
needed a real timer delay, since a microtask let it pass on timing luck.

Unit tests 459 → 519.

### 2026-09-18 — project context, tabs → routes

`pages/admin/projects/[projectId].vue` is now a parent route and the `UTabs`
kitchen sink is gone. The route-manifest snapshot confirmed the `unrouting`
behaviour on the real tree — the only change from adding the parent was:

```
+ (unnamed)                 /admin/projects/:projectId()   [projectId].vue
- admin-projects-projectId  /admin/projects/:projectId()   [projectId]/index.vue
+ admin-projects-projectId  /admin/projects/:projectId()/  [projectId]/index.vue
```

The parent goes unnamed, `admin-projects-projectId` stays with `index.vue`, and
all ~60 named-route bindings keep resolving. Nothing else moved.

**`useCurrentProject()` owns the project**, not the parent page. The consumers
that need it most — sidebar, project switcher, navbar title — are rendered by
the _layout_, which is an ancestor of the page, so anything the page provided
would be invisible to them. urql's document cache keys on operation + variables,
so every caller shares one result. `requestPolicy: 'cache-first'` because the
client default would otherwise fire an identical request per subscriber on first
paint.

**Every query variable in the project subtree became a `computed`** (12 files).
This was the single most likely regression: pages passed
`variables: { projectId: route.params.projectId }` as a plain object, which is
correct only while each page remounts on navigation. With a persistent parent,
switching project changes the param _without_ a remount and a static object
silently keeps querying the old project.

New routes:

| Route                       | Source                                                                        |
| --------------------------- | ----------------------------------------------------------------------------- |
| `challenges/index.vue`      | the challenges tab                                                            |
| `achievements/index.vue`    | the achievements tab, drag-reorder intact                                     |
| `events/index.vue`          | new — un-orphans `events/new` and `events/[eventId]`, which nothing linked to |
| `superteams/index.vue`      | the superteams tab — a real list at last                                      |
| `superteams/distribute.vue` | the ladder-to-heaven tool that used to occupy `superteams/index.vue`          |

`[projectId]/index.vue` is now an overview: project header plus a counts-only
query (`first: 0`, `totalCount`) linking to each section. The old mega-query is
gone from `generated.ts` entirely, and with it the hardcoded `first: 50` that
silently truncated any project with more than 50 of anything.

`?tab=` links were written to history by the old tab state, so the overview
redirects the three known values to their replacement routes for one release.

`PROJECT_NAV` grew from 3 entries to 6 now the routes exist — an append, as
predicted, not a redesign.

The domain-boundary test earned its keep here: `useCurrentProject.ts` calls the
generated `useAdminProjectShellQuery`, which matches its `useAdmin*` probe, and
it sits in the shared `composables/` folder. It is admin code, so it joined
`ADMIN_DOMAIN` in both the test and the ESLint rule.

Unit tests hold at 519; the work was structural rather than new logic.

### 2026-09-18 — teams and scores moved under the project

The four remaining project-scoped pages moved to
`projects/[projectId]/{teams,scores}/`, and **Lag** and **Poeng** moved out of
`GLOBAL_NAV` into `PROJECT_NAV`. Both list queries gained
`filter: { projectId }` merged with their pagination cursor, and their now
redundant "Prosjekt" column is gone.

`scores/new.vue` lost its project picker — the route says which project — and
with it the `AdminScoresNewPage` query, which existed only to populate that
dropdown.

**Legacy URLs still resolve.** The route-manifest diff was additions only:
`admin-teams`, `admin-teams-teamId`, `admin-scores` and `admin-scores-new` are
all still in the manifest, now served by stubs.

`/admin/teams/:teamId` is the interesting one. It cannot be a
`definePageMeta({ redirect })`, because that is synchronous and the old URL
carries no projectId — the target has to be looked up. `Team.parentProject`
exists server-side, so the stub queries it and redirects, falling back to an
explanatory empty state when the team is gone or the project is not visible to
the user. The other three are static redirects to the project picker, since
there is no current project to infer.

The user detail page needed no stub: its query already selected
`team.parentProject.id`, so it addresses the project-scoped route directly. Its
"Vis alle" score-journal button was removed instead — that panel spans projects,
and there is no single journal to send it to any more.

Two nav tests were updated rather than deleted: the guard one previously
asserted that the _global_ "Lag" entry must not match
`admin-projects-projectId-teams`. That relationship inverted with the move, so
it now asserts the project entry matches its own branch and its `:teamId` child
but not the legacy global route — the same substring-matching bug, guarded from
the other side.

### 2026-09-18 — project nav in a secondary sidebar

The project section moved out of the primary sidebar into a column of its own, so
the global nav stays a fixed landmark while the project one comes and goes with
the route. Resizing was dropped from both at the same time; with nothing
resizable left, `UDashboardGroup`'s `storage`/`storage-key` went too — they only
persist resize state.

The **project switcher heads the secondary column**, not the primary. Keeping it
in both printed the same truncated project name twice, side by side, and moving
it gives the column an identity: this one _is_ the project. The trade is that
the switcher is unreachable outside a project — the "Prosjekter" list is the way
in, which is the right order anyway. Below `lg`, where there is no second
column, the switcher rides in the primary's slideover with the project nav.

Visual hierarchy is one class: the secondary carries `shadow-none` while the
primary keeps its shadow, so the app rail reads as the top layer and the
contextual column sits flatter behind it.

**Grids moved to container queries, and the project card lost its aspect ratio.**
The card's `aspect-video` was the reason the projects page looked so empty: a
16:9 box meant a card in a wide column stretched to ~320px tall to hold three
lines of text and a date. It is `h-full` now, sized by content, with `block h-full`
on the wrapping link so cards in a row still match each other via the grid's
stretch.

**The overview grid moved to container queries.** `lg:grid-cols-4` measured the
_viewport_, but two sidebars take ~600px out of it, so four cards overflowed the
panel at exactly the widths the breakpoint was meant to cover. It is now
`@container` with `@md:grid-cols-2 @4xl:grid-cols-4`, measured against the panel.
Verified in the built CSS as `@container (min-width:28rem)` and `(min-width:56rem)`
rules rather than `@media`. The same swap was applied to `projects/index.vue`
(`@xl:grid-cols-2 @4xl:grid-cols-3`) and `admin/index.vue`. Any admin page that
sizes in-panel layout off `sm:`/`lg:` has this latent bug — the sidebars take
~300-600px out of the window before the panel gets any.

Both sidebars are capped at `max-w-[250px]`. `default-size` is a _percentage_ of the
viewport (the dashboard context sets `unit: '%'`), so 16% passes 250px on
anything wider than ~1560px; a ceiling is the right fix rather than a smaller
percentage, which would leave the sidebar cramped on a laptop.

Two things that are not obvious from the markup:

- **The secondary sidebar's header is load-bearing.** Its
  `min-h-(--ui-header-height)` is what lines the first nav item up with the
  primary sidebar's and with the navbar. Without it the column starts higher
  than everything beside it.
- **Only one sidebar may be mounted on mobile.** `UDashboardSidebar` registers
  `useRuntimeHook('dashboard:sidebar:toggle', () => open.value = !open.value)`,
  and that hook is global — _every_ mounted sidebar listens. Two of them means
  one tap on the hamburger opens two slideovers, stacked. The secondary is
  therefore mounted on desktop only (`useMediaQuery`, safe here because
  `ssr: false`), and below `lg` the project nav rides along in the primary's
  slideover, which renders that sidebar's default slot.

### 2026-09-18 — breadcrumbs moved into the shell

Twenty-five pages each carried their own `<UBreadcrumb :items="[...]">` inside an
identical bordered wrapper, rendered in the page body — below the navbar's own
border, and scrolling away with the content. They are now derived once and
rendered in the layout's navbar, **in place of the title**: the last crumb is the
title, so showing both said the same thing twice and cost a second
`--ui-header-height` row of chrome on every page.

It goes in the navbar's `#left` slot, which replaces the default
leading/title/trailing group but _not_ the mobile toggle — the navbar renders
that just outside the slot. Routes with no ancestors (`/admin`) fall back to a
plain heading, since a one-item breadcrumb is a title with extra markup.

**Almost all of it is derivable.** `useAdminPage()` builds
`Prosjekter → <project> → <section>` from the route plus the nav model plus
`useCurrentProject()`. A page supplies only what the URL cannot know — the name
of the thing it is showing — via `useAdminPage(() => data.value?.challenge.name)`,
which also becomes the navbar title. Nine detail pages set one; the rest need
nothing.

The label lives in a module-level ref because the layout renders the chrome and
is an _ancestor_ of the page, so provide/inject cannot carry it upward. It is
cleared on scope dispose, or the previous page's name lingers on the next route
until that page's own query resolves.

Two details worth keeping:

- The overview never appears as its own section crumb. `PROJECT_NAV`'s
  "Oversikt" resolves to the same route as the project crumb, so listing it
  would say the same thing twice.
- Only prefix-matching nav entries can be ancestors. "Hjem" is an exact-match
  entry — a destination, not a parent — so `/admin` correctly has no breadcrumb
  at all.

Removing the blocks also killed a query: `superteams/new.vue` fetched
`project { id name }` solely to label its breadcrumb. The shell reads the same
data from `useCurrentProject`'s shared cached query, so that page now makes one
request fewer. Worth checking for elsewhere.

**Follow-up after review: the first cut was incomplete in two ways.**

It supported only one trailing crumb, so routes nested below a detail page —
`challenges/:id/quiz`, `challenges/:id/sessions`, `users/:id/achievements` —
stopped at their section and lost both the entity name and their own. The page
label is now `string | {label, to} | Array<…>`, so those pages supply two, the
first linking back to the detail view. `edit.vue` had the opposite problem: it
set the _project_ name as its label, which `PROJECT_NAV`'s "Innstillinger" crumb
already covers, printing the project twice.

Worse, all seven `my-church/**` pages use the `church-admin` layout, which
renders no breadcrumb at all — so removing their inline ones left them with no
navigation whatsoever, on the one admin surface that has no sidebar either.
Their breadcrumbs are restored as they were, i18n intact. Giving that layout
real navigation is still the separate, deferred item it always was.

The maintenance tools were never regressed but now name themselves rather than
all reading "Vedlikehold".

An audit of every admin page against what its old trail said is what surfaced
these; checking only the two routes that prompted the review would have missed
the my-church regression entirely.

`useAdminPage.ts` tripped the domain-boundary rule on its way in, correctly —
admin code in the shared `composables/` folder, like `useAdminNav` and
`useCurrentProject` before it.

Net: **591 lines deleted, 55 added.** Unit tests 519 → 529.

### 2026-09-18 — viewport breakpoints swept; churches decision recorded

Every grid in an admin-layout page now measures the panel rather than the
window. Five files converted: the four maintenance tools and the LADD
distribution page.

The audit corrected two of my own earlier claims:

- **`users/[userId]/index.vue` has no viewport grid breakpoints at all.** I had
  named it as a likely offender; it is not one.
- **`my-church/units.vue` is not affected either**, and must not be converted:
  those pages use the `church-admin` layout, which is a header plus `<slot />`
  with no sidebar, so their content really is viewport-wide. A container query
  there would be wrong.

A grep for `\b(md|lg):grid-cols` also reports false positives, because `@` is
not a word character — it matches inside the `@md:` container variants it is
meant to distinguish from. Anchor on `(^|[" ])` instead.

**Churches:** `churches/[churchId].vue` stays a drill-down leaf with no list
page and no nav entry, reachable from a user's detail page. Churches are rarely
added or edited, so a top-level section would be dead weight. This closes the
open question rather than deferring it again.

### 2026-09-18 — admin split into a Nuxt layer

`layers/admin/` now holds the 49 admin pages, both admin layouts, all 35 admin
components, `QuickAccess`, five admin composables, `adminNav.ts` and `admin.css`.
Root `app/` is the shared base layer. **Admin layer only** — the user app is 12
pages leaning hard on shared code, so a `layers/user` would have meant a
user-vs-shared verdict on every file for no benefit.

**The routing contract held.** 61 routes, every name and path identical before
and after; only the snapshot's file column changed. That was the whole reason
the route-manifest test exists.

**Two assumptions I had wrong, both corrected by measurement rather than
reasoning:**

- **`~` inside a layer falls back to the root.** The plugin rewrites `~` to the
  layer's srcDir and then calls `this.resolve(...)`, which returns null when the
  layer-local path does not exist — so Vite continues to the global alias. A
  moved page importing `~/api/generated` therefore builds fine, and the `#api`
  alias I had added for it was dead weight and was removed.
- **The breakage is the other direction, and only TypeScript sees it.**
  `.nuxt/tsconfig.json` has a single `~/* -> ../app/*` mapping, so an
  _intra-layer_ `~/utils/adminNav` resolves to the root, where it no longer
  exists. Vite would have resolved it layer-first and built happily. Eight files
  needed relative imports; `pnpm typecheck` found all of them.

The net rule, now in `frontend/CLAUDE.md`: `~/...` for shared root code,
relative paths within the layer, `#layers/admin/app/...` for deliberate
cross-layer references. (The docs say `#layers/<name>` points at the layer's
srcDir; it actually points at its rootDir, hence the `app/` segment.)

**Three silent-failure modes, each checked directly rather than assumed:**

| Hazard                                                                      | Check                                                                    | Result                                                                        |
| --------------------------------------------------------------------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| A layer without `components: { pathPrefix: false }` renames every component | sorted `export const` list from `.nuxt/components.d.ts`, before vs after | 547 names, identical                                                          |
| Tailwind not scanning the layer                                             | admin-only classes in the built CSS                                      | `bg-glass`, `@container (min-width:56rem)`, `dark:bg-neutral-950` all present |
| Codegen reordering `generated.ts`                                           | sorted export-symbol set before vs after `pnpm codegen`                  | 969 symbols, identical set; a 105-line pure permutation                       |

`test/unit/layers.test.ts` (new) locks the three that have no other detector: the
layer must have a `nuxt.config.ts` (a layer without one is silently skipped and
its routes vanish), that config must declare the bare relative `components` path,
and **no global middleware or plugins may live in a layer** — they are gathered
per layer with extended layers first, so a filename prefix only orders within one
layer and a layer's `*.global.ts` would run before `01.auth.global.ts`.

That last one is why `02.admin-permission.global.ts` and `utils/adminPermissions.ts`
stay in root despite being admin by content: they are the boundary itself, the
same argument that keeps `01.auth.global.ts` there.

The boundary enforcers got sharper as a side effect. `ADMIN_DOMAIN` was a
hand-maintained file list in two places that had drifted out of date; it is now
just the layer directory. The auto-import scanner still matters — layers do not
close that vector — and gained the admin-only composables whose names do not
start with `useAdmin` (`useConfirm` above all: its dialog is mounted by the admin
layout, so a user-facing caller gets a promise that never resolves).

Unit tests 531 → 536.

### 2026-09-18 — user-facing app split into a second layer

`layers/user/` now holds the 12 user pages, `default.vue`, all 11 component
directories plus 9 top-level components (`Design*`, `ErrorState`,
`LoadingState`, `EmptyState`, `PageLayout`, `TitleBar`, …), 6 composables, 5
utils and `user.css`. Root `app/` is now **shared only**: the generated client,
the auth and permission guards, the plugins, the GraphQL documents, `main.css`,
`app.config.ts`.

This reverses the "admin layer only" decision recorded in the previous entry.
That call rested on the user-vs-shared verdict being expensive; with admin
already extracted the remaining set was small and mostly unambiguous, and the
three-way split is what actually makes root `app/` mean something.

**Root `app/pages/` keeps exactly three pages** — `auth0-callback.vue`,
`login.vue`, `logout-callback.vue`. Not an oversight: `01.auth.global.ts`
hardcodes those paths, so they belong to the auth boundary, not to either domain.

**Severing admin's dependency on the user design system came first.** The admin
layer reached the user-facing `DesignButton` through the shared `ErrorState` /
`LoadingState`. New `AdminErrorState` / `AdminLoadingState` (Nuxt UI tokens,
identical prop and emit surface) took over in 32 admin files, after which
`DesignButton` is reachable from admin only via `AdminProjectThemePreview` and
`AdminChallengeCardPreview` — the sanctioned admin→user preview direction.
Verified by walking the import graph: "via ErrorState: 0 files".

**The routing contract held again.** 61 routes, name and path columns
byte-identical; only the file column moved. Verified by parsing both snapshots
and comparing the columns directly — and the comparison asserts a non-zero row
count first, because an earlier version of that parser matched nothing and
"passed" vacuously.

**A fourth silent-failure mode, specific to this layer.** `components/global/**`
loses global registration unless the layer _also_ declares
`{ path: 'components/global', global: true }` alongside the `pathPrefix: false`
entry. Three icons are passed as **strings** (`icon="IconSettings"`,
`IconClose`, `IconChevronRight`) and resolve by name alone, so they would render
nothing with no error. Checked in the built bundle rather than the source: the
`nuxt:global-components` plugin chunk registers all 19 icons.

**The ESLint boundary had never worked for the alias form.** A probe import of
`#layers/admin/app/utils/adminNav` produced no error — in `layers/user/` _or_ in
`app/`, so this shipped broken with the admin layer. `no-restricted-imports`
matches patterns with gitignore semantics, where **an unescaped leading `#`
marks a comment**: `'#layers/admin/**'` was silently dropped, and only the
relative form `'**/layers/admin/**'` was ever enforced. There is no config error
and no failing lint for this. Fixed by escaping (`'\\#layers/admin/**'`, which
does match) and by extending `files` to `layers/user/**`, since the rule scoped
to `app/**` would no longer see any user-facing code at all.

Both halves are now pinned by tests that lint a synthetic snippet against the
**real** resolved config (`eslint.config.mjs` exports a thenable
FlatConfigComposer, not an array). Mutation-checked: reverting the escape fails
the new case and nothing else.

`test/unit/layers.test.ts` became layer-generic — it iterates the `layers/`
directory instead of naming admin, and asserts up front that it found
`['admin', 'user']` so it cannot pass vacuously. Its plugin check was also a
false positive waiting to happen: matching `/\/plugins\//` anywhere in a path
flagged `pages/plugins/ladder-to-heaven/excalibur.vue`, which is a page _route_.
It now checks for `app/plugins/` and `app/middleware/` as directories.

`test/unit/domain-boundary.test.ts` scans both non-admin roots now; its vacuity
guard asserts each root contributes files, since scanning only the shared root
would silently miss every user-facing page.

Unit tests 536 → 545.

### Gate status after the above

| Check           | Before | After                          |
| --------------- | ------ | ------------------------------ |
| Lint errors     | 47     | **0**                          |
| Type errors     | 34     | **0**                          |
| Unit tests      | 436    | **545**                        |
| Component tests | 149    | 149                            |
| Build           | ok     | ok                             |
| Frontend in CI  | none   | lint + typecheck + both suites |

---

## Remaining sequence

The sequenced work is complete. Optional follow-ups, deliberately out of scope: admin i18n (the nav model
should hold keys from day one so this is a labelling change later); splitting
the 1,000-line outliers (`my-church/units.vue` 1,105, `users/[userId]/index.vue`
1,068); `churches/[churchId].vue` is settled: it stays a drill-down leaf, see the log.

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
- ~~**`ErrorState` / `EmptyState` use user-app design tokens** and will look
  wrong inside the dashboard chrome.~~ Resolved: the admin layer has its own
  `AdminErrorState` / `AdminLoadingState` on Nuxt UI tokens. Keeping them
  separate is also what severs admin's dependency on the user design system, so
  do not "deduplicate" them back into one shared pair.
