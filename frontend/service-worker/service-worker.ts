/// <reference lib="WebWorker" />

declare let self: ServiceWorkerGlobalScope

// Inline push notification handlers to avoid module import issues in dev
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

  await self.registration.showNotification(payload.title, options)
}

async function onNotificationClick(event: NotificationEvent): Promise<void> {
  event.notification.close()

  const data = event.notification.data as NotificationData | undefined
  const urlToOpen = data?.url || '/'

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

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting()
  }
})

// Only use workbox caching in production
if (import.meta.env.PROD) {
  // Dynamic import to avoid issues in dev mode
  Promise.all([
    import('workbox-cacheable-response'),
    import('workbox-expiration'),
    import('workbox-precaching'),
    import('workbox-routing'),
    import('workbox-strategies'),
  ]).then(
    ([
      { CacheableResponsePlugin },
      { ExpirationPlugin },
      { cleanupOutdatedCaches, createHandlerBoundToURL, precacheAndRoute },
      { NavigationRoute, registerRoute },
      { NetworkFirst },
    ]) => {
      const entries = self.__WB_MANIFEST || []
      if (entries.length > 0) {
        precacheAndRoute(entries)
      }

      cleanupOutdatedCaches()

      const denylist = [
        /^\/sw.js$/,
        /^\/service-worker.js$/,
        /^\/manifest-(.*).webmanifest$/,
      ]

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

      registerRoute(
        new NavigationRoute(createHandlerBoundToURL('/'), { denylist }),
      )
    },
  )
}

self.addEventListener('push', onPush)
self.addEventListener('notificationclick', onNotificationClick)
