import {
  useRegisterPushSubscriptionMutation,
  useUnregisterPushSubscriptionMutation,
} from '~/api/generated'

/**
 * Composable for managing push notification subscriptions
 *
 * Usage:
 * ```ts
 * const { subscribe, unsubscribe, isSubscribed, permission, isSupported } = usePushNotifications()
 *
 * // Request permission and subscribe
 * await subscribe()
 *
 * // Unsubscribe
 * await unsubscribe()
 * ```
 */
// Shared state across all composable instances
const permission = ref<NotificationPermission>('default')
const subscription = ref<PushSubscription | null>(null)
const isLoading = ref(false)
const error = ref<Error | null>(null)
const isInitialized = ref(false)

export function usePushNotifications() {
  const config = useRuntimeConfig()
  const { track } = useAnalytics()
  const { executeMutation: registerSubscription } =
    useRegisterPushSubscriptionMutation()
  const { executeMutation: unregisterSubscription } =
    useUnregisterPushSubscriptionMutation()

  const isSupported = computed(() => {
    if (typeof window === 'undefined') return false
    return (
      'serviceWorker' in navigator &&
      'PushManager' in window &&
      typeof Notification !== 'undefined'
    )
  })

  const isSubscribed = computed(() => subscription.value !== null)

  /**
   * Get the current push subscription if it exists
   */
  async function getSubscription(): Promise<PushSubscription | null> {
    if (!isSupported.value) return null

    try {
      const registration = await getServiceWorkerRegistration()
      const sub = await registration.pushManager.getSubscription()
      subscription.value = sub
      return sub
    } catch {
      // Don't log timeout errors during init - service worker may not be ready yet
      return null
    }
  }

  /**
   * Request notification permission from the user
   */
  async function requestPermission(): Promise<NotificationPermission> {
    if (!isSupported.value) return 'denied'

    const result = await Notification.requestPermission()
    permission.value = result
    track(AnalyticsEvent.PushPermissionRequested, {
      permission_granted: result === 'granted',
      permission_result: result,
    })
    return result
  }

  /**
   * Subscribe to push notifications
   * Returns the subscription object to send to your backend
   */
  async function subscribe(): Promise<PushSubscription | null> {
    if (!isSupported.value) {
      error.value = new Error('Push notifications are not supported')
      return null
    }

    isLoading.value = true
    error.value = null

    try {
      // Request permission if not already granted
      if (permission.value !== 'granted') {
        const result = await requestPermission()
        if (result !== 'granted') {
          error.value = new Error('Notification permission denied')
          isLoading.value = false
          return null
        }
      }

      // Get service worker with timeout
      const registration = await getServiceWorkerRegistration()

      // Check for existing subscription
      let sub = await registration.pushManager.getSubscription()

      if (!sub) {
        // Get VAPID public key from config
        const vapidPublicKey = config.public.vapidPublicKey as
          | string
          | undefined

        if (!vapidPublicKey) {
          error.value = new Error('Push notifications not configured')
          isLoading.value = false
          return null
        }

        // Subscribe with timeout
        sub = await withTimeout(
          registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: urlBase64ToUint8Array(vapidPublicKey),
          }),
          PUSH_SUBSCRIBE_TIMEOUT,
          'Push subscription timed out. Check your network connection.',
        )
      }

      subscription.value = sub

      // Send subscription to backend via GraphQL
      const input = {
        endpoint: sub.endpoint,
        p256dh: arrayBufferToBase64(sub.getKey('p256dh')),
        auth: arrayBufferToBase64(sub.getKey('auth')),
      }

      const result = await registerSubscription({ input })

      if (result.error) {
        throw new Error(result.error.message)
      }

      track(AnalyticsEvent.PushSubscriptionEnabled)

      return sub
    } catch (err) {
      error.value =
        err instanceof Error ? err : new Error('Failed to subscribe')
      return null
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Unsubscribe from push notifications
   */
  async function unsubscribe(): Promise<boolean> {
    if (!subscription.value) return true

    isLoading.value = true
    error.value = null

    const endpoint = subscription.value.endpoint

    try {
      // Unsubscribe locally first - this is the critical operation
      await subscription.value.unsubscribe()
      subscription.value = null

      // Then remove from backend (can be retried if it fails)
      const result = await unregisterSubscription({ endpoint })

      if (result.error) {
        // Local unsubscribe succeeded, backend failed - not critical
        // Backend will eventually clean up stale subscriptions
        console.warn('[Push] Backend unsubscribe failed:', result.error.message)
      }

      track(AnalyticsEvent.PushSubscriptionDisabled)

      return true
    } catch (err) {
      error.value =
        err instanceof Error ? err : new Error('Failed to unsubscribe')
      return false
    } finally {
      isLoading.value = false
    }
  }

  // Initialize on mount
  onMounted(async () => {
    if (!isSupported.value || typeof Notification === 'undefined') return

    // Always sync permission state (user may have changed in browser settings)
    permission.value = Notification.permission

    // Only fetch subscription once per app lifecycle (expensive operation)
    if (!isInitialized.value) {
      isInitialized.value = true
      await getSubscription()
    }
  })

  return {
    isSupported,
    permission,
    subscription: readonly(subscription),
    isSubscribed,
    isInitialized: readonly(isInitialized),
    isLoading: readonly(isLoading),
    error: readonly(error),
    requestPermission,
    subscribe,
    unsubscribe,
    getSubscription,
  }
}

const SW_READY_TIMEOUT = 10000 // 10 seconds
const PUSH_SUBSCRIBE_TIMEOUT = 15000 // 15 seconds

/**
 * Promise wrapper with timeout
 */
function withTimeout<T>(
  promise: Promise<T>,
  ms: number,
  errorMessage: string,
): Promise<T> {
  return Promise.race([
    promise,
    new Promise<T>((_, reject) =>
      setTimeout(() => reject(new Error(errorMessage)), ms),
    ),
  ])
}

/**
 * Get service worker registration with timeout
 */
async function getServiceWorkerRegistration(): Promise<ServiceWorkerRegistration> {
  return withTimeout(
    navigator.serviceWorker.ready,
    SW_READY_TIMEOUT,
    'Service worker failed to activate. Try refreshing the page.',
  )
}

/**
 * Convert a VAPID public key from base64 URL-safe string to Uint8Array
 */
function urlBase64ToUint8Array(base64String: string): BufferSource {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')

  const rawData = atob(base64)
  const outputArray = new Uint8Array(rawData.length)

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i)
  }

  return outputArray.buffer
}

/**
 * Convert an ArrayBuffer to base64 string
 */
function arrayBufferToBase64(buffer: ArrayBuffer | null): string {
  if (!buffer) return ''
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]!)
  }
  return btoa(binary)
}
