import { RudderAnalytics } from '@rudderstack/analytics-js'

export default defineNuxtPlugin(() => {
  const rudderstack = new RudderAnalytics()

  // Skip loading analytics in dev mode
  if (!import.meta.dev) {
    const config = useRuntimeConfig()
    rudderstack.load(
      config.public.rudderstackWriteKey,
      config.public.rudderstackDataPlaneUrl,
      {},
    )
  }

  return {
    provide: {
      rudderstack,
    },
  }
})
