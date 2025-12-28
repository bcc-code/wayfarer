/// <reference lib="WebWorker" />

import { CacheableResponsePlugin } from 'workbox-cacheable-response'
import { ExpirationPlugin } from 'workbox-expiration'
import {
  cleanupOutdatedCaches,
  createHandlerBoundToURL,
  precacheAndRoute,
} from 'workbox-precaching'
import { NavigationRoute, registerRoute } from 'workbox-routing'
import { NetworkFirst } from 'workbox-strategies'

declare const self: ServiceWorkerGlobalScope & {
  __WB_MANIFEST: Array<{ url: string; revision: string | null }>
}
interface PushPayload {
  title: string
  body: string
  icon?: string
  badge?: string
  notificationId: string
  type: string
  url?: string
  data?: Record<string, unknown>
  tag?: string
}

interface NotificationData {
  url: string
  notificationId: string
  type: string
  [key: string]: unknown
}

// Track clicked notification IDs to distinguish clicks from dismissals
// (notificationclose fires for both, but we handle clicks separately)
const clickedNotifications = new Set<string>()

async function postAnalyticsEvent(
  event: string,
  properties: Record<string, unknown>,
): Promise<void> {
  const clients = await self.clients.matchAll({ type: 'window' })
  for (const client of clients) {
    client.postMessage({
      type: 'ANALYTICS_EVENT',
      event,
      properties,
    })
  }
}

async function onPush(event: PushEvent): Promise<void> {
  if (!event.data) {
    console.warn('[SW] Push event received without data')
    return
  }

  let payload: PushPayload

  try {
    payload = event.data.json() as PushPayload
  } catch (error) {
    console.error('[SW] Failed to parse push payload:', error)
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
    requireInteraction: [
      'achievement_unlocked',
      'challenge_available',
    ].includes(payload.type),
  }

  // Track notification received
  await postAnalyticsEvent('notification_received', {
    notification_id: payload.notificationId,
    notification_type: payload.type,
  })

  await self.registration.showNotification(payload.title, options)

  // Track notification displayed
  await postAnalyticsEvent('notification_displayed', {
    notification_id: payload.notificationId,
    notification_type: payload.type,
  })
}

async function onNotificationClick(event: NotificationEvent): Promise<void> {
  event.notification.close()

  const data = event.notification.data as NotificationData | undefined
  const urlToOpen = data?.url || '/'

  // Mark as clicked so we don't count it as dismissed
  if (data?.notificationId) {
    clickedNotifications.add(data.notificationId)
    // Clean up after 5 seconds
    setTimeout(() => clickedNotifications.delete(data.notificationId), 5000)
  }

  // Track notification clicked
  await postAnalyticsEvent('notification_clicked', {
    notification_id: data?.notificationId,
    notification_type: data?.type,
    url: data?.url,
  })

  const clients = await self.clients.matchAll({
    type: 'window',
    includeUncontrolled: true,
  })

  for (const client of clients) {
    if ('focus' in client) {
      await client.focus()
      if ('navigate' in client) {
        await (client as WindowClient).navigate(urlToOpen)
      }
      return
    }
  }

  await self.clients.openWindow(urlToOpen)
}

async function onNotificationClose(event: NotificationEvent): Promise<void> {
  const data = event.notification.data as NotificationData | undefined

  // Only track as dismissed if it wasn't clicked
  if (data?.notificationId && !clickedNotifications.has(data.notificationId)) {
    await postAnalyticsEvent('notification_dismissed', {
      notification_id: data.notificationId,
      notification_type: data.type,
    })
  }
}

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting()
  }
})

// Workbox caching setup
const entries = self.__WB_MANIFEST || []
if (entries.length > 0) {
  precacheAndRoute(entries)
}

cleanupOutdatedCaches()

const denylist = [/^\/service-worker.js$/, /^\/manifest-(.*).webmanifest$/]

registerRoute(
  ({ request, sameOrigin }) =>
    sameOrigin && request.destination === 'manifest',
  new NetworkFirst({
    cacheName: 'webmanifest',
    plugins: [
      new CacheableResponsePlugin({ statuses: [200] }),
      new ExpirationPlugin({ maxEntries: 100 }),
    ],
  }),
)

registerRoute(new NavigationRoute(createHandlerBoundToURL('/'), { denylist }))

self.addEventListener('push', onPush)
self.addEventListener('notificationclick', onNotificationClick)
self.addEventListener('notificationclose', onNotificationClose)
