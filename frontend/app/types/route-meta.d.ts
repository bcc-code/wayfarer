import type { AdminPermission } from '~/utils/adminPermissions'

declare module 'vue-router' {
  interface RouteMeta {
    /**
     * Page-level admin permission, enforced by the global `admin-permission`
     * middleware. Omitting it under `/admin` means `'admin'` — any admin role.
     */
    permission?: AdminPermission
  }
}

declare module '#app' {
  /**
   * Nuxt's own PageMeta carries an index signature, so augmenting vue-router's
   * RouteMeta alone types `to.meta` in middleware but lets `definePageMeta`
   * accept any key — including a mistyped permission. Declaring it here is what
   * makes a typo fail `pnpm typecheck`.
   */
  interface PageMeta {
    permission?: AdminPermission
  }
}

export {}
