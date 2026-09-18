// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'
import { globalIgnores } from 'eslint/config'

/**
 * Files that make up the admin domain. Everything else is user-facing or
 * shared, and must not reach into admin code.
 *
 * Admin -> user is deliberately allowed: the admin panel legitimately renders
 * user-facing components to preview the end-user experience (see
 * AdminProjectThemePreview and AdminChallengeCardPreview).
 */
const ADMIN_DOMAIN = [
  'app/pages/admin/**',
  'app/components/admin/**',
  'app/layouts/admin.vue',
  'app/layouts/church-admin.vue',
  'app/composables/useAdminNav.ts',
  'app/utils/adminNav.ts',
]

export default withNuxt([
  globalIgnores(['**/api/']),
  {
    name: 'wayfarer/domain-boundary',
    files: ['app/**/*.{vue,ts}'],
    ignores: ADMIN_DOMAIN,
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: [
                '**/components/admin/*',
                '**/composables/useAdminNav',
                '**/utils/adminNav',
                // Applies once the admin layer exists; harmless until then.
                '#layers/admin/**',
              ],
              message:
                'User-facing and shared code must not import admin code. Move the shared part into a neutral location instead. See notes/frontend-admin-restructure.md.',
            },
          ],
        },
      ],
    },
  },
])
