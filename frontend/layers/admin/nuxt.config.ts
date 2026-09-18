/**
 * Admin panel layer.
 *
 * Registered automatically: Nuxt globs `layers/*` (@nuxt/kit loadNuxtConfig),
 * and `srcDir` resolves to this directory's `app/` on its own.
 *
 * This file is not optional. A layer without a `components` config falls back
 * to Nuxt's default, which path-prefixes component names — `components/admin/
 * AdminUserMenu.vue` would register as `AdminAdminUserMenu` and every reference
 * in the layer would break, at runtime, with no build or type error.
 *
 * The path is relative rather than `~/components`: `~` is rewritten per layer,
 * and a bare relative path is resolved against this layer's own srcDir, which
 * is what we want and is immune to any alias change.
 */
export default defineNuxtConfig({
  components: {
    dirs: [{ path: 'components', pathPrefix: false }],
  },
})
