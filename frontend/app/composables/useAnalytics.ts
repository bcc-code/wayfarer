import type { ApiObject, IdentifyTraits } from '@rudderstack/analytics-js'

export function useAnalytics() {
  const { $rudderstack, $posthog } = useNuxtApp()

  function track(event: AnalyticsEvent, properties?: ApiObject) {
    $rudderstack.track(event, properties)
    $posthog().capture(event, properties)
  }

  function identify(userId: string, traits?: IdentifyTraits) {
    $rudderstack.identify(userId, traits)
    $posthog().identify(userId, traits)
  }

  function page() {
    $rudderstack.page()
  }

  function reset() {
    $rudderstack.reset()
    $posthog().reset()
  }

  return {
    track,
    identify,
    page,
    reset,
  }
}
