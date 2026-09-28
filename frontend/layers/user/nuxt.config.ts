/**
 * User-facing app layer.
 *
 * Registered automatically: Nuxt globs `layers/*`, and `srcDir` resolves to this
 * directory's `app/` on its own.
 *
 * Both dirs entries are mandatory and mirror the root config:
 *
 * - Without `pathPrefix: false`, a layer falls through to Nuxt's default, which
 *   path-prefixes component names by directory — `design/DesignButton.vue`
 *   would register as `DesignDesignButton`.
 * - Without the `global: true` entry, `components/global/**` loses global
 *   registration. That breaks silently and invisibly: three icons are passed as
 *   *strings* (`icon="IconSettings"`, `IconClose`, `IconChevronRight`) and are
 *   resolved only by name, so they would render nothing with no error.
 *
 * Paths are bare relative strings on purpose — they resolve against this
 * layer's srcDir, whereas `~/components` would resolve through the global alias
 * back to the root app.
 */
export default defineNuxtConfig({
  components: {
    dirs: [
      { path: 'components', pathPrefix: false },
      { path: 'components/global', global: true },
    ],
  },
})
