# Frontend — Nuxt 4 SPA

## Tech Stack

- **Framework**: Nuxt 4.3 (Vue 3, `ssr: false` — SPA only)
- **Language**: TypeScript
- **Package Manager**: pnpm
- **GraphQL Client**: urql with auth exchange
- **Styling**: Tailwind CSS 4.1 via @nuxt/ui
- **i18n**: @nuxtjs/i18n (default locale: `nb`, 12+ languages, `no_prefix` strategy)
- **Auth**: Auth0 → exchanged for Wayfarer JWT
- **Analytics**: PostHog + RudderStack
- **Error Tracking**: Sentry
- **Testing**: Vitest — two projects: `unit` (pure logic, node env) and `component` (rendered via `@nuxt/test-utils` in a Nuxt runtime env). No e2e suite yet.
- **PWA**: @vite-pwa/nuxt with service worker

## Directory Layout

```
frontend/
├── app/                    # SHARED + user-facing (the base layer)
│   ├── pages/              # User-facing routes only; admin lives in the layer
│   ├── components/         # Design* design system, global/ icons, shared states
│   ├── layouts/            # default.vue
│   ├── composables/        # useAuth, useAuthReady, usePermissions, ...
│   ├── utils/              # Pure utilities, adminPermissions (see below)
│   ├── plugins/            # Numbered for load order (0.urql, 1.auth0, ...)
│   ├── middleware/         # 01.auth.global, 02.admin-permission.global
│   ├── graphql/            # *.gql — one contract, one generated output
│   ├── api/generated.ts    # ⚠ GENERATED — do not edit
│   └── assets/             # Images, fonts, main.css (the Tailwind entry)
├── layers/
│   └── admin/              # The admin panel, auto-registered by Nuxt
│       ├── nuxt.config.ts  # ⚠ Required — see below
│       └── app/
│           ├── pages/admin/**   # ⚠ The admin/ dir must stay inside
│           ├── components/admin/, components/devtools/
│           ├── layouts/         # admin.vue, church-admin.vue
│           ├── composables/     # useAdminNav, useAdminPage, useCurrentProject, ...
│           ├── utils/adminNav.ts
│           └── assets/styles/admin.css
├── test/
│   ├── unit/               # Vitest unit tests
│   └── utils/              # Test utilities and mocks
├── service-worker/         # PWA service worker
├── nuxt.config.ts          # Nuxt configuration
├── codegen.ts              # GraphQL codegen config
├── package.json
└── .prettierrc             # Code style: no semi, single quotes, trailing commas
```

### Working with the admin layer

A file belongs in `layers/admin/` when its name or content names admin concepts.
It stays in root `app/` when both domains use it, or when it is domain-neutral
infrastructure — the generated client, the auth and permission guards, the shared
UI kit, pure utilities. Root `app/` must not import from the layer; the reverse
is allowed (the admin panel renders user-facing components to preview the
end-user experience). `eslint.config.mjs` and `test/unit/domain-boundary.test.ts`
enforce this — layers organise code, they do not isolate it.

Four rules with no compile-time or runtime error to warn you:

- **`layers/admin/nuxt.config.ts` is required.** A layer directory without one is
  silently skipped and its routes simply vanish.
- **It must declare `components: { dirs: [{ path: 'components', pathPrefix: false }] }`.**
  A layer that declares none defaults to path-prefixed names, so `AdminUserMenu`
  would register as `AdminAdminUserMenu`. The bare relative path matters:
  `~/components` would resolve through the _global_ alias back to root.
- **No global middleware or plugins in the layer.** They are gathered per layer
  with extended layers first, so a filename prefix only orders within one layer —
  a layer's `*.global.ts` would run before `01.auth.global.ts`.
- **Imports within the layer must be relative.** `~` maps to the root `app/` in
  `tsconfig`, so an intra-layer `~/utils/adminNav` fails typecheck. Use
  `~/...` for shared root code, relative paths within the layer, and
  `#layers/admin/app/...` only for deliberate cross-layer references.

`test/unit/layers.test.ts` asserts the first three.

## Key Commands

