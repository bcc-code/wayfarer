import posthog from 'posthog-js'
import type { PostHog } from 'posthog-js'

export default defineNuxtPlugin<{ posthog: PostHog | undefined }>(() => {
  const config = useRuntimeConfig()

  if (!config.public.posthogKey) {
    return {
      provide: {
        posthog: undefined,
      },
    }
  }

  posthog.init(config.public.posthogKey, {
    api_host: config.public.posthogHost,
    capture_pageview: true,
    capture_pageleave: true,
    capture_exceptions: true,
    capture_heatmaps: true,
    capture_performance: true,
  })

  return {
    provide: {
      posthog,
    },
  }
})
