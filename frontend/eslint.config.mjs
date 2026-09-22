// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'
import prettier from 'eslint-config-prettier/flat'
import { globalIgnores } from 'eslint/config'

export default withNuxt([
  globalIgnores(['**/api/']),
  {
    /**
     * The admin panel lives in `layers/admin/`, so the boundary is a directory
     * rather than the hand-maintained file list this used to be — which had
     * drifted, missing several admin-only modules.
     *
     * Two directories must not reach into it: root `app/`, the shared layer
     * (the API client, the auth and permission guards, the shared UI kit), and
     * `layers/user/`, the user-facing app.
     *
     * Admin -> shared and admin -> user stay allowed on purpose: the admin
     * panel renders user-facing components to preview the end-user experience
     * (AdminProjectThemePreview, AdminChallengeCardPreview).
     */
    name: 'interact/domain-boundary',
    files: ['app/**/*.{vue,ts}', 'layers/user/**/*.{vue,ts}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              // The leading `#` MUST stay escaped. These patterns are matched
              // with gitignore semantics, where an unescaped leading `#` marks
              // a comment — `'#layers/admin/**'` is silently dropped and the
              // alias form goes uncaught, which is exactly what happened here.
              group: ['\\#layers/admin/**', '**/layers/admin/**'],
              message:
                'User-facing and shared code must not import admin code. Move the shared part into a neutral location instead. See notes/frontend-admin-restructure.md.',
            },
          ],
        },
      ],
    },
  },
  /**
   * Last, so it wins: Prettier owns formatting, ESLint owns correctness.
   * `vue/html-self-closing` and Prettier disagreed about `<img />` vs `<img >`
   * and each undid the other on every run, so `pnpm lint` rewrote files it had
   * no business touching. This turns off every rule that overlaps.
   */
  prettier,
])
