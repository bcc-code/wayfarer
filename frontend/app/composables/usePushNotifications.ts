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
  const { executeMutation: registerSubscription } =
    useRegisterPushSubscriptionMutation()
  const { executeMutation: unregisterSubscription } =
    useUnregisterPushSubscriptionMutation()

  const isSupported = computed(() => {
    if (typeof window === 'undefined') return false
    // Service workers don't work in dev mode
    if (import.meta.env.DEV) return false
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
      const registration = await navigator.serviceWorker.ready
      const sub = await registration.pushManager.getSubscription()
      subscription.value = sub
      return sub
    } catch (err) {
      console.error('[Push] Failed to get subscription:', err)
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
    return result
  }

  /**
   * Subscribe to push notifications
   * Returns the subscription object to send to your backend
   */
  async function subscribe(): Promise<PushSubscription | null> {
    console.log('[Push] Subscribe called, isSupported:', isSupported.value)

    if (!isSupported.value) {
      error.value = new Error('Push notifications are not supported')
      return null
    }

    isLoading.value = true
    error.value = null

    try {
      // Request permission if not already granted
      console.log('[Push] Current permission:', permission.value)
      if (permission.value !== 'granted') {
        const result = await requestPermission()
        console.log('[Push] Permission result:', result)
        if (result !== 'granted') {
          error.value = new Error('Notification permission denied')
          isLoading.value = false
          return null
        }
      }

      console.log('[Push] Getting service worker registration...')
      const registration = await navigator.serviceWorker.ready
      console.log('[Push] Service worker ready')

      // Check for existing subscription
      let sub = await registration.pushManager.getSubscription()
      console.log('[Push] Existing subscription:', sub)

      if (!sub) {
        // Get VAPID public key from config
        const vapidPublicKey = config.public.vapidPublicKey as
          | string
          | undefined
        console.log(
          '[Push] VAPID public key:',
          vapidPublicKey ? 'present' : 'missing',
        )

        if (!vapidPublicKey) {
          error.value = new Error('VAPID public key not configured')
          isLoading.value = false
          return null
        }

        console.log('[Push] Creating new push subscription...')
        sub = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(vapidPublicKey),
        })
        console.log('[Push] New subscription created:', sub)
      }

      subscription.value = sub

      // Send subscription to backend via GraphQL
      const input = {
        endpoint: sub.endpoint,
        p256dh: arrayBufferToBase64(sub.getKey('p256dh')),
        auth: arrayBufferToBase64(sub.getKey('auth')),
      }
      console.log('[Push] Registering subscription:', input)

      const result = await registerSubscription({ input })
      console.log('[Push] Registration result:', result)

      if (result.error) {
        console.error('[Push] Registration error:', result.error)
        throw new Error(result.error.message)
      }

      return sub
    } catch (err) {
      error.value =
        err instanceof Error ? err : new Error('Failed to subscribe')
      console.error('[Push] Subscribe error:', err)
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

    try {
      // Remove from backend via GraphQL
      const result = await unregisterSubscription({
        endpoint: subscription.value.endpoint,
      })

      if (result.error) {
        throw new Error(result.error.message)
      }

      // Then unsubscribe locally
      await subscription.value.unsubscribe()
      subscription.value = null

      return true
    } catch (err) {
      error.value =
        err instanceof Error ? err : new Error('Failed to unsubscribe')
      console.error('[Push] Unsubscribe error:', err)
      return false
    } finally {
      isLoading.value = false
    }
  }

  // Initialize: check for existing subscription (only once)
  onMounted(async () => {
    if (isInitialized.value) return
    if (isSupported.value && typeof Notification !== 'undefined') {
      isInitialized.value = true
      permission.value = Notification.permission
      await getSubscription()
    }
  })

  return {
    isSupported,
    permission,
    subscription: readonly(subscription),
    isSubscribed,
    isLoading: readonly(isLoading),
    error: readonly(error),
    requestPermission,
    subscribe,
    unsubscribe,
    getSubscription,
  }
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
