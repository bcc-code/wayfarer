import { RudderAnalytics } from '@rudderstack/analytics-js'

export default defineNuxtPlugin(() => {
  const rudderstack = new RudderAnalytics()
  const config = useRuntimeConfig()

  rudderstack.load(
    config.public.rudderstackWriteKey,
    config.public.rudderstackDataPlaneUrl,
    {},
  )

  return {
    provide: {
      rudderstack,
    },
  }
})
