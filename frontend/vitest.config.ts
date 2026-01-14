import { defineConfig } from 'vitest/config'
import { resolve } from 'path'

export default defineConfig({
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
})
