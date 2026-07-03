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
├── app/
│   ├── pages/              # File-based routing
│   ├── components/         # Vue components (Design* = design system)
│   │   ├── global/         # Auto-imported globally (icons, buttons)
│   │   ├── achievements/   # Feature-scoped components
│   │   ├── standings/      # Feature-scoped components
│   │   └── ...
│   ├── layouts/            # default, admin, church-admin
│   ├── composables/        # Vue composables (useAuth, usePushNotifications, etc.)
│   ├── utils/              # Pure utility functions
│   ├── plugins/            # Numbered for load order (0.urql, 1.auth0, 2.rudderstack, ...)
│   ├── middleware/          # Route guards (auth.global, admin, superadmin)
│   ├── graphql/
│   │   ├── fragments/      # Reusable GraphQL fragments (*.gql)
│   │   ├── queries/        # GraphQL queries (*.gql)
│   │   └── mutations/      # GraphQL mutations (*.gql)
│   ├── api/
│   │   └── generated.ts    # ⚠ GENERATED — do not edit
│   └── assets/             # Images, fonts, styles
├── test/
│   ├── unit/               # Vitest unit tests
│   └── utils/              # Test utilities and mocks
├── service-worker/         # PWA service worker
├── nuxt.config.ts          # Nuxt configuration
├── codegen.ts              # GraphQL codegen config
├── package.json
└── .prettierrc             # Code style: no semi, single quotes, trailing commas
```

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
- Stub heavy child components with `stubs: { Foo: true }` (auto-stub preserves
  props + component name); use a custom template stub only when you need real
  DOM (e.g. an `<a>` to click)

```typescript
// @vitest-environment nuxt
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
```

See `test/component/EmptyState.test.ts` (simplest) and
`ChallengeCard.test.ts` (computed logic + mocked composable + stubs) as
worked examples.
