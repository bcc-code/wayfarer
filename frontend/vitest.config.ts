import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    name: 'unit',
    include: ['test/{e2e,unit}/*.{test,spec}.ts'],
    environment: 'node',
  },
})
