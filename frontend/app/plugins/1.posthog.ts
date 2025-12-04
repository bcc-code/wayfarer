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
    defaults: '2025-11-30',
    person_profiles: 'identified_only',
    capture_exceptions: true,
    capture_heatmaps: true,
    capture_performance: true,
    loaded: (posthog) => {
      if (import.meta.env.DEV) posthog.debug()
    },
  })

  return {
    provide: {
      posthog,
    },
  }
})
