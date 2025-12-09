import type { NotificationData, PushPayload } from './types'

declare let self: ServiceWorkerGlobalScope

/**
 * Handle incoming push notifications
 */
export async function onPush(event: PushEvent): Promise<void> {
  if (!event.data) {
    console.warn('[SW] Push event received without data')
    return
  }

  const promise = handlePushEvent(event)
  event.waitUntil(promise)
}

async function handlePushEvent(event: PushEvent): Promise<void> {
  let payload: PushPayload

  try {
    payload = event.data!.json() as PushPayload
  } catch (error) {
    console.error('[SW] Failed to parse push payload:', error)
    // Show a generic notification if parsing fails
    await self.registration.showNotification('Interact', {
      body: event.data?.text() || 'You have a new notification',
      icon: '/pwa-192x192.png',
      badge: '/pwa-64x64.png',
    })
    return
  }

  const notificationData: NotificationData = {
    url: payload.url || '/',
    notificationId: payload.notificationId,
    type: payload.type,
    ...payload.data,
  }

  const options: NotificationOptions = {
    body: payload.body,
    icon: payload.icon || '/pwa-192x192.png',
    badge: payload.badge || '/pwa-64x64.png',
    tag: payload.tag || payload.notificationId,
    data: notificationData,
    requireInteraction: shouldRequireInteraction(payload.type),
  }

  await self.registration.showNotification(payload.title, options)
}

/**
 * Determine if notification should stay until user interacts
 */
function shouldRequireInteraction(type: PushPayload['type']): boolean {
  // Important notifications should require interaction
  return ['achievement_unlocked', 'challenge_available'].includes(type)
}

/**
 * Handle notification click events
 */
export async function onNotificationClick(
  event: NotificationEvent,
): Promise<void> {
  event.notification.close()

  const promise = handleNotificationClick(event)
  event.waitUntil(promise)
}

async function handleNotificationClick(
  event: NotificationEvent,
): Promise<void> {
  const data = event.notification.data as NotificationData | undefined
  const urlToOpen = data?.url || '/'

  // Try to focus an existing window or open a new one
  const clients = await self.clients.matchAll({
    type: 'window',
    includeUncontrolled: true,
  })

  // Check if there's already an open window we can use
  for (const client of clients) {
    if ('focus' in client) {
      await client.focus()
      // Navigate to the notification URL
      if ('navigate' in client) {
        await (client as WindowClient).navigate(urlToOpen)
      }
      return
    }
  }

  // No existing window, open a new one
  await self.clients.openWindow(urlToOpen)
}
