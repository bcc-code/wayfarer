import { defineConfig } from 'vitest/config'
import { defineVitestProject } from '@nuxt/test-utils/config'
import { resolve } from 'path'

export default defineConfig({
  test: {
    projects: [
      // Pure-logic unit tests (composables, utils). Fast, no DOM.
      {
        resolve: {
          alias: {
            '~': resolve(__dirname, './app'),
          },
        },
        test: {
          name: 'unit',
          include: ['test/{e2e,unit}/*.{test,spec}.ts'],
          environment: 'node',
        },
      },
      // Component tests — rendered in a Nuxt runtime environment so that
      // auto-imports (computed, composables, global components, $t) resolve.
      await defineVitestProject({
        test: {
          name: 'component',
          include: ['test/component/**/*.{test,spec}.ts'],
          environment: 'nuxt',
          setupFiles: ['test/component/setup.ts'],
        },
      }),
    ],
  },
})