```bash
pnpm dev            # Start dev server
pnpm codegen        # Generate GraphQL types from gql/ schemas + local operations
pnpm test           # Run all tests (watch mode)
pnpm test:unit      # Run unit tests only (pure logic, node env)
pnpm test:component # Run component tests only (rendered, Nuxt env)
pnpm lint           # ESLint
pnpm format         # Prettier
pnpm build          # Production build
pnpm typecheck      # Vue + TypeScript type checking
```

## Conventions

### Code Style

Prettier config: no semicolons, single quotes, trailing commas, 2-space indent, arrow parens always.

### Pages

Pages use `<script setup lang="ts">` with composition API:

```vue
<script setup lang="ts">
const { isAuthReady } = useAuthReady()
const { data, fetching } = useMyPageQuery({
  pause: computed(() => !isAuthReady.value),
})
</script>
```

- Always pause GraphQL queries until auth is ready
- Use `<LoadingState>` and `<ErrorState>` components for loading/error states

### Components

- **Design system** components use `Design*` prefix (DesignButton, DesignInput, DesignCard, etc.)
- Use CVA (Class Variance Authority) for component variants
- Use `withDefaults(defineProps<...>(), {...})` for prop defaults
- Global components go in `components/global/`, feature components in feature folders

### GraphQL

- Put fragments in `app/graphql/fragments/*.gql`
- Put queries in `app/graphql/queries/*.gql`
- Put mutations in `app/graphql/mutations/*.gql`
- After adding/changing operations, run `pnpm codegen` to regenerate `app/api/generated.ts`
- Generated file provides typed composables (e.g., `useProfilePageQuery()`)
- URQL uses `cache-and-network` request policy

### Composables

- Use `gql()` template tag (from `graphql-tag`) to declare inline fragments/queries
- Leverage VueUse composables (`useLocalStorage`, `useWindowScroll`, etc.)
- Module-level shared state is OK for singletons (e.g., `const permission = ref(...)` outside the function)

### Plugins

Numbered prefix controls load order: `0.urql.ts` → `1.auth0.ts` → `2.rudderstack.ts` → etc.

### Testing

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('Feature', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-06-15T12:00:00'))
  })
  it('should work', () => {
    /* ... */
  })
})
```

- Mock utilities in `test/utils/auth-mocks.ts`
- Use noon times in date tests to avoid timezone boundary issues

#### Component tests

Component tests live in `test/component/` and render real components. They run
in a Nuxt runtime environment (via `@nuxt/test-utils`) so auto-imports
(`computed`, composables, global `Design*`/`NuxtLink` components, `$t`) resolve.

- Start each file with `// @vitest-environment nuxt`
- Mount with `mountSuspended` from `@nuxt/test-utils/runtime`
- Mock Nuxt auto-imports (composables) with `mockNuxtImport`
- `test/component/setup.ts` globally stubs Auth0 and Sentry so app init doesn't
  throw during mount — no per-test boilerplate needed

**Mock composables, render real components.** The boundary of a component test
is data, not UI:

- **Always mock** composables/queries — `useAuth`, the generated
  `use*PageQuery`/`use*Mutation` composables, `useRoute`/`useRouter`, `useNow`.
  There is no backend, auth, or router history in a component test.
- **Prefer real child components** — `DesignButton`, `DesignInput`,
  `DesignPanel`, icons, `NuxtLink`, etc. all render fine and give higher
  fidelity (a broken slot or prop binding actually fails the test). Select them
  with `findComponent(RealComponent)` and read `.props(...)`.
- **Only stub a child when it actively resists the test:**
  - _Teleporting UI_ — `DesignDrawer`/modals wrap `@nuxt/ui` components that
    teleport content out of the wrapper and only render it while open. Stub with
    a template that renders `<slot />` + `<slot name="content" />` and declares
    `emits: ['update:open']` so you can query the form and drive open/close.
  - _Heavy list/data children you assert props on_ — e.g. `LeaderboardList`
    (entrance animations, needs full entry data for its item children). Stub
    with `{ Foo: true }` (auto-stub preserves props + name) and assert on the
    props passed to it.

```typescript
// @vitest-environment nuxt
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
```

Worked examples:

- `EmptyState.test.ts` — simplest render test
- `ChallengeCard.test.ts` — real children, mocked `useAnalytics`
- `StandingsGlobal.test.ts` — driving query state (loading/error/empty/data)
- `StandingsUnitEdit.test.ts` — interactive form flow with a stubbed
  `DesignDrawer` (the only place a stub is required)
