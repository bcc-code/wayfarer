import { RudderAnalytics } from '@rudderstack/analytics-js'

export default defineNuxtPlugin(() => {
  const rudderstack = new RudderAnalytics()
  const config = useRuntimeConfig()

  console.log('Rudderstack config:', {
    writeKey: config.public.rudderstackWriteKey,
    dataPlaneUrl: config.public.rudderstackDataPlaneUrl,
    apiUrl: config.public.apiUrl,
    fullPublic: config.public,
  })

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
