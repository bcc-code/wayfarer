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
export function usePushNotifications() {
  const config = useRuntimeConfig()

  const isSupported = computed(() => {
    if (import.meta.server) return false
    return 'serviceWorker' in navigator && 'PushManager' in window
  })

  const permission = ref<NotificationPermission>(
    import.meta.client ? Notification.permission : 'default',
  )

  const subscription = ref<PushSubscription | null>(null)
  const isSubscribed = computed(() => subscription.value !== null)
  const isLoading = ref(false)
  const error = ref<Error | null>(null)

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
          return null
        }
      }

      const registration = await navigator.serviceWorker.ready

      // Check for existing subscription
      let sub = await registration.pushManager.getSubscription()

      if (!sub) {
        // Create new subscription
        // VAPID public key should come from your backend/config
        const vapidPublicKey = config.public.vapidPublicKey as
          | string
          | undefined

        if (!vapidPublicKey) {
          error.value = new Error('VAPID public key not configured')
          return null
        }

        sub = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(vapidPublicKey),
        })
      }

      subscription.value = sub

      // Send subscription to backend
      await sendSubscriptionToBackend(sub)

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
      // Remove from backend first
      await removeSubscriptionFromBackend(subscription.value)

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

  /**
   * Send the push subscription to your backend
   */
  async function sendSubscriptionToBackend(
    sub: PushSubscription,
  ): Promise<void> {
    const { getAccessToken } = useAuth()
    const token = await getAccessToken()

    // TODO: Replace with your actual GraphQL mutation or API call
    const response = await fetch(`${config.public.apiUrl}/push/subscribe`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        endpoint: sub.endpoint,
        keys: {
          p256dh: arrayBufferToBase64(sub.getKey('p256dh')),
          auth: arrayBufferToBase64(sub.getKey('auth')),
        },
      }),
    })

    if (!response.ok) {
      throw new Error('Failed to save subscription on server')
    }
  }

  /**
   * Remove the push subscription from your backend
   */
  async function removeSubscriptionFromBackend(
    sub: PushSubscription,
  ): Promise<void> {
    const { getAccessToken } = useAuth()
    const token = await getAccessToken()

    // TODO: Replace with your actual GraphQL mutation or API call
    const response = await fetch(`${config.public.apiUrl}/push/unsubscribe`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        endpoint: sub.endpoint,
      }),
    })

    if (!response.ok) {
      throw new Error('Failed to remove subscription from server')
    }
  }

  // Initialize: check for existing subscription
  onMounted(async () => {
    if (isSupported.value) {
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
