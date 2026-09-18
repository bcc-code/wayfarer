// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'
import { globalIgnores } from 'eslint/config'

export default withNuxt([
  globalIgnores(['**/api/']),
  {
    /**
     * The admin panel lives in `layers/admin/`, so the boundary is a directory
     * rather than the hand-maintained file list this used to be — which had
     * drifted, missing several admin-only modules.
     *
     * Root `app/` is the shared layer: the API client, the auth and permission
     * guards, the shared UI kit. It must not reach into the admin layer.
     *
     * Admin -> shared stays allowed on purpose: the admin panel renders
     * user-facing components to preview the end-user experience
     * (AdminProjectThemePreview, AdminChallengeCardPreview).
     */
    name: 'interact/domain-boundary',
    files: ['app/**/*.{vue,ts}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['#layers/admin/**', '**/layers/admin/**'],
              message:
                'User-facing and shared code must not import admin code. Move the shared part into a neutral location instead. See notes/frontend-admin-restructure.md.',
            },
          ],
        },
      ],
    },
  },
])
