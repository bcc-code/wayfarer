export function useAnalytics() {
  const { $posthog } = useNuxtApp()

  function track(event: AnalyticsEvent, properties?: Record<string, unknown>) {
    $posthog().capture(event, properties)
  }

  function identify(userId: string, properties?: Record<string, unknown>) {
    $posthog().identify(userId, properties)
  }

  function reset() {
    $posthog().reset()
  }

  return {
    track,
    identify,
    reset,
  }
}
